package similarity

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/david22573/GoRadio/app/types"
)

func newTestVectorCache(t *testing.T, ttl time.Duration) *VectorCache {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return NewVectorCache(ctx, ttl)
}

func tracks(ids ...uint) []types.Track {
	out := make([]types.Track, 0, len(ids))
	for _, id := range ids {
		out = append(out, types.Track{ID: id})
	}
	return out
}

func TestVectorCacheSweepRemovesExpiredEntries(t *testing.T) {
	vc := newTestVectorCache(t, 15*time.Minute)

	vc.Set("a", tracks(1))
	vc.Set("b", tracks(2))

	// Sweeping before the TTL expires must keep both entries.
	vc.sweep(time.Now().Add(time.Minute))
	if got := vc.size(); got != 2 {
		t.Fatalf("expected 2 entries before TTL, got %d", got)
	}

	// Sweeping past the TTL must drop both.
	vc.sweep(time.Now().Add(16 * time.Minute))
	if got := vc.size(); got != 0 {
		t.Fatalf("expected 0 entries after TTL, got %d", got)
	}
	if _, ok := vc.Get("a"); ok {
		t.Error("expected expired entry to be a miss")
	}
}

func TestVectorCacheCapacityKeepsRecentEntries(t *testing.T) {
	vc := newTestVectorCache(t, time.Hour)
	vc.maxEntries = 5

	for i := 0; i < 50; i++ {
		vc.Set(fmt.Sprintf("k%02d", i), tracks(uint(i)))
	}

	if got := vc.size(); got > 5 {
		t.Fatalf("cache exceeded its capacity: %d entries", got)
	}
}

func TestVectorCacheEvictsOldestAtCapacity(t *testing.T) {
	vc := newTestVectorCache(t, 15*time.Minute)
	vc.maxEntries = 3

	for _, key := range []string{"a", "b", "c", "d"} {
		vc.Set(key, tracks(1))
		// Distinct expiry timestamps so "oldest" is unambiguous.
		time.Sleep(time.Millisecond)
	}

	if got := vc.size(); got != 3 {
		t.Fatalf("expected capacity 3, got %d entries", got)
	}
	if _, ok := vc.Get("a"); ok {
		t.Error("expected the oldest entry to be evicted, but it is still cached")
	}
	for _, key := range []string{"b", "c", "d"} {
		if _, ok := vc.Get(key); !ok {
			t.Errorf("expected %q to still be cached", key)
		}
	}
}

func TestVectorCacheCleanupStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	vc := NewVectorCache(ctx, time.Minute)

	cancel()

	select {
	case <-vc.stopSignal():
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup goroutine did not exit after context cancellation")
	}
}

func TestVectorCacheReplacingKeyAtCapacityKeepsOtherEntries(t *testing.T) {
	vc := newTestVectorCache(t, time.Minute)
	vc.maxEntries = 2

	vc.Set("a", tracks(1))
	time.Sleep(time.Millisecond)
	vc.Set("b", tracks(2))
	time.Sleep(time.Millisecond)

	// Replacing an existing key must not evict unrelated live entries.
	vc.Set("a", tracks(3))

	if got := vc.size(); got != 2 {
		t.Fatalf("expected 2 entries, got %d", got)
	}
	if _, ok := vc.Get("b"); !ok {
		t.Error("expected the unrelated entry to survive a replacement")
	}
	got, ok := vc.Get("a")
	if !ok || len(got) != 1 || got[0].ID != 3 {
		t.Errorf("expected the replacement value to be cached, got %+v (found=%v)", got, ok)
	}
}

func TestVectorCacheEmptyKeyRespectsCapacity(t *testing.T) {
	vc := newTestVectorCache(t, time.Minute)
	vc.maxEntries = 1

	// An empty key is legal through the public Set and must not defeat the bound.
	vc.Set("", tracks(1))
	if got := vc.size(); got != 1 {
		t.Fatalf("expected 1 entry, got %d", got)
	}

	time.Sleep(time.Millisecond)
	vc.Set("k", tracks(2))

	if got := vc.size(); got != 1 {
		t.Fatalf("expected the empty key to be evictable and the bound respected, got %d entries", got)
	}
	if _, ok := vc.Get(""); ok {
		t.Error("expected the oldest (empty) key to be evicted")
	}
	if _, ok := vc.Get("k"); !ok {
		t.Error("expected the new key to be cached")
	}
}
