package types

import (
	"math"
	"time"
)

type SessionState struct {
	ID              string      `json:"id"`
	UserID          uint        `json:"user_id,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	LastActivityAt  time.Time   `json:"last_activity_at"`
	CurrentVector   []float64   `json:"current_vector"` // 128-dim session position
	OriginVector    []float64   `json:"origin_vector"`  // 128-dim seed position
	PlayHistory     []PlayEvent `json:"play_history"`
	SkipHistory     []SkipEvent `json:"skip_history"`
	ExplorationRate float64     `json:"exploration_rate"` // 0.1-0.2
}

type PlayEvent struct {
	TrackID     uint      `json:"track_id"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	Completion  float64   `json:"completion"` // 0.0-1.0
}

type SkipEvent struct {
	TrackID   uint      `json:"track_id"`
	SkippedAt time.Time `json:"skipped_at"`
	PlayedFor int       `json:"played_for"` // seconds
}

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

// UpdateVector moves the session vector based on play events
// Note: config dependency removed here to avoid circular dep if config was in session.
// We'll pass the necessary values directly.
func (s *SessionState) UpdateVector(completion float64, trackVector []float64, skipThreshold, listenThreshold, maxExplorationRate, minExplorationRate, learningRate, repulsionRate float64) {
	if completion < skipThreshold {
		// Skip detected: repel from track vector
		s.CurrentVector = VectorMove(s.CurrentVector, trackVector, repulsionRate)
		// Increase exploration if skips are high
		s.ExplorationRate = math.Min(maxExplorationRate, s.ExplorationRate+0.02)
	} else if completion > listenThreshold {
		// Full listen: attract toward track vector
		s.CurrentVector = VectorMove(s.CurrentVector, trackVector, learningRate)
		// Decay exploration if listens are high
		s.ExplorationRate = math.Max(minExplorationRate, s.ExplorationRate-0.01)
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

// ProjectTo2D uses a simplified PCA-like projection to map 128D to 2D
func ProjectTo2D(vector []float64) (x, y float64) {
	if len(vector) < 2 {
		return 0, 0
	}

	mid := len(vector) / 2
	for i := 0; i < mid; i++ {
		x += vector[i]
	}
	for i := mid; i < len(vector); i++ {
		y += vector[i]
	}

	x = math.Mod(math.Abs(x)*10, 100)
	y = math.Mod(math.Abs(y)*10, 100)
	
	return x, y
}

type JourneyPoint struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Mode string  `json:"mode"`
}

func (s *SessionState) GetJourney() []JourneyPoint {
	ox, oy := ProjectTo2D(s.OriginVector)
	points := []JourneyPoint{{X: ox, Y: oy, Mode: "start"}}

	cx, cy := ProjectTo2D(s.CurrentVector)
	mode := "exploitation"
	if s.ExplorationRate > 0.15 {
		mode = "exploration"
	}
	points = append(points, JourneyPoint{X: cx, Y: cy, Mode: mode})

	return points
}
