package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/david22573/GoRadio/app/config"
	"github.com/david22573/GoRadio/app/db/sqlite"
)

const (
	// defaultSessionCleanupInterval is how often idle in-memory sessions are evicted.
	defaultSessionCleanupInterval = 10 * time.Minute
	// defaultSessionDBCleanupInterval is how often old sessions are pruned from the DB.
	defaultSessionDBCleanupInterval = 24 * time.Hour
	// sessionRetention is how long sessions are kept in the database.
	sessionRetention = 30 * 24 * time.Hour
)

type Manager struct {
	db       *sqlite.Client
	sessions map[string]*SessionState
	mu       sync.RWMutex
	ttl      time.Duration

	cleanupInterval   time.Duration
	dbCleanupInterval time.Duration

	// persistMu serialises asynchronous session writes; persistedSeq holds, per
	// session, the newest snapshot sequence that has been claimed for writing so
	// a late goroutine can never roll the stored session back to an older state.
	persistMu    sync.Mutex
	persistedSeq map[string]uint64
	// persistSeq numbers the snapshots taken under mu.
	persistSeq uint64

	// vectorFor and persistSession dispatch to the SQLite client; tests replace
	// them to observe persistence ordering without a live database.
	vectorFor      func(trackID uint) ([]float64, error)
	persistSession func(*SessionState) error
}

// NewManager starts the cleanup loop. ctx controls the loop's lifetime: it stops
// when ctx is cancelled (application shutdown).
func NewManager(ctx context.Context, db *sqlite.Client) *Manager {
	m := &Manager{
		db:                db,
		sessions:          make(map[string]*SessionState),
		ttl:               time.Hour,
		cleanupInterval:   defaultSessionCleanupInterval,
		dbCleanupInterval: defaultSessionDBCleanupInterval,
	}
	go m.cleanupLoop(ctx)
	return m
}

func (m *Manager) CreateSession(ctx context.Context, seedTrackID uint) (*SessionState, error) {
	embedding, err := m.db.GetVectorByID(seedTrackID)
	if err != nil {
		return nil, fmt.Errorf("seed track embedding not found: %w", err)
	}

	session := &SessionState{
		ID:              uuid.New().String(),
		CreatedAt:       time.Now(),
		LastActivityAt:  time.Now(),
		CurrentVector:   embedding,
		OriginVector:    embedding,
		ExplorationRate: 0.1,
	}

	if err := m.db.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to persist session: %w", err)
	}

	m.mu.Lock()
	m.sessions[session.ID] = session
	m.mu.Unlock()

	return session, nil
}

func (m *Manager) GetSession(id string) (*SessionState, error) {
	// Look up, refresh the activity clock and queue the persisted update in one
	// critical section: the cleanup loop reads LastActivityAt under the same
	// mutex, so doing both together prevents it from evicting the session in the
	// gap and handing the caller a session that is no longer tracked.
	if s, ok := m.touchSession(id); ok {
		return s, nil
	}

	// Try to load from DB
	loaded, err := m.db.GetSession(id)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Add to cache (another request may have loaded it meanwhile).
	m.mu.Lock()
	s, ok := m.sessions[id]
	if !ok {
		s = loaded
		m.sessions[id] = s
	}
	s.LastActivityAt = time.Now()
	m.persistAsync(snapshotSession(s), m.nextSeq())
	m.mu.Unlock()

	return s, nil
}

// touchSession refreshes the activity clock of a cached session under the map
// lock, in the same critical section as the lookup, so the cleanup loop cannot
// evict the session in between. It also queues an ordered asynchronous persist
// of the refreshed snapshot.
func (m *Manager) touchSession(id string) (*SessionState, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[id]
	if !ok {
		return nil, false
	}
	s.LastActivityAt = time.Now()
	m.persistAsync(snapshotSession(s), m.nextSeq())
	return s, true
}

// nextSeq allocates the next persistence sequence number. Callers must hold mu,
// which makes the numbering order equal to the order the snapshots are taken.
func (m *Manager) nextSeq() uint64 {
	m.persistSeq++
	return m.persistSeq
}

