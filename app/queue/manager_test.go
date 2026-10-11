package queue

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestEvictIdleQueues(t *testing.T) {
	now := time.Now()
	m := &Manager{
		queues: map[string]*Queue{
			"idle":   {SessionID: "idle", LastAccess: now.Add(-2 * time.Hour)},
			"active": {SessionID: "active", LastAccess: now.Add(-time.Minute)},
		},
		ttl: time.Hour,
	}

	m.evictIdleQueues(now)

	if m.size() != 1 {
		t.Fatalf("expected 1 remaining queue, got %d", m.size())
	}
	if _, ok := m.queues["active"]; !ok {
		t.Error("expected the recently accessed queue to be kept")
	}
	if _, ok := m.queues["idle"]; ok {
		t.Error("expected the idle queue to be evicted")
	}
}

func TestCleanupLoopEvictsAndStopsOnContextCancel(t *testing.T) {
	m := &Manager{
		queues: map[string]*Queue{
			"idle": {SessionID: "idle", LastAccess: time.Now().Add(-2 * time.Hour)},
		},
		ttl:             time.Hour,
		cleanupInterval: time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		m.cleanupLoop(ctx)
		close(done)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for m.size() != 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if m.size() != 0 {
		t.Fatal("cleanup loop did not evict the idle queue")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup loop did not exit after context cancellation")
	}
}

func TestEvictionDoesNotDropActivelyUsedQueues(t *testing.T) {
	// ttl is an hour so eligibility is deterministic: only the idle queue can be
	// evicted, while "active" is refreshed concurrently by another goroutine.
	m := &Manager{
		queues: make(map[string]*Queue),
		ttl:    time.Hour,
	}
	active := &Queue{SessionID: "active", LastAccess: time.Now()}
	m.queues["active"] = active
	m.queues["idle"] = &Queue{SessionID: "idle", LastAccess: time.Now().Add(-2 * time.Hour)}

	// Hammer the access timestamp while the cleanup loop evicts, so the
	// read/write pair on the eviction clock is exercised concurrently.
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				m.GetQueue("active")
				runtime.Gosched()
			}
		}
	}()

	for i := 0; i < 1000; i++ {
		m.evictIdleQueues(time.Now())
	}

	close(stop)
	wg.Wait()

	if _, ok := m.queues["idle"]; ok {
		t.Error("expected the idle queue to be evicted")
	}
	if _, ok := m.queues["active"]; !ok {
		t.Error("expected the actively used queue to survive eviction")
	}
}
