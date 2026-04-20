package queue

import (
	"testing"

	"github.com/david22573/GoRadio/app/config"
	"github.com/david22573/GoRadio/app/session"
)

func TestExploitationWeightedRandom(t *testing.T) {
	// Logic moved to internal test or requires mocks
}

func TestExplorationRateDecay(t *testing.T) {
	scheduler := NewExplorationScheduler(config.DefaultPlaybackConfig())
	
	// Happy user: 0 skips, high completion
	s := &session.SessionState{
		PlayHistory: []session.PlayEvent{
			{Completion: 1.0},
			{Completion: 0.9},
			{Completion: 1.0},
		},
	}
	
	rate := scheduler.CalculateRate(s)
	if rate > scheduler.BaseRate {
		t.Errorf("Expected rate to decrease or stay base, got %f", rate)
	}
	
	// Bored user: many skips
	s.SkipHistory = []session.SkipEvent{
		{TrackID: 1},
		{TrackID: 2},
		{TrackID: 3},
		{TrackID: 4},
	}
	
	boredRate := scheduler.CalculateRate(s)
	if boredRate <= rate {
		t.Errorf("Expected rate to increase after skips, got %f (prev %f)", boredRate, rate)
	}
}
