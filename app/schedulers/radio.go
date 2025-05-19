package scheduler

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

func NewRadioScheduler(app *app.App, name, url string, loc *time.Location) (*RadioScheduler, error) {
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
		Name:  name,
		URL:   url,
		Shows: []types.Show{},

		app: app,

		Scheduler: sched,
	}, nil
}

func (s *RadioScheduler) ScheduleShow(show types.Show) {
	s.Shows = append(s.Shows, show)
}

func (s *RadioScheduler) Start() {
	for _, show := range s.Shows {
		s.ScheduleJob(show, func() {
			s.app.Radio.Record(&show)
		})
	}
}
v