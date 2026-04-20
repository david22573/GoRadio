package session

import (
	"math"
	"github.com/david22573/GoRadio/app/config"
)

// UpdateVector moves the session vector based on play events
func (s *SessionState) UpdateVector(event PlayEvent, trackVector []float64, cfg config.PlaybackConfig) {
	if event.Completion < cfg.SkipThreshold {
		// Skip detected: repel from track vector
		s.CurrentVector = VectorMove(s.CurrentVector, trackVector, cfg.VectorRepulsionRate)
		// Increase exploration if skips are high
		s.ExplorationRate = math.Min(cfg.MaxExplorationRate, s.ExplorationRate+0.02)
	} else if event.Completion > cfg.ListenThreshold {
		// Full listen: attract toward track vector
		s.CurrentVector = VectorMove(s.CurrentVector, trackVector, cfg.VectorLearningRate)
		// Decay exploration if listens are high
		s.ExplorationRate = math.Max(cfg.MinExplorationRate, s.ExplorationRate-0.01)
	}
}

// VectorMove performs a weighted move from current toward target
func VectorMove(current, target []float64, alpha float64) []float64 {
	if len(current) != len(target) {
		return current
	}
	result := make([]float64, len(current))
	for i := range current {
		// weighted average: current + alpha * (target - current)
		result[i] = current[i] + alpha*(target[i]-current[i])
	}
	return Normalize(result)
}

// Normalize ensures the vector is on the unit hypersphere
func Normalize(v []float64) []float64 {
	var norm float64
	for _, x := range v {
		norm += x * x
	}
	norm = math.Sqrt(norm)
	if norm == 0 {
		return v
	}
	result := make([]float64, len(v))
	for i, x := range v {
		result[i] = x / norm
	}
	return result
}
