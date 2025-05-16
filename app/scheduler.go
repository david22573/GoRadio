package app

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/david22573/GoRadio/app/types"
	"github.com/go-co-op/gocron/v2"
)

const (
	defaultTimeZone = "America/Los_Angeles"
)

type Scheduler interface {
	Schedule()
	Start()
	Shutdown()
}

// RadioScheduler holds the Scheduler interface and the client.
type RadioScheduler struct {
	radioClient *RadioClient
	scheduler   gocron.Scheduler
	schedules   []types.Show
}

// NewRadioScheduler creates a new Scheduler configured with the local timezone.
func NewRadioScheduler(station *types.Station) *RadioScheduler {
	tz, _ := time.LoadLocation(defaultTimeZone)
	s, err := gocron.NewScheduler(
		gocron.WithLocation(tz),
	)
	if err != nil {
		log.Fatalf("failed to create scheduler: %v", err)
	}
	cleanStationName := filepath.Clean(station.Name)
	rootFolder := filepath.Join("recordings", cleanStationName)
	return &RadioScheduler{
		scheduler:   s,
		radioClient: NewRadioClient(rootFolder, station.URL),
	}
}

// AddShows schedules each Show to record weekly using NewJob.
func (rs *RadioScheduler) Schedule(shows ...types.Show) {
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
				rs.radioClient.Record, show,
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
