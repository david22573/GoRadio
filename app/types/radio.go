package types

import "time"

type Station struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ShowSchedule struct {
	Day  time.Weekday `json:"day"`
	Hour uint         `json:"hour"`
	Min  uint         `json:"min"`
}

type Show struct {
	ID       int           `json:"id"`
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	ShowSchedule
	StationID int `json:"station_id"`
}
