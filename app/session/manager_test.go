package session

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestEvictIdleSessions(t *testing.T) {
	now := time.Now()
	idle := &SessionState{ID: "idle", LastActivityAt: now.Add(-2 * time.Hour)}
	active := &SessionState{ID: "active", LastActivityAt: now.Add(-time.Minute)}

	m := &Manager{
		sessions: map[string]*SessionState{"idle": idle, "active": active},
		ttl:      time.Hour,
	}

	m.evictIdleSessions(now)

	if m.size() != 1 {
		t.Fatalf("expected 1 remaining session, got %d", m.size())
	}
	if _, ok := m.sessions["active"]; !ok {
		t.Error("expected the recently active session to be kept")
	}
	if _, ok := m.sessions["idle"]; ok {
		t.Error("expected the idle session to be evicted")
	}
}

func TestCleanupLoopEvictsAndStopsOnContextCancel(t *testing.T) {
	m := &Manager{
		sessions: map[string]*SessionState{
			"idle": {ID: "idle", LastActivityAt: time.Now().Add(-2 * time.Hour)},
		},
		ttl:               time.Hour,
		cleanupInterval:   time.Millisecond,
		dbCleanupInterval: time.Hour,
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
		t.Fatal("cleanup loop did not evict the idle session")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup loop did not exit after context cancellation")
	}
}

func TestEvictionDoesNotDropActivelyUsedSessions(t *testing.T) {
	// ttl is an hour so eligibility is deterministic: only the idle session can
	// be evicted, while "active" is refreshed concurrently by another goroutine.
	m := &Manager{
		sessions:       make(map[string]*SessionState),
		ttl:            time.Hour,
		persistSession: func(*SessionState) error { return nil },
	}
	active := &SessionState{ID: "active", LastActivityAt: time.Now()}
	m.sessions["active"] = active
	m.sessions["idle"] = &SessionState{ID: "idle", LastActivityAt: time.Now().Add(-2 * time.Hour)}

	// Hammer the activity timestamp while the cleanup loop evicts, so the
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
				m.touchSession("active")
				runtime.Gosched()
			}
		}
	}()

	for i := 0; i < 1000; i++ {
		m.evictIdleSessions(time.Now())
	}

	close(stop)
	wg.Wait()

	if _, ok := m.sessions["idle"]; ok {
		t.Error("expected the idle session to be evicted")
	}
	if _, ok := m.sessions["active"]; !ok {
		t.Error("expected the actively used session to survive eviction")
	}
}

func TestTouchSessionPersistsDetachedSnapshot(t *testing.T) {
	s := &SessionState{
		ID:             "s",
		CurrentVector:  []float64{1, 0},
		OriginVector:   []float64{1, 0},
		LastActivityAt: time.Now().Add(-time.Hour),
	}
	written := make(chan *SessionState, 4)
	m := &Manager{
		sessions:       map[string]*SessionState{"s": s},
		ttl:            time.Hour,
		persistSession: func(snap *SessionState) error { written <- snap; return nil },
	}

	before := time.Now()
	live, ok := m.touchSession("s")
	if !ok {
		t.Fatal("expected the cached session to be found")
	}
	if live != s {
		t.Error("expected the live cached session to be returned")
	}

	var snap *SessionState
	select {
	case snap = <-written:
	case <-time.After(2 * time.Second):
		t.Fatal("expected the refreshed session to be persisted")
	}
	if snap == s {
		t.Fatal("expected a detached snapshot for asynchronous persistence")
	}
	if !snap.LastActivityAt.After(before) {
		t.Error("expected the persisted snapshot to carry the refreshed activity clock")
	}

	// Mutating the live session must not change what the DB writer sees.
	m.mu.Lock()
	s.CurrentVector[0] = 99
	m.mu.Unlock()

	if snap.CurrentVector[0] != 1 {
		t.Errorf("snapshot shares vector storage with the live session: %v", snap.CurrentVector)
	}
}

func TestTouchSessionRefreshesTheTrackedEntry(t *testing.T) {
	s := &SessionState{ID: "s", LastActivityAt: time.Now().Add(-2 * time.Hour)}
	m := &Manager{
		sessions:       map[string]*SessionState{"s": s},
		ttl:            time.Hour,
		persistSession: func(*SessionState) error { return nil },
	}

	live, ok := m.touchSession("s")
	if !ok {
		t.Fatal("expected the cached session to be found")
	}
	if m.sessions["s"] != live {
		t.Error("expected touchSession to return the entry the map tracks")
	}

	// The refreshed clock must be visible to the eviction decision, otherwise a
	// session in active use would still look idle.
	m.evictIdleSessions(time.Now())
	if m.size() != 1 {
		t.Fatalf("expected the touched session to survive eviction, size=%d", m.size())
	}
}

func TestPersistSnapshotRejectsStaleWrites(t *testing.T) {
	t.Run("newer first", func(t *testing.T) {
		written := make(chan string, 4)
		m := &Manager{persistSession: func(snap *SessionState) error {
			written <- fmt.Sprintf("%s:%v", snap.ID, snap.CurrentVector[0])
			return nil
		}}

		m.persistSnapshot(&SessionState{ID: "s", CurrentVector: []float64{2}}, 2)
		m.persistSnapshot(&SessionState{ID: "s", CurrentVector: []float64{1}}, 1)

		if got := drainWrites(written); len(got) != 1 || got[0] != "s:2" {
			t.Fatalf("expected only the newest snapshot to be written, got %v", got)
		}
	})

	t.Run("older first", func(t *testing.T) {
		written := make(chan string, 4)
		m := &Manager{persistSession: func(snap *SessionState) error {
			written <- fmt.Sprintf("%s:%v", snap.ID, snap.CurrentVector[0])
			return nil
		}}

		m.persistSnapshot(&SessionState{ID: "s", CurrentVector: []float64{1}}, 1)
		m.persistSnapshot(&SessionState{ID: "s", CurrentVector: []float64{2}}, 2)

		if got := drainWrites(written); len(got) != 2 || got[0] != "s:1" || got[1] != "s:2" {
			t.Fatalf("expected both snapshots in order, got %v", got)
		}
	})
}

func TestEvolveVectorPersistsWithoutTrackVector(t *testing.T) {
	s := &SessionState{ID: "s", CurrentVector: []float64{1, 0}, LastActivityAt: time.Now().Add(-time.Minute)}
	written := make(chan *SessionState, 4)
	m := &Manager{
		sessions:       map[string]*SessionState{"s": s},
		ttl:            time.Hour,
		vectorFor:      func(uint) ([]float64, error) { return nil, nil },
		persistSession: func(snap *SessionState) error { written <- snap; return nil },
	}

	m.evolveVector("s", 7, 0.9)

	var snap *SessionState
	select {
	case snap = <-written:
	case <-time.After(2 * time.Second):
		t.Fatal("expected the activity update to be persisted even without a track vector")
	}
	if snap.ID != "s" {
		t.Errorf("expected the session to be persisted, got %q", snap.ID)
	}
	if len(snap.CurrentVector) == 0 || snap.CurrentVector[0] != 1 {
		t.Errorf("expected the vector to be left unchanged when no track vector is available, got %v", snap.CurrentVector)
	}
}

// drainWrites collects everything the writer has been given so far.
func drainWrites(written chan string) []string {
	var out []string
	for {
		select {
		case w := <-written:
			out = append(out, w)
		default:
			return out
		}
	}
}
