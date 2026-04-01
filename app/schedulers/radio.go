package schedulers

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/david22573/GoRadio/app"
	"github.com/david22573/GoRadio/app/clients"
	"github.com/david22573/GoRadio/app/types"
	"github.com/go-co-op/gocron/v2"
)

// RadioScheduler embeds gocron.Scheduler for scheduler functionality
// while adding radio-specific state
type RadioScheduler struct {
	StationID uint
	Name      string
	URL       string
	Shows     []types.Show

	app  *app.App
	jobs map[uint]gocron.Job
	mu   sync.Mutex
	// The embedded scheduler interface
	gocron.Scheduler
}

// NewRadioScheduler creates a new RadioScheduler
func NewRadioScheduler(application *app.App, station types.Station, loc *time.Location) (*RadioScheduler, error) {
	shows, err := application.DB.GetAllShows()
	if err != nil {
		return nil, fmt.Errorf("failed to get shows: %w", err)
	}

	if loc == nil {
		var err error
		loc, err = time.LoadLocation("America/Los_Angeles")
		if err != nil {
			return nil, fmt.Errorf("failed to load location: %w", err)
		}
	}

	// Create the base scheduler
	scheduler, err := gocron.NewScheduler(gocron.WithLocation(loc))
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	jobs := make(map[uint]gocron.Job)

	// Create the radio scheduler with embedding
	rs := &RadioScheduler{
		StationID: station.ID,
		Name:      station.Name,
		URL:       station.URL,
		Shows:     shows,

		app:       application,
		jobs:      jobs,
		Scheduler: scheduler, // Assign the created scheduler to the embedded field
	}

	// Schedule shows
	if err := rs.scheduleShows(); err != nil {
		// Clean up on failure
		if shutdownErr := scheduler.Shutdown(); shutdownErr != nil {
			fmt.Printf("Failed to shutdown scheduler: %v\n", shutdownErr)
		}
		return nil, fmt.Errorf("failed to schedule shows: %w", err)
	}

	return rs, nil
}

func (r *RadioScheduler) Start() {
	fmt.Printf("Starting radio scheduler for %s...\n", r.Name)
	r.Scheduler.Start()
}

// scheduleShows adds all shows to the scheduler
func (s *RadioScheduler) scheduleShows() error {
	for i, show := range s.Shows {
		// Only schedule shows belonging to this specific station
		if show.StationID == s.StationID {
			show := show // shadow to close properly
			if err := s.ScheduleShow(show); err != nil {
				return fmt.Errorf("failed to schedule show %d (%s): %w", i, show.Name, err)
			}
		}
	}
	return nil
}

// ScheduleShow adds a single show to the scheduler
func (s *RadioScheduler) ScheduleShow(show types.Show) error {
	jobDef := gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(show.StartTime())))

	task := gocron.NewTask(func() {
		// Fire and forget in a goroutine so cmd.Wait() doesn't block gocron
		go func(targetShow types.Show) {
			client := clients.NewRadioClient("data/recordings", s.URL)
			// Pass the global application context here
			if err := client.Record(s.app.Ctx, &targetShow); err != nil {
				log.Printf("Recording failed for show %s: %v", targetShow.Name, err)
			}
		}(show)
	})

	job, err := s.NewJob(jobDef, task)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.jobs[show.ID] = job
	s.mu.Unlock()
	return nil
}

// CancelShow cancels a previously scheduled show by ID.
func (s *RadioScheduler) CancelShow(showID uint) error {
	s.mu.Lock()
	job, ok := s.jobs[showID]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("no job found for show %d", showID)
	}
	s.mu.Lock()
	delete(s.jobs, showID)
	s.mu.Unlock()
	return s.RemoveJob(job.ID())
}
