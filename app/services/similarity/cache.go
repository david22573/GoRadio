package similarity

import (
	"sync"
	"time"

	"github.com/david22573/GoRadio/app/types"
)

type CacheEntry struct {
	Tracks    []types.Track
	ExpiresAt time.Time
}

type VectorCache struct {
	mu    sync.RWMutex
	data  map[string]CacheEntry
	ttl   time.Duration
}

func NewVectorCache(ttl time.Duration) *VectorCache {
	return &VectorCache{
		data: make(map[string]CacheEntry),
		ttl:  ttl,
	}
}

func (vc *VectorCache) Get(key string) ([]types.Track, bool) {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	
	entry, ok := vc.data[key]
	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Tracks, true
}

func (vc *VectorCache) Set(key string, tracks []types.Track) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	
	vc.data[key] = CacheEntry{
		Tracks:    tracks,
		ExpiresAt: time.Now().Add(vc.ttl),
	}
}

// Warm pre-computes neighborhoods for popular tracks (placeholder)
func (vc *VectorCache) Warm(popularIDs []uint, engine interface{}) {
	// Implementation would call engine.FindNearestByVector for each ID
}
