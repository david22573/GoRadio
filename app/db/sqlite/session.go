package sqlite

import (
	"database/sql"
	"time"

	"github.com/david22573/GoRadio/app/types"
)

func (c *Client) CreateSession(s *types.SessionState) error {
	cVect := SerializeFloat32(float64ToFloat32(s.CurrentVector))
	oVect := SerializeFloat32(float64ToFloat32(s.OriginVector))

	query := `INSERT INTO sessions (id, user_id, created_at, last_activity_at, current_vector, origin_vector, exploration_rate)
              VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := c.db.Exec(query, s.ID, s.UserID, s.CreatedAt, s.LastActivityAt, cVect, oVect, s.ExplorationRate)
	return err
}

func (c *Client) GetSession(id string) (*types.SessionState, error) {
	var s types.SessionState
	var cVect, oVect []byte
	query := `SELECT id, user_id, created_at, last_activity_at, current_vector, origin_vector, exploration_rate 
              FROM sessions WHERE id = ?`
	
	err := c.db.QueryRow(query, id).Scan(&s.ID, &s.UserID, &s.CreatedAt, &s.LastActivityAt, &cVect, &oVect, &s.ExplorationRate)
	if err != nil {
		return nil, err
	}

	s.CurrentVector = float32ToFloat64(DeserializeFloat32(cVect))
	s.OriginVector = float32ToFloat64(DeserializeFloat32(oVect))

	// Load history
	plays, skips, err := c.GetSessionEvents(id)
	if err == nil {
		s.PlayHistory = plays
		s.SkipHistory = skips
	}

	return &s, nil
}

func (c *Client) UpdateSession(s *types.SessionState) error {
	cVect := SerializeFloat32(float64ToFloat32(s.CurrentVector))

	query := `UPDATE sessions SET last_activity_at = ?, current_vector = ?, exploration_rate = ?
              WHERE id = ?`
	_, err := c.db.Exec(query, s.LastActivityAt, cVect, s.ExplorationRate, s.ID)
	return err
}

func (c *Client) DeleteSession(id string) error {
	_, err := c.db.Exec("DELETE FROM sessions WHERE id = ?", id)
	return err
}

func (c *Client) RecordPlayEvent(sessionID string, event types.PlayEvent) error {
	query := `INSERT INTO session_events (session_id, track_id, event_type, completion, created_at)
              VALUES (?, ?, 'play', ?, ?)`
	_, err := c.db.Exec(query, sessionID, event.TrackID, event.Completion, event.CompletedAt)
	return err
}

func (c *Client) RecordSkipEvent(sessionID string, event types.SkipEvent) error {
	query := `INSERT INTO session_events (session_id, track_id, event_type, played_for, created_at)
              VALUES (?, ?, 'skip', ?, ?)`
	_, err := c.db.Exec(query, sessionID, event.TrackID, event.PlayedFor, event.SkippedAt)
	return err
}

func (c *Client) GetSessionEvents(sessionID string) ([]types.PlayEvent, []types.SkipEvent, error) {
	query := `SELECT track_id, event_type, completion, played_for, created_at 
              FROM session_events WHERE session_id = ? ORDER BY created_at ASC`
	
	rows, err := c.db.Query(query, sessionID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var plays []types.PlayEvent
	var skips []types.SkipEvent

	for rows.Next() {
		var trackID uint
		var eventType string
		var completion sql.NullFloat64
		var playedFor sql.NullInt64
		var createdAt time.Time

		if err := rows.Scan(&trackID, &eventType, &completion, &playedFor, &createdAt); err != nil {
			return nil, nil, err
		}

		if eventType == "play" {
			plays = append(plays, types.PlayEvent{
				TrackID:     trackID,
				CompletedAt: createdAt,
				Completion:  completion.Float64,
				// StartedAt is not stored, can be estimated if needed
				StartedAt: createdAt.Add(-time.Duration(completion.Float64*300) * time.Second), 
			})
		} else if eventType == "skip" {
			skips = append(skips, types.SkipEvent{
				TrackID:   trackID,
				SkippedAt: createdAt,
				PlayedFor: int(playedFor.Int64),
			})
		}
	}

	return plays, skips, nil
}

func float64ToFloat32(v []float64) []float32 {
	res := make([]float32, len(v))
	for i, x := range v {
		res[i] = float32(x)
	}
	return res
}

func float32ToFloat64(v []float32) []float64 {
	res := make([]float64, len(v))
	for i, x := range v {
		res[i] = float64(x)
	}
	return res
}
