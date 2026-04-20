package sqlite_test

import (
	"os"
	"testing"
	"time"

	"github.com/david22573/GoRadio/app/db/sqlite"
	"github.com/david22573/GoRadio/app/types"
)

func setupTestClient(t *testing.T) *sqlite.Client {
	t.Helper()
	dbFile := "test_db.sqlite"
	_ = os.Remove("data/" + dbFile)

	client, err := sqlite.NewSQLiteClient(dbFile)
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
		`CREATE TABLE IF NOT EXISTS shows (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT NOT NULL UNIQUE,
			duration   INTEGER NOT NULL,
			day        INTEGER NOT NULL,
			hour       INTEGER NOT NULL,
			min        INTEGER NOT NULL,
			scheduled  INTEGER NOT NULL DEFAULT 0,
			station_id INTEGER NOT NULL,
			FOREIGN KEY(station_id) REFERENCES stations(id) ON DELETE CASCADE
		)`,
	}
	for _, q := range queries {
		if _, err := client.GetDB().Exec(q); err != nil {
			t.Fatalf("failed to create table: %v", err)
		}
	}

	t.Cleanup(func() {
		client.Close()
		os.Remove("data/" + dbFile)
	})
	return client
}

func TestStationCRUD(t *testing.T) {
	client := setupTestClient(t)

	station := types.Station{Name: "Test Station", URL: "http://localhost"}
	err := client.CreateStation(station)
	if err != nil {
		t.Fatalf("CreateStation failed: %v", err)
	}

	fetched, err := client.GetStationByName("Test Station")
	if err != nil || fetched == nil || fetched.URL != station.URL {
		t.Fatalf("GetStationByName failed: %v", err)
	}

	fetched.URL = "http://updated"
	err = client.UpdateStation(fetched)
	if err != nil {
		t.Fatalf("UpdateStation failed: %v", err)
	}

	updated, _ := client.GetStationByID(fetched.ID)
	if updated.URL != "http://updated" {
		t.Errorf("expected URL to be updated, got %s", updated.URL)
	}

	err = client.DeleteStation(fetched.ID)
	if err != nil {
		t.Fatalf("DeleteStation failed: %v", err)
	}

	none, _ := client.GetStationByID(fetched.ID)
	if none != nil {
		t.Error("expected station to be deleted")
	}
}

func TestShowCRUD(t *testing.T) {
	client := setupTestClient(t)

	station := types.Station{Name: "Station for Show", URL: "http://station"}
	err := client.CreateStation(station)
	if err != nil {
		t.Fatalf("CreateStation for Show failed: %v", err)
	}
	st, _ := client.GetStationByName("Station for Show")

	show := types.Show{
		Name:      "Test Show",
		Duration:  types.ShowDuration{Duration: 30 * time.Minute},
		Day:       2,
		Hour:      10,
		Min:       15,
		Scheduled: true,
		StationID: st.ID,
	}
	err = client.CreateShow(show)
	if err != nil {
		t.Fatalf("CreateShow failed: %v", err)
	}

	fetched, err := client.GetShowByName("Test Show")
	if err != nil || fetched == nil || fetched.Hour != 10 {
		t.Fatalf("GetShowByName failed: %v", err)
	}

	fetched.Hour = 12
	err = client.UpdateShow(fetched)
	if err != nil {
		t.Fatalf("UpdateShow failed: %v", err)
	}

	updated, _ := client.GetShowByID(fetched.ID)
	if updated.Hour != 12 {
		t.Errorf("expected hour to be updated, got %d", updated.Hour)
	}

	err = client.DeleteShow(fetched.ID)
	if err != nil {
		t.Fatalf("DeleteShow failed: %v", err)
	}

	none, _ := client.GetShowByID(fetched.ID)
	if none != nil {
		t.Error("expected show to be deleted")
	}
}
