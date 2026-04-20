package sqlite

import (
	"os"
	"testing"
	"time"

	"github.com/david22573/GoRadio/app/types"
)

func setupTestDB(t *testing.T) *Client {
	t.Helper()
	dbFile := "test_full.sqlite"
	_ = os.Remove("data/" + dbFile)

	client, err := NewSQLiteClient(dbFile)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	// Create tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS stations (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			url  TEXT NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS tracks (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			station_id  INTEGER NOT NULL,
			title       TEXT NOT NULL,
			artist      TEXT NOT NULL,
			url         TEXT NOT NULL,
			duration    INTEGER NOT NULL,
			analyzed_at DATETIME
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS track_vectors USING vec0(
			track_id INTEGER PRIMARY KEY,
			embedding FLOAT[128] distance_metric=l2
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id               TEXT PRIMARY KEY,
			user_id          INTEGER,
			created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_activity_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			current_vector   BLOB,
			origin_vector    BLOB,
			exploration_rate REAL DEFAULT 0.15
		)`,
		`CREATE TABLE IF NOT EXISTS session_events (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			track_id   INTEGER NOT NULL,
			event_type TEXT NOT NULL,
			completion REAL,
			played_for INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, q := range queries {
		if _, err := client.db.Exec(q); err != nil {
			// Skip if vec0 is not supported in the test environment (e.g. CI without extension)
			// but we should ideally have it.
			t.Logf("Warning: failed to run query: %v", err)
		}
	}

	t.Cleanup(func() {
		client.Close()
		os.Remove("data/" + dbFile)
	})
	return client
}

func TestTrackOperations(t *testing.T) {
	client := setupTestDB(t)

	track := types.Track{
		StationID:  1,
		Title:      "Test Track",
		Artist:     "Test Artist",
		Duration:   180,
		AnalyzedAt: time.Now(),
	}

	id, err := client.CreateTrack(track)
	if err != nil {
		t.Fatalf("CreateTrack failed: %v", err)
	}

	fetched, err := client.GetTrackByID(id)
	if err != nil || fetched == nil || fetched.Title != track.Title {
		t.Fatalf("GetTrackByID failed: %v", err)
	}

	tracks, err := client.GetTracksByStation(1)
	if err != nil || len(tracks) != 1 {
		t.Fatalf("GetTracksByStation failed: %v", err)
	}

	random, err := client.GetRandomTrack(nil)
	if err != nil || random == nil {
		t.Fatalf("GetRandomTrack failed: %v", err)
	}
}

func TestVectorOperations(t *testing.T) {
	client := setupTestDB(t)
	
	// Check if vec0 is available
	var count int
	err := client.db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name='track_vectors'").Scan(&count)
	if err != nil || count == 0 {
		t.Skip("sqlite-vec extension not available or table not created")
	}

	embedding := make([]float64, 128)
	embedding[0] = 1.0

	err = client.InsertTrackVector(1, embedding)
	if err != nil {
		t.Fatalf("InsertTrackVector failed: %v", err)
	}

	fetched, err := client.GetTrackVector(1)
	if err != nil || len(fetched) != 128 || fetched[0] != 1.0 {
		t.Fatalf("GetTrackVector failed: %v", err)
	}
}

func TestSessionOperations(t *testing.T) {
	client := setupTestDB(t)

	s := &types.SessionState{
		ID:              "test-session",
		CreatedAt:       time.Now(),
		LastActivityAt:  time.Now(),
		CurrentVector:   make([]float64, 128),
		OriginVector:    make([]float64, 128),
		ExplorationRate: 0.15,
	}

	err := client.CreateSession(s)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	fetched, err := client.GetSession(s.ID)
	if err != nil || fetched == nil || fetched.ID != s.ID {
		t.Fatalf("GetSession failed: %v", err)
	}

	s.ExplorationRate = 0.2
	err = client.UpdateSession(s)
	if err != nil {
		t.Fatalf("UpdateSession failed: %v", err)
	}

	updated, _ := client.GetSession(s.ID)
	if updated.ExplorationRate != 0.2 {
		t.Errorf("expected ExplorationRate to be 0.2, got %f", updated.ExplorationRate)
	}

	play := types.PlayEvent{TrackID: 1, Completion: 0.9, CompletedAt: time.Now()}
	err = client.RecordPlayEvent(s.ID, play)
	if err != nil {
		t.Fatalf("RecordPlayEvent failed: %v", err)
	}

	skip := types.SkipEvent{TrackID: 2, PlayedFor: 10, SkippedAt: time.Now()}
	err = client.RecordSkipEvent(s.ID, skip)
	if err != nil {
		t.Fatalf("RecordSkipEvent failed: %v", err)
	}

	plays, skips, err := client.GetSessionEvents(s.ID)
	if err != nil || len(plays) != 1 || len(skips) != 1 {
		t.Fatalf("GetSessionEvents failed: %v", err)
	}

	err = client.DeleteSession(s.ID)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}
}
