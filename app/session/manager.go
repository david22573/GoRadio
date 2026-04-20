package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
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

	m.mu.Lock()
	m.sessions[session.ID] = session
	m.mu.Unlock()

	return session, nil
}

func (m *Manager) GetSession(id string) (*SessionState, error) {
	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	session.LastActivityAt = time.Now()
	return session, nil
}

func (m *Manager) LogPlayEvent(sessionID string, event PlayEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found")
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
