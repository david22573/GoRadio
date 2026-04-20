package queue

import (
	"context"
	"sync"

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
}

type Queue struct {
	SessionID string         `json:"session_id"`
	Current   *types.Track   `json:"current"`
	Next      *types.Track   `json:"next"`
	Upcoming  []*types.Track `json:"upcoming"`
	PlayedIDs []uint         `json:"played_ids"`
}

func NewManager(sm *session.Manager, se *similarity.Engine, db *sqlite.Client) *Manager {
	cfg := config.DefaultPlaybackConfig()
	return &Manager{
		sessionMgr:    sm,
		similarityEng: se,
		scheduler:     NewExplorationScheduler(cfg),
		config:        cfg,
		db:            db,
		queues:        make(map[string]*Queue),
	}
}

func (m *Manager) GetQueue(sessionID string) (*Queue, error) {
	m.mu.RLock()
	q, ok := m.queues[sessionID]
	m.mu.RUnlock()

	if !ok {
		q = &Queue{SessionID: sessionID}
		m.mu.Lock()
		m.queues[sessionID] = q
		m.mu.Unlock()
	}
	return q, nil
}

func (m *Manager) Advance(ctx context.Context, sessionID string) (*types.Track, error) {
	q, err := m.GetQueue(sessionID)
	if err != nil {
		return nil, err
	}

	if q.Current != nil {
		q.PlayedIDs = append(q.PlayedIDs, q.Current.ID)
		if len(q.PlayedIDs) > m.config.HistorySize {
			q.PlayedIDs = q.PlayedIDs[len(q.PlayedIDs)-m.config.HistorySize:]
		}
	}

	q.Current = q.Next
	if len(q.Upcoming) > 0 {
		q.Next = q.Upcoming[0]
		q.Upcoming = q.Upcoming[1:]
	} else {
		next, err := m.GenerateNext(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		q.Next = next
	}

	// Async preload to fill upcoming
	go m.Preload(ctx, sessionID)

	return q.Current, nil
}
