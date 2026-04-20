package queue

import (
	"context"
	"math/rand"

	"github.com/david22573/GoRadio/app/session"
	"github.com/david22573/GoRadio/app/types"
)

func (m *Manager) selectExplorationTrack(ctx context.Context, s *session.SessionState, excludeIDs []uint) (*types.Track, error) {
	// Decide between controlled and wild exploration based on config
	if rand.Float64() < m.config.ControlledRatio {
		// Controlled exploration: fetch tracks in the configured distance range
		ids, err := m.db.SearchRange(s.CurrentVector, m.config.ExplorationMinDistance, m.config.ExplorationMaxDistance, 10)
		if err == nil && len(ids) > 0 {
			// Filter exclusions
			excludeMap := make(map[uint]bool)
			for _, id := range excludeIDs {
				excludeMap[id] = true
			}

			var candidates []uint
			for _, id := range ids {
				if !excludeMap[id] {
					candidates = append(candidates, id)
				}
			}

			if len(candidates) > 0 {
				selectedID := candidates[rand.Intn(len(candidates))]
				return m.db.GetTrackByID(selectedID)
			}
		}
	}

	// Wild exploration: completely random
	return m.db.GetRandomTrack(excludeIDs)
}
