package similarity

import (
	"context"
	"sync"
	"time"

	"github.com/david22573/GoRadio/app/db/sqlite"
	"github.com/david22573/GoRadio/app/types"
)

type Engine struct {
	db *sqlite.Client
	
	cache *VectorCache
	mu    sync.RWMutex
}

func NewEngine(db *sqlite.Client) *Engine {
	return &Engine{
		db:    db,
		cache: NewVectorCache(time.Hour),
	}
}

// FindNearestByVector finds K nearest tracks for a given raw vector
func (e *Engine) FindNearestByVector(ctx context.Context, vector []float64, k int, excludeIDs []uint) ([]types.Track, []float64, error) {
	// Simple caching: only if no exclusions for now
	// In production, we'd use a more sophisticated key
	
	neighbors, err := e.db.SearchKNN(vector, k+len(excludeIDs), sqlite.DistanceL2)
	if err != nil {
		return nil, nil, err
	}

	excludeMap := make(map[uint]bool)
	for _, id := range excludeIDs {
		excludeMap[id] = true
	}

	var resultTracks []types.Track
	for _, id := range neighbors {
		if excludeMap[id] {
			continue
		}
		track, err := e.db.GetTrackByID(id)
		if err == nil {
			resultTracks = append(resultTracks, *track)
		}
		if len(resultTracks) >= k {
			break
		}
	}

	return resultTracks, nil, nil
}

// FindExplorationByVector finds K tracks furthest from the given vector
func (e *Engine) FindExplorationByVector(ctx context.Context, vector []float64, k int, excludeIDs []uint) ([]types.Track, error) {
	tracks, err := e.db.GetDistantTracks(vector, k+len(excludeIDs))
	if err != nil {
		return nil, err
	}

	excludeMap := make(map[uint]bool)
	for _, id := range excludeIDs {
		excludeMap[id] = true
	}

	var result []types.Track
	for _, t := range tracks {
		if excludeMap[t.ID] {
			continue
		}
		result = append(result, t)
		if len(result) >= k {
			break
		}
	}
	return result, nil
}

func (e *Engine) FindNearestNeighbors(ctx context.Context, trackID uint, k int, excludeIDs []uint) ([]types.Track, []float64, error) {
	embedding, err := e.db.GetVectorByID(trackID)
	if err != nil {
		return nil, nil, err
	}
	return e.FindNearestByVector(ctx, embedding, k, excludeIDs)
}

func NormalizeScore(distance float64, metric sqlite.DistanceMetric) float64 {
	if metric == sqlite.DistanceCosine {
		return 1.0 - distance
	}
	return 1.0 / (1.0 + distance)
}