// persistAsync writes a snapshot on a best-effort background goroutine so
// callers never block on the database.
func (m *Manager) persistAsync(snap *SessionState, seq uint64) {
	if snap == nil {
		return
	}
	go m.persistSnapshot(snap, seq)
}

// persistSnapshot writes snap unless a newer snapshot of the same session has
// already been claimed. Snapshots carry the order in which they were taken, so a
// goroutine that runs late can never overwrite newer session state with an older
// one. The sequence is claimed before the write is attempted: a failed newer
// write must not open the door for an older state to land after it.
func (m *Manager) persistSnapshot(snap *SessionState, seq uint64) {
	if snap == nil {
		return
	}

	m.persistMu.Lock()
	defer m.persistMu.Unlock()

	if m.persistedSeq == nil {
		m.persistedSeq = make(map[string]uint64)
	}
	if m.persistedSeq[snap.ID] >= seq {
		return
	}
	m.persistedSeq[snap.ID] = seq

	m.writeSession(snap)
}

// writeSession dispatches to the injectable writer or the SQLite client.
func (m *Manager) writeSession(s *SessionState) error {
	if s == nil {
		return nil
	}
	if m.persistSession != nil {
		return m.persistSession(s)
	}
	return m.db.UpdateSession(s)
}

// lookupVector dispatches to the injectable vector reader or the SQLite client.
func (m *Manager) lookupVector(trackID uint) ([]float64, error) {
	if m.vectorFor != nil {
		return m.vectorFor(trackID)
	}
	return m.db.GetVectorByID(trackID)
}

// forgetPersistSeq drops the recorded sequence for an evicted session so the
// bookkeeping cannot outgrow the session cache. Callers must hold mu.
func (m *Manager) forgetPersistSeq(id string) {
	m.persistMu.Lock()
	delete(m.persistedSeq, id)
	m.persistMu.Unlock()
}

// snapshotSession copies the mutable state that an asynchronous DB write reads,
// so persistence never observes a session being mutated by a concurrent
// request. History slices are append-only under the lock and are not read by
// UpdateSession, so they are intentionally shared.
func snapshotSession(s *SessionState) *SessionState {
	if s == nil {
		return nil
	}
	cp := *s
	cp.CurrentVector = append([]float64(nil), s.CurrentVector...)
	cp.OriginVector = append([]float64(nil), s.OriginVector...)
	return &cp
}

func (m *Manager) Create(seedTrackID uint) (*SessionState, error) {
	return m.CreateSession(context.Background(), seedTrackID)
}

func (m *Manager) RecordPlay(sessionID string, trackID uint, completion float64) error {
	// Clamp completion to [0,1] so the started-at heuristic stays sane even for
	// malformed client input.
	if completion < 0 {
		completion = 0
	}
	if completion > 1 {
		completion = 1
	}

	event := PlayEvent{
		TrackID:     trackID,
		StartedAt:   time.Now().Add(-time.Duration(completion*300) * time.Second), // Heuristic
		CompletedAt: time.Now(),
		Completion:  completion,
	}

	if err := m.LogPlayEvent(sessionID, event); err != nil {
		return err
	}

	// Trigger vector evolution
	m.evolveVector(sessionID, trackID, event.Completion)
	return nil
}

// evolveVector applies a playback event to the cached session vector and
// persists the result. The whole read-modify-persist sequence runs under the map
// lock so concurrent events for the same session cannot race on the shared
// session state, and the DB write receives an immutable snapshot.
//
// The snapshot is persisted even when the track vector is unavailable: the
// activity timestamp written by LogPlayEvent/LogSkipEvent must reach storage
// regardless of whether vector evolution could happen.
func (m *Manager) evolveVector(sessionID string, trackID uint, completion float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[sessionID]
	if !ok {
		return
	}

	vec, err := m.lookupVector(trackID)
	if err == nil && vec != nil {
		cfg := config.DefaultPlaybackConfig()
		s.UpdateVector(completion, vec, cfg.SkipThreshold, cfg.ListenThreshold, cfg.MaxExplorationRate, cfg.MinExplorationRate, cfg.VectorLearningRate, cfg.VectorRepulsionRate)
		s.LastActivityAt = time.Now()
	}

	m.persistAsync(snapshotSession(s), m.nextSeq())
}

