package session

import (
	"time"
)

type SessionState struct {
	ID             string      `json:"id"`
	UserID         uint        `json:"user_id,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	LastActivityAt time.Time   `json:"last_activity_at"`
	CurrentVector  []float64   `json:"current_vector"` // 128-dim session position
	OriginVector   []float64   `json:"origin_vector"`  // 128-dim seed position
	PlayHistory    []PlayEvent `json:"play_history"`
	SkipHistory    []SkipEvent `json:"skip_history"`
	ExplorationRate float64    `json:"exploration_rate"` // 0.1-0.2
}

type PlayEvent struct {
	TrackID     uint      `json:"track_id"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	Completion  float64   `json:"completion"` // 0.0-1.0
}

type SkipEvent struct {
	TrackID   uint      `json:"track_id"`
	SkippedAt time.Time `json:"skipped_at"`
	PlayedFor int       `json:"played_for"` // seconds
}
