package schedulers

import (
	"fmt"
	"time"

	"github.com/david22573/GoRadio/app"
	"github.com/david22573/GoRadio/app/types"
	"github.com/go-co-op/gocron/v2"
)

type RadioScheduler struct {
	Name  string
	URL   string
	Shows []types.Show

	app *app.App

	gocron.Scheduler
}

func NewRadioScheduler(app *app.App, station types.Station, loc *time.Location) (gocron.Scheduler, error) {
	shows, err := app.Repo.GetAllShowsByStation(station.ID)
	if err != nil {
		return nil, err
	}
	if loc == nil {
		var err error
		loc, err = time.LoadLocation("America/Los_Angeles")
		if err != nil {
			return nil, fmt.Errorf("failed to load location: %w", err)
		}
	}
	sched, err := gocron.NewScheduler(gocron.WithLocation(loc))
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}
	return &RadioScheduler{
		Name:  station.Name,
		URL:   station.URL,
		Shows: shows,

		app: app,

		Scheduler: sched,
	}, nil
}

func (s *RadioScheduler) scheduleShows() error {
	for _, show := range s.Shows {
		if err := s.ScheduleShow(show); err != nil {
			return err
		}
	}
	return nil
}

func (s *RadioScheduler) ScheduleShow(show types.Show) error {
	return nil
}
