package queue

import (
	"context"
	"math/rand"

	"github.com/david22573/GoRadio/app/types"
)

func (m *Manager) GenerateNext(ctx context.Context, sessionID string) (*types.Track, string, error) {
	q, err := m.GetQueue(sessionID)
	if err != nil {
		return nil, "", err
	}
	return m.generateNextWithQueue(ctx, sessionID, q)
}

func (m *Manager) generateNextWithQueue(ctx context.Context, sessionID string, q *Queue) (*types.Track, string, error) {
	s, err := m.sessionMgr.GetSession(sessionID)
	if err != nil {
		return nil, "", err
	}

	// Combine PlayedIDs and Upcoming IDs to exclude
	excludeIDs := make([]uint, 0, len(q.PlayedIDs)+len(q.Upcoming)+2)
	excludeIDs = append(excludeIDs, q.PlayedIDs...)
	for _, t := range q.Upcoming {
		excludeIDs = append(excludeIDs, t.ID)
	}
	if q.Current != nil {
		excludeIDs = append(excludeIDs, q.Current.ID)
	}
	if q.Next != nil {
		excludeIDs = append(excludeIDs, q.Next.ID)
	}

	// Decide: exploitation or exploration?
	if rand.Float64() < m.scheduler.CalculateRate(s) {
		t, err := m.selectExplorationTrack(ctx, s, excludeIDs)
		return t, "exploration", err
	}
	t, err := m.selectExploitationTrack(ctx, s, excludeIDs)
	return t, "exploitation", err
}
