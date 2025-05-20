package types

import "time"

type Station struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ShowSchedule struct {
	Day  time.Weekday `json:"day"`
	Hour uint         `json:"hour"`
	Min  uint         `json:"min"`
}

type Show struct {
	ID       uint          `json:"id"`
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	ShowSchedule
	StationID uint `json:"station_id"`
}

func (s Show) StartTime() (uint, uint, uint) {
	return s.Hour, s.Min, 0
}
