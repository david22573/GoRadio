package session

import (
	"testing"
	"github.com/david22573/GoRadio/app/config"
)

func TestVectorMovement(t *testing.T) {
	// 3D vectors for simplicity
	origin := []float64{1.0, 0.0, 0.0}
	target := []float64{0.0, 1.0, 0.0}
	
	s := &SessionState{
		CurrentVector: origin,
		OriginVector:  origin,
		ExplorationRate: 0.1,
	}

	cfg := config.DefaultPlaybackConfig()

	// 1. Test Attraction (Listen)
	// Move 20% toward target
	s.UpdateVector(PlayEvent{Completion: 0.9}, target, cfg)
	
	// Check if Y component increased (moved toward target)
	if s.CurrentVector[1] <= 0 {
		t.Errorf("Expected Y component to increase after listen, got %f", s.CurrentVector[1])
	}

	// 2. Test Repulsion (Skip)
	beforeRepulsion := make([]float64, len(s.CurrentVector))
	copy(beforeRepulsion, s.CurrentVector)
	
	// Skip the same target vector
	s.UpdateVector(PlayEvent{Completion: 0.1}, target, cfg)
	
	// Check if Y component decreased (moved away from target)
	if s.CurrentVector[1] >= beforeRepulsion[1] {
		t.Errorf("Expected Y component to decrease after skip, got %f", s.CurrentVector[1])
	}
}

func TestExplorationRateScaling(t *testing.T) {
	cfg := config.DefaultPlaybackConfig()
	s := &SessionState{
		CurrentVector: []float64{1.0, 0.0},
		ExplorationRate: 0.15,
	}

	// Skip should increase exploration
	s.UpdateVector(PlayEvent{Completion: 0.05}, []float64{0.0, 1.0}, cfg)
	if s.ExplorationRate <= 0.15 {
		t.Errorf("Expected exploration rate to increase after skip, got %f", s.ExplorationRate)
	}

	// Full listen should decrease exploration
	currentRate := s.ExplorationRate
	s.UpdateVector(PlayEvent{Completion: 1.0}, []float64{1.0, 0.0}, cfg)
	if s.ExplorationRate >= currentRate {
		t.Errorf("Expected exploration rate to decrease after full listen, got %f", s.ExplorationRate)
	}
}
