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

type Manager struct {
	sessionMgr    *session.Manager
	similarityEng *similarity.Engine
	scheduler     *ExplorationScheduler
	config        config.PlaybackConfig
	db            *sqlite.Client

	queues map[string]*Queue
	mu     sync.RWMutex
	ttl    time.Duration
}

type Queue struct {
	SessionID string         `json:"session_id"`
	Current   *types.Track   `json:"current"`
	Next      *types.Track   `json:"next"`
	NextMode  string         `json:"next_mode"`
	Upcoming  []*types.Track `json:"upcoming"`
	PlayedIDs []uint         `json:"played_ids"`
	LastAccess time.Time     `json:"-"`
}

func NewManager(sm *session.Manager, se *similarity.Engine, db *sqlite.Client) *Manager {
	cfg := config.DefaultPlaybackConfig()
	m := &Manager{
		sessionMgr:    sm,
		similarityEng: se,
		scheduler:     NewExplorationScheduler(cfg),
		config:        cfg,
		db:            db,
		queues:        make(map[string]*Queue),
		ttl:           time.Hour,
	}
	go m.cleanupLoop()
	return m
}

func (m *Manager) GetQueue(sessionID string) (*Queue, error) {
	m.mu.RLock()
	q, ok := m.queues[sessionID]
	m.mu.RUnlock()

	if !ok {
		m.mu.Lock()
		// Double check
		q, ok = m.queues[sessionID]
		if !ok {
			q = &Queue{
				SessionID:  sessionID,
				LastAccess: time.Now(),
			}
			m.queues[sessionID] = q
		}
		m.mu.Unlock()
	}
	q.LastAccess = time.Now()
	return q, nil
}

func (m *Manager) getQueue(sessionID string) *Queue {
	return m.queues[sessionID]
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		m.mu.Lock()
		for id, q := range m.queues {
			if time.Since(q.LastAccess) > m.ttl {
				delete(m.queues, id)
			}
		}
		m.mu.Unlock()
	}
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
