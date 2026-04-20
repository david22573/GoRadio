package session

import (
	"fmt"

	"github.com/david22573/GoRadio/app/db/sqlite"
)

// UpdateSessionVector moves the current session vector based on user behavior
func UpdateSessionVector(s *SessionState, db *sqlite.Client, trackID uint, behavior string) error {
	newTrackVector, err := db.GetVectorByID(trackID)
	if err != nil {
		return err
	}

	if len(s.CurrentVector) != len(newTrackVector) {
		return fmt.Errorf("vector dimension mismatch")
	}

	var alpha float64
	switch behavior {
	case "listen":
		alpha = 0.1 // Move towards listened tracks
	case "skip":
		alpha = -0.05 // Move away from skipped tracks
	default:
		return nil
	}

	// Evolution: V_new = (1-alpha)V_curr + alpha*V_track
	for i := range s.CurrentVector {
		s.CurrentVector[i] = (1-alpha)*s.CurrentVector[i] + alpha*newTrackVector[i]
	}

	return nil
}
