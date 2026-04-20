package session

import (
	"math"
)

type SessionMetrics struct {
	SkipRate          float64 // Skips / Total plays
	AverageCompletion float64
	VectorDrift       float64 // Distance from start
}

// CalculateMetrics performs real-time calculation of listening patterns
func (s *SessionState) CalculateMetrics() SessionMetrics {
	totalPlays := float64(len(s.PlayHistory) + len(s.SkipHistory))
	if totalPlays == 0 {
		return SessionMetrics{}
	}

	var totalCompletion float64
	for _, p := range s.PlayHistory {
		totalCompletion += p.Completion
	}

	return SessionMetrics{
		SkipRate:          float64(len(s.SkipHistory)) / totalPlays,
		AverageCompletion: totalCompletion / math.Max(1.0, float64(len(s.PlayHistory))),
		VectorDrift:       CalculateDrift(s.CurrentVector, s.OriginVector),
	}
}

// CalculateDrift measures distance from current vector to a reference point
func CalculateDrift(current, origin []float64) float64 {
	if len(current) != len(origin) {
		return 0
	}
	var sum float64
	for i := range current {
		diff := current[i] - origin[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}
