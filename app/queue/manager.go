package queue

import (
	"context"
	"sync"
	"time"

	"github.com/david22573/GoRadio/app/config"
	"github.com/david22573/GoRadio/app/db/sqlite"
	"github.com/david22573/GoRadio/app/services/similarity"
	"github.com/david22573/GoRadio/app/session"
	"github.com/david22573/GoRadio/app/types"
)

const (
	// defaultQueueCleanupInterval is how often idle in-memory queues are evicted.
	defaultQueueCleanupInterval = 10 * time.Minute
)

type Manager struct {
	sessionMgr    *session.Manager
	similarityEng *similarity.Engine
	scheduler     *ExplorationScheduler
	config        config.PlaybackConfig
	db            *sqlite.Client

	queues map[string]*Queue
	mu     sync.RWMutex
	ttl    time.Duration

	cleanupInterval time.Duration
}

type Queue struct {
	SessionID  string         `json:"session_id"`
	Current    *types.Track   `json:"current"`
	Next       *types.Track   `json:"next"`
	NextMode   string         `json:"next_mode"`
	Upcoming   []*types.Track `json:"upcoming"`
	PlayedIDs  []uint         `json:"played_ids"`
	LastAccess time.Time      `json:"-"`
}

// NewManager starts the cleanup loop. ctx controls the loop's lifetime: it stops
// when ctx is cancelled (application shutdown).
func NewManager(ctx context.Context, sm *session.Manager, se *similarity.Engine, db *sqlite.Client) *Manager {
	cfg := config.DefaultPlaybackConfig()
	m := &Manager{
		sessionMgr:      sm,
		similarityEng:   se,
		scheduler:       NewExplorationScheduler(cfg),
		config:          cfg,
		db:              db,
		queues:          make(map[string]*Queue),
		ttl:             time.Hour,
		cleanupInterval: defaultQueueCleanupInterval,
	}
	go m.cleanupLoop(ctx)
	return m
}

func (m *Manager) GetQueue(sessionID string) (*Queue, error) {
	// Look up and refresh the access clock in one critical section: the cleanup
	// loop reads LastAccess under the same mutex, so doing both together
	// prevents it from evicting the queue in the gap and handing the caller a
	// queue that is no longer tracked.
	m.mu.Lock()
	defer m.mu.Unlock()

	q, ok := m.queues[sessionID]
	if !ok {
		q = &Queue{SessionID: sessionID}
		m.queues[sessionID] = q
	}
	q.LastAccess = time.Now()

	return q, nil
}

func (m *Manager) cleanupLoop(ctx context.Context) {
	interval := m.cleanupInterval
	if interval <= 0 {
		interval = defaultQueueCleanupInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.evictIdleQueues(time.Now())
		}
	}
}

// evictIdleQueues drops in-memory queues untouched for longer than the TTL.
func (m *Manager) evictIdleQueues(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, q := range m.queues {
		if now.Sub(q.LastAccess) > m.ttl {
			delete(m.queues, id)
		}
	}
}

// size returns the number of in-memory queues.
func (m *Manager) size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.queues)
}

// Snapshot returns a deep copy of a queue that is safe to marshal to JSON
// without holding m.mu. Without this, API handlers would serialize the live
// Queue while Advance/Preload mutate it under the lock, causing data races.
func (m *Manager) Snapshot(sessionID string) *Queue {
	m.mu.RLock()
	defer m.mu.RUnlock()
	q, ok := m.queues[sessionID]
	if !ok {
		return &Queue{SessionID: sessionID}
	}
	return cloneQueue(q)
}

func cloneQueue(q *Queue) *Queue {
	if q == nil {
		return nil
	}
	c := *q
	c.Current = cloneTrack(q.Current)
	c.Next = cloneTrack(q.Next)
	c.Upcoming = make([]*types.Track, len(q.Upcoming))
	for i, t := range q.Upcoming {
		c.Upcoming[i] = cloneTrack(t)
	}
	c.PlayedIDs = append([]uint(nil), q.PlayedIDs...)
	return &c
}

func cloneTrack(t *types.Track) *types.Track {
	if t == nil {
		return nil
	}
	cp := *t
	return &cp
}

func (m *Manager) Advance(ctx context.Context, sessionID string) (*types.Track, string, error) {
	q, err := m.GetQueue(sessionID)
	if err != nil {
		return nil, "", err
	}

	m.mu.Lock()
	if q.Current != nil {
		q.PlayedIDs = append(q.PlayedIDs, q.Current.ID)
		if len(q.PlayedIDs) > m.config.HistorySize {
			q.PlayedIDs = q.PlayedIDs[len(q.PlayedIDs)-m.config.HistorySize:]
		}
	}

	mode := q.NextMode
	q.Current = q.Next

	needsGenerate := false
	if len(q.Upcoming) > 0 {
		q.Next = q.Upcoming[0]
		q.NextMode = "exploitation"
		q.Upcoming = q.Upcoming[1:]
	} else {
		needsGenerate = true
	}

	needsDoubleGenerate := q.Current == nil
	m.mu.Unlock()

	if needsGenerate {
		next, nextMode, err := m.generateNextWithQueue(ctx, sessionID, q)
		if err != nil {
			return nil, "", err
		}
		m.mu.Lock()
		q.Next = next
		q.NextMode = nextMode
		m.mu.Unlock()
	}

	if needsDoubleGenerate {
		m.mu.Lock()
		q.Current = q.Next
		mode = q.NextMode
		m.mu.Unlock()

		next, nextMode, err := m.generateNextWithQueue(ctx, sessionID, q)
		if err == nil {
			m.mu.Lock()
			q.Next = next
			q.NextMode = nextMode
			m.mu.Unlock()
		}
	}

	go m.Preload(context.Background(), sessionID)

	m.mu.RLock()
	retCurrent := q.Current
	m.mu.RUnlock()

	return retCurrent, mode, nil
}