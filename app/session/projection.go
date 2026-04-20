package session

import (
	"math"
)

// ProjectTo2D uses a simplified PCA-like projection to map 128D to 2D
// For a real production app, we'd use a more robust library or t-SNE
func ProjectTo2D(vector []float64) (x, y float64) {
	if len(vector) < 2 {
		return 0, 0
	}

	// Sum of first half for X, second half for Y (extremely simplified projection)
	mid := len(vector) / 2
	for i := 0; i < mid; i++ {
		x += vector[i]
	}
	for i := mid; i < len(vector); i++ {
		y += vector[i]
	}

	// Normalize to a roughly 0-100 range for the canvas
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
	// Start with origin
	ox, oy := ProjectTo2D(s.OriginVector)
	points := []JourneyPoint{{X: ox, Y: oy, Mode: "start"}}

	// Add history (we'd need to store vectors in history for a true path)
	// For now, we show the current position as the target
	cx, cy := ProjectTo2D(s.CurrentVector)
	mode := "exploitation"
	if s.ExplorationRate > 0.15 {
		mode = "exploration"
	}
	points = append(points, JourneyPoint{X: cx, Y: cy, Mode: mode})

	return points
}
