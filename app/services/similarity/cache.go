package similarity

import (
	"context"
	"sync"
	"time"

	"github.com/david22573/GoRadio/app/types"
)

const (
	// defaultVectorCacheMaxEntries bounds memory growth from distinct
	// neighborhood lookups.
	defaultVectorCacheMaxEntries = 1000
	// defaultVectorCacheCleanupInterval is how often expired entries are swept.
	defaultVectorCacheCleanupInterval = 15 * time.Minute
)

type CacheEntry struct {
	Tracks    []types.Track
	ExpiresAt time.Time
}

type VectorCache struct {
	mu              sync.RWMutex
	data            map[string]CacheEntry
	ttl             time.Duration
	maxEntries      int
	cleanupInterval time.Duration
	// stopped is closed when the cleanup goroutine exits (ctx cancelled).
	stopped chan struct{}
}

func NewVectorCache(ctx context.Context, ttl time.Duration) *VectorCache {
	vc := &VectorCache{
		data:            make(map[string]CacheEntry),
		ttl:             ttl,
		maxEntries:      defaultVectorCacheMaxEntries,
		cleanupInterval: defaultVectorCacheCleanupInterval,
		stopped:         make(chan struct{}),
	}
	go vc.cleanupLoop(ctx)
	return vc
}

func (vc *VectorCache) cleanupLoop(ctx context.Context) {
	interval := vc.cleanupInterval
	if interval <= 0 {
		interval = defaultVectorCacheCleanupInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer close(vc.stopped)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			vc.sweep(time.Now())
		}
	}
}

// sweep drops every entry that has expired as of now.
func (vc *VectorCache) sweep(now time.Time) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.dropExpiredLocked(now)
}

func (vc *VectorCache) dropExpiredLocked(now time.Time) {
	for k, v := range vc.data {
		if now.After(v.ExpiresAt) {
			delete(vc.data, k)
		}
	}
}

// evictOldestLocked removes the entry closest to expiry. Entries expire in
// insertion order, so this approximates FIFO eviction.
func (vc *VectorCache) evictOldestLocked() {
	var (
		oldestKey  string
		oldestTime time.Time
		found      bool
	)
	for k, v := range vc.data {
		if !found || v.ExpiresAt.Before(oldestTime) {
			oldestKey, oldestTime, found = k, v.ExpiresAt, true
		}
	}
	if found {
		delete(vc.data, oldestKey)
	}
}

// makeRoomLocked frees one slot for a new key: it reclaims expired entries
// first and only then evicts the entry closest to expiry, so the cache degrades
// one entry at a time instead of being cleared wholesale. A non-positive
// maxEntries disables the bound explicitly.
func (vc *VectorCache) makeRoomLocked(now time.Time) {
	if vc.maxEntries <= 0 || len(vc.data) < vc.maxEntries {
		return
	}
	vc.dropExpiredLocked(now)
	if len(vc.data) < vc.maxEntries {
		return
	}
	vc.evictOldestLocked()
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

	// Replacing an existing key must not evict an unrelated live entry.
	if _, exists := vc.data[key]; !exists {
		vc.makeRoomLocked(time.Now())
	}

	vc.data[key] = CacheEntry{
		Tracks:    tracks,
		ExpiresAt: time.Now().Add(vc.ttl),
	}
}

// size returns the number of cached entries.
func (vc *VectorCache) size() int {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return len(vc.data)
}

// stopSignal returns a channel that is closed when the cleanup loop exits.
func (vc *VectorCache) stopSignal() <-chan struct{} {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return vc.stopped
}

// Warm pre-computes neighborhoods for popular tracks (placeholder)
func (vc *VectorCache) Warm(popularIDs []uint, engine interface{}) {
	// Implementation would call engine.FindNearestByVector for each ID
}
