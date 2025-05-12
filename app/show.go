package app

import "time"

// Show ties a name and schedule to a duration.
type Show struct {
	Name     string
	Duration time.Duration
	ShowSchedule
}

type ShowSchedule struct {
	Day       time.Weekday
	Hour, Min uint
}

// NewShow constructs a Show.
func NewShow(name string, sch ShowSchedule, duration time.Duration) Show {
	return Show{
		Name:         name,
		Duration:     duration,
		ShowSchedule: sch,
	}
}
