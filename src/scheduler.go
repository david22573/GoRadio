package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// RadioScheduler holds the Scheduler interface and the client.
type RadioScheduler struct {
	scheduler gocron.Scheduler // interface returned by NewScheduler() :contentReference[oaicite:0]{index=0}
	rc        *RadioClient
}

// NewRadioScheduler creates a new Scheduler configured with the local timezone.
func NewRadioScheduler(rootFolder, radioURL string) *RadioScheduler {
	s, err := gocron.NewScheduler(
		gocron.WithLocation(time.Local),
	)
	if err != nil {
		log.Fatalf("failed to create scheduler: %v", err)
	}

	return &RadioScheduler{
		scheduler: s,
		rc:        NewRadioClient(rootFolder, radioURL),
	}
}

// AddShows schedules each Show to record weekly using NewJob.
func (rs *RadioScheduler) AddShows(shows ...Show) {
	for _, show := range shows {
		hhmm := fmt.Sprintf("%02d:%02d", show.Hour, show.Min)

		_, err := rs.scheduler.NewJob(
			gocron.WeeklyJob(
				1,                            // every 1 week
				gocron.NewWeekdays(show.Day), // on the specified weekday
				gocron.NewAtTimes(
					gocron.NewAtTime(show.Hour, show.Min, 0),
				),
			),
			gocron.NewTask(
				rs.rc.Record, show,
			),
		)
		if err != nil {
			log.Printf("❌ error scheduling %s: %v", show.Name, err)
		} else {
			log.Printf("✅ scheduled %s on %s at %s for %v",
				show.Name, show.Day, hhmm, show.Duration)
		}
	}
}

// Start begins executing scheduled jobs.
func (rs *RadioScheduler) Start() {
	rs.scheduler.Start()
}

// Shutdown cleanly stops the scheduler and waits for jobs to finish.
func (rs *RadioScheduler) Shutdown() {
	if err := rs.scheduler.Shutdown(); err != nil {
		log.Printf("error shutting down scheduler: %v", err)
	}
}