func (m *Manager) RecordSkip(sessionID string, trackID uint, playedFor int) error {
	event := SkipEvent{
		TrackID:   trackID,
		SkippedAt: time.Now(),
		PlayedFor: playedFor,
	}

	if err := m.LogSkipEvent(sessionID, event); err != nil {
		return err
	}

	// Trigger vector evolution (skip = low completion)
	m.evolveVector(sessionID, trackID, 0.1)

	return nil
}

func (m *Manager) GetMetrics(sessionID string) (SessionMetrics, error) {
	s, err := m.GetSession(sessionID)
	if err != nil {
		return SessionMetrics{}, err
	}
	return s.CalculateMetrics(), nil
}

func (m *Manager) GetJourney(sessionID string) ([]JourneyPoint, error) {
	s, err := m.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	return s.GetJourney(), nil
}

func (m *Manager) LogPlayEvent(sessionID string, event PlayEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found")
	}

	if err := m.db.RecordPlayEvent(sessionID, event); err != nil {
		return fmt.Errorf("failed to persist play event: %w", err)
	}

	s.PlayHistory = append(s.PlayHistory, event)
	s.LastActivityAt = time.Now()

	// Prune history
	if len(s.PlayHistory) > 100 {
		s.PlayHistory = s.PlayHistory[len(s.PlayHistory)-100:]
	}
	return nil
}

func (m *Manager) LogSkipEvent(sessionID string, event SkipEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found")
	}

	if err := m.db.RecordSkipEvent(sessionID, event); err != nil {
		return fmt.Errorf("failed to persist skip event: %w", err)
	}

	s.SkipHistory = append(s.SkipHistory, event)
	s.LastActivityAt = time.Now()

	// Prune history
	if len(s.SkipHistory) > 100 {
		s.SkipHistory = s.SkipHistory[len(s.SkipHistory)-100:]
	}
	return nil
}

func (m *Manager) cleanupLoop(ctx context.Context) {
	interval := m.cleanupInterval
	if interval <= 0 {
		interval = defaultSessionCleanupInterval
	}
	dbInterval := m.dbCleanupInterval
	if dbInterval <= 0 {
		dbInterval = defaultSessionDBCleanupInterval
	}

	ticker := time.NewTicker(interval)
	dbTicker := time.NewTicker(dbInterval)
	defer ticker.Stop()
	defer dbTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.evictIdleSessions(time.Now())
		case <-dbTicker.C:
			// Prune old sessions from DB (30 days)
			if m.db != nil {
				m.db.CleanupOldSessions(sessionRetention)
			}
		}
	}
}

// evictIdleSessions drops in-memory sessions untouched for longer than the TTL.
func (m *Manager) evictIdleSessions(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, s := range m.sessions {
		if now.Sub(s.LastActivityAt) > m.ttl {
			delete(m.sessions, id)
			m.forgetPersistSeq(id)
		}
	}
}

// size returns the number of in-memory sessions.
func (m *Manager) size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// Snapshot returns a deep copy of a session that is safe to marshal to JSON
// without holding m.mu. API handlers would otherwise serialize the live
// SessionState while evolveVector/RecordPlay/RecordSkip mutate it under the lock.
func (m *Manager) Snapshot(id string) (*SessionState, error) {
	s, err := m.GetSession(id)
	if err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneSession(s), nil
}

// cloneSession deep-copies all mutable slices of a session so the copy can be
// read outside the lock. Callers must hold m.mu (read or write).
func cloneSession(s *SessionState) *SessionState {
	if s == nil {
		return nil
	}
	cp := *s
	cp.CurrentVector = append([]float64(nil), s.CurrentVector...)
	cp.OriginVector = append([]float64(nil), s.OriginVector...)
	cp.PlayHistory = append([]PlayEvent(nil), s.PlayHistory...)
	cp.SkipHistory = append([]SkipEvent(nil), s.SkipHistory...)
	return &cp
}