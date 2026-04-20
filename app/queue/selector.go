package queue

import (
	"context"
	"math/rand"

	"github.com/david22573/GoRadio/app/types"
)

func (m *Manager) GenerateNext(ctx context.Context, sessionID string) (*types.Track, string, error) {
	s, err := m.sessionMgr.GetSession(sessionID)
	if err != nil {
		return nil, "", err
	}

	q, _ := m.GetQueue(sessionID)
	excludeIDs := q.PlayedIDs

	// Decide: exploitation or exploration?
	if rand.Float64() < m.scheduler.CalculateRate(s) {
		t, err := m.selectExplorationTrack(ctx, s, excludeIDs)
		return t, "exploration", err
	}
	t, err := m.selectExploitationTrack(ctx, s, excludeIDs)
	return t, "exploitation", err
}
