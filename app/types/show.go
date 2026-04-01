package types

import (
	"fmt"
	"time"
)

type Show struct {
	ID        uint         `json:"id" form:"id"`
	Name      string       `json:"name" form:"name"`
	Day       uint         `json:"day" form:"day"`   // 0 = Sunday
	Hour      uint         `json:"hour" form:"hour"` // 0–23
	Min       uint         `json:"min" form:"min"`   // 0–59
	Duration  ShowDuration `json:"duration" form:"duration"`
	Scheduled bool         `json:"scheduled" form:"scheduled"`
	StationID uint         `json:"station_id" form:"station_id"`
}

type ShowSchedule struct {
	Day  time.Weekday
	Hour uint
	Min  uint
}

func (s *Show) GetSchedule() ShowSchedule {
	return ShowSchedule{Day: time.Weekday(s.Day), Hour: s.Hour, Min: s.Min}
}

func (s *Show) UpdateSchedule(schedule ShowSchedule) {
	s.Day = uint(schedule.Day)
	s.Hour = schedule.Hour
	s.Min = schedule.Min
}

func (s Show) StartTime() (uint, uint, uint) {
	return s.Hour, s.Min, 0
}

type ShowDuration struct{ time.Duration }

func (s ShowDuration) String() string {
	return time.Duration(s.Duration).String()
}

func (s ShowDuration) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", int64(s.Duration/time.Second))), nil
}

func (s *ShowDuration) UnmarshalJSON(b []byte) error {
	var seconds int64
	_, err := fmt.Sscanf(string(b), "%d", &seconds)
	if err != nil {
		return err
	}
	*s = ShowDuration{Duration: time.Duration(seconds) * time.Second}
	return nil
}
