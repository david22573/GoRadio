package types

import (
	"time"
)

type Station struct {
	ID   uint   `json:"id" form:"id"`
	Name string `json:"name" form:"name"`
	URL  string `json:"url" form:"url"`
}

type Track struct {
	ID         uint      `json:"id"`
	StationID  uint      `json:"station_id"`
	Title      string    `json:"title"`
	Artist     string    `json:"artist"`
	Duration   int       `json:"duration"` // seconds
	AnalyzedAt time.Time `json:"analyzed_at"`
}
