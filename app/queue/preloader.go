package queue

import (
	"context"
)

// Preload ensures the upcoming buffer is filled with 3 tracks
func (m *Manager) Preload(ctx context.Context, sessionID string) {
	q, err := m.GetQueue(sessionID)
	if err != nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Fill upcoming until we have the desired prefetch size
	for len(q.Upcoming) < m.config.QueuePrefetchSize {
		next, err := m.GenerateNext(ctx, sessionID)
		if err != nil {
			break
		}
		
		// Avoid duplicates within upcoming
		isDup := false
		for _, u := range q.Upcoming {
			if u.ID == next.ID {
				isDup = true
				break
			}
		}
		if !isDup {
			q.Upcoming = append(q.Upcoming, next)
		} else {
			// If we hit a duplicate, stop to avoid infinite loop
			break
		}
	}
}
