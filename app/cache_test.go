package app

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// waitFor polls cond until it returns true or the timeout elapses.
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

// expireNow backdates an entry so it is expired without waiting out the TTL.
func expireNow(c *searchCache, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		return
	}
	entry.expiresAt = time.Now().Add(-time.Second)
	c.entries[key] = entry
}

func TestSearchCacheSweepRemovesExpiredEntries(t *testing.T) {
	c := newSearchCache(15 * time.Minute)
	c.set("a", []byte(`[1]`))
	c.set("b", []byte(`[2]`))

	// Sweeping before the TTL expires must keep both entries.
	c.sweep(time.Now().Add(time.Minute))
	if got := c.len(); got != 2 {
		t.Fatalf("expected 2 entries before TTL, got %d", got)
	}

	// Sweeping past the TTL must drop both.
	c.sweep(time.Now().Add(16 * time.Minute))
	if got := c.len(); got != 0 {
		t.Fatalf("expected 0 entries after TTL, got %d", got)
	}
	if _, found := c.get("a"); found {
		t.Error("expected expired entry to be a miss")
	}
}

func TestSearchCacheEvictsOldestAtCapacity(t *testing.T) {
	c := newSearchCache(15 * time.Minute)
	c.maxEntries = 3

	for _, key := range []string{"a", "b", "c", "d"} {
		c.set(key, []byte(`[]`))
		// Distinct expiry timestamps so "oldest" is unambiguous.
		time.Sleep(time.Millisecond)
	}

	if got := c.len(); got != 3 {
		t.Fatalf("expected capacity 3, got %d entries", got)
	}
	if _, found := c.get("a"); found {
		t.Error("expected the oldest entry to be evicted, but it is still cached")
	}
	for _, key := range []string{"b", "c", "d"} {
		if _, found := c.get(key); !found {
			t.Errorf("expected %q to still be cached", key)
		}
	}
}

func TestSearchCacheCapacityKeepsRecentEntries(t *testing.T) {
	c := newSearchCache(time.Minute)
	c.maxEntries = 5

	for i := 0; i < 50; i++ {
		c.set(fmt.Sprintf("k%02d", i), []byte(`[]`))
	}

	if got := c.len(); got > 5 {
		t.Fatalf("cache exceeded its capacity: %d entries", got)
	}
	if _, found := c.get("k49"); !found {
		t.Error("expected the most recent entry to be cached")
	}
	if _, found := c.get("k44"); found {
		t.Error("expected entries beyond capacity to be evicted")
	}
}

func TestSearchCacheCapacityPrefersExpiredEntries(t *testing.T) {
	c := newSearchCache(time.Minute)
	c.maxEntries = 2

	c.set("fresh", []byte(`[]`))
	c.set("stale", []byte(`[]`))
	expireNow(c, "stale")

	// "stale" is expired, so inserting at capacity must drop it and keep "fresh".
	c.set("newest", []byte(`[]`))

	if _, found := c.get("fresh"); !found {
		t.Error("expected the live entry to survive capacity enforcement")
	}
	if _, found := c.get("newest"); !found {
		t.Error("expected the new entry to be cached")
	}
	if got := c.len(); got != 2 {
		t.Fatalf("expected 2 entries, got %d", got)
	}
}

func TestSearchCacheCleanupSweepsAndStopsOnContextCancel(t *testing.T) {
	c := newSearchCache(time.Minute)
	c.cleanupInterval = time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	c.StartCleanup(ctx)

	c.set("stale", []byte(`[]`))
	expireNow(c, "stale")

	// The cleanup loop must actually sweep expired entries.
	waitFor(t, 2*time.Second, func() bool { return c.len() == 0 })

	cancel()

	// ...and it must exit once the context is cancelled.
	select {
	case <-c.stopSignal():
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup goroutine did not exit after context cancellation")
	}
}

func TestSearchCacheReplacingKeyAtCapacityKeepsOtherEntries(t *testing.T) {
	c := newSearchCache(time.Minute)
	c.maxEntries = 2

	c.set("a", []byte(`[a1]`))
	time.Sleep(time.Millisecond)
	c.set("b", []byte(`[b]`))
	time.Sleep(time.Millisecond)

	// Replacing an existing key must not evict unrelated live entries.
	c.set("a", []byte(`[a2]`))

	if got := c.len(); got != 2 {
		t.Fatalf("expected 2 entries, got %d", got)
	}
	if _, found := c.get("b"); !found {
		t.Error("expected the unrelated entry to survive a replacement")
	}
	data, found := c.get("a")
	if !found || string(data) != `[a2]` {
		t.Errorf("expected the replacement value to be cached, got %q (found=%v)", data, found)
	}
}

func TestSearchCacheEmptyKeyRespectsCapacity(t *testing.T) {
	c := newSearchCache(time.Minute)
	c.maxEntries = 1

	// An empty key is a legal cache key and must not defeat the bound.
	c.set("", []byte(`[empty]`))
	if got := c.len(); got != 1 {
		t.Fatalf("expected 1 entry, got %d", got)
	}

	time.Sleep(time.Millisecond)
	c.set("k", []byte(`[k]`))

	if got := c.len(); got != 1 {
		t.Fatalf("expected the empty key to be evictable and the bound respected, got %d entries", got)
	}
	if _, found := c.get(""); found {
		t.Error("expected the oldest (empty) key to be evicted")
	}
	if _, found := c.get("k"); !found {
		t.Error("expected the new key to be cached")
	}
}

func TestSearchCacheStartCleanupIsSingleton(t *testing.T) {
	c := newSearchCache(time.Minute)
	c.cleanupInterval = time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	c.StartCleanup(ctx)

	first := c.stopSignal()
	if first == nil {
		t.Fatal("expected a stop signal for the running cleanup loop")
	}

	// Repeated and concurrent calls must not start another loop: every caller
	// has to observe the same cleanup lifetime.
	var wg sync.WaitGroup
	signals := make([]<-chan struct{}, 8)
	for i := range signals {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.StartCleanup(ctx)
			signals[i] = c.stopSignal()
		}(i)
	}
	wg.Wait()

	for i, s := range signals {
		if s != first {
			t.Errorf("goroutine %d observed a different cleanup loop than the first caller", i)
		}
	}

	cancel()
	select {
	case <-first:
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup loop did not exit after context cancellation")
	}

	// Once the loop has stopped, a new one may be started.
	ctx2, cancel2 := context.WithCancel(context.Background())
	t.Cleanup(cancel2)
	c.StartCleanup(ctx2)

	second := c.stopSignal()
	if second == nil || second == first {
		t.Fatal("expected a new cleanup loop after the previous one stopped")
	}

	cancel2()
	select {
	case <-second:
	case <-time.After(2 * time.Second):
		t.Fatal("second cleanup loop did not exit after context cancellation")
	}
}
