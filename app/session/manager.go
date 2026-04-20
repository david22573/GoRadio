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

type Manager struct {
	db       *sqlite.Client
	sessions map[string]*SessionState
	mu       sync.RWMutex
	ttl      time.Duration
}

func NewManager(db *sqlite.Client) *Manager {
	m := &Manager{
		db:       db,
		sessions: make(map[string]*SessionState),
		ttl:      time.Hour,
	}
	go m.cleanupLoop()
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
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()

	if !ok {
		// Try to load from DB
		var err error
		s, err = m.db.GetSession(id)
		if err != nil {
			return nil, fmt.Errorf("session not found: %w", err)
		}
		
		// Add to cache
		m.mu.Lock()
		m.sessions[id] = s
		m.mu.Unlock()
	}

	s.LastActivityAt = time.Now()
	// Async update DB
	go m.db.UpdateSession(s)
	
	return s, nil
}

func (m *Manager) Create(seedTrackID uint) (*SessionState, error) {
	return m.CreateSession(context.Background(), seedTrackID)
}

func (m *Manager) RecordPlay(sessionID string, trackID uint, completion float64) error {
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
	s, _ := m.GetSession(sessionID)
	if s != nil {
		vec, _ := m.db.GetVectorByID(trackID)
		if vec != nil {
			cfg := config.DefaultPlaybackConfig()
			s.UpdateVector(event.Completion, vec, cfg.SkipThreshold, cfg.ListenThreshold, cfg.MaxExplorationRate, cfg.MinExplorationRate, cfg.VectorLearningRate, cfg.VectorRepulsionRate)
			// Persist updated vector
			go m.db.UpdateSession(s)
		}
	}
	return nil
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
	s, _ := m.GetSession(sessionID)
	if s != nil {
		vec, _ := m.db.GetVectorByID(trackID)
		if vec != nil {
			cfg := config.DefaultPlaybackConfig()
			s.UpdateVector(0.1, vec, cfg.SkipThreshold, cfg.ListenThreshold, cfg.MaxExplorationRate, cfg.MinExplorationRate, cfg.VectorLearningRate, cfg.VectorRepulsionRate)
			// Persist updated vector
			go m.db.UpdateSession(s)
		}
	}

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

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		m.mu.Lock()
		for id, s := range m.sessions {
			if time.Since(s.LastActivityAt) > m.ttl {
				delete(m.sessions, id)
			}
		}
		m.mu.Unlock()
	}
}
