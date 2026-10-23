package queue

import (
	"context"
)

// Preload ensures the upcoming buffer is filled with 3 tracks.
//
// It must never call generateNextWithQueue while holding m.mu: that helper takes
// m.mu.RLock() internally, and Go's sync.RWMutex is not reentrant, so holding the
// write lock across the call self-deadlocks.
func (m *Manager) Preload(ctx context.Context, sessionID string) {
	q, err := m.GetQueue(sessionID)
	if err != nil {
		return
	}

	for {
		// Generate the next track without holding m.mu.
		next, _, err := m.generateNextWithQueue(ctx, sessionID, q)
		if err != nil || next == nil {
			return
		}

		m.mu.Lock()
		if len(q.Upcoming) >= m.config.QueuePrefetchSize {
			m.mu.Unlock()
			return
		}

		// Avoid duplicates within upcoming (already handled by generateNextWithQueue, but double check)
		isDup := false
		for _, u := range q.Upcoming {
			if u.ID == next.ID {
				isDup = true
				break
			}
		}
		if isDup {
			m.mu.Unlock()
			return
		}
		q.Upcoming = append(q.Upcoming, next)
		m.mu.Unlock()
	}
}
