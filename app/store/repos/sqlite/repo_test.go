package sqlite_test

import (
	"os"
	"testing"
	"time"

	"github.com/david22573/GoRadio/app/store"
	"github.com/david22573/GoRadio/app/store/repos/sqlite"
	"github.com/david22573/GoRadio/app/types"
)

// setupRepo creates a new SQLiteRepo in a temporary directory,
// registers t.Cleanup to restore cwd and close the DB.
func setupRepo(t *testing.T) store.RadioRepository {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}
	// ensure we chdir back and close the DB
	var repo store.RadioRepository
	var closer interface{ Close() error }

	t.Cleanup(func() {
		// restore cwd
		os.Chdir(oldWd)
		// close the DB if we can
		if closer != nil {
			closer.Close()
		}
	})

	var err2 error
	repo, err2 = sqlite.NewSQLiteRepo("test.db")
	if err2 != nil {
		t.Fatalf("failed to create repo: %v", err2)
	}
	// capture the closer
	if c, ok := repo.(interface{ Close() error }); ok {
		closer = c
	}
	return repo
}

func TestStationCRUD(t *testing.T) {
	repo := setupRepo(t)

	// Create
	st := types.Station{Name: "TestStation", URL: "https://example.com"}
	if err := repo.CreateStation(&st); err != nil {
		t.Fatalf("CreateStation error: %v", err)
	}
	if st.ID == 0 {
		t.Fatalf("expected station ID to be set, got 0")
	}

	// GetByID
	gotByID, err := repo.GetStationByID(st.ID)
	if err != nil {
		t.Fatalf("GetStationByID error: %v", err)
	}
	if gotByID.Name != st.Name || gotByID.URL != st.URL {
		t.Errorf("GetStationByID mismatch, got %+v, want %+v", gotByID, st)
	}

	// GetByName
	gotByName, err := repo.GetStationByName(st.Name)
	if err != nil {
		t.Fatalf("GetStationByName error: %v", err)
	}
	if gotByName.ID != st.ID {
		t.Errorf("GetStationByName ID mismatch, got %d, want %d", gotByName.ID, st.ID)
	}

	// GetAll
	all, err := repo.GetAllStations()
	if err != nil {
		t.Fatalf("GetAllStations error: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("GetAllStations length, got %d, want 1", len(all))
	}

	// Update
	st.Name = "UpdatedStation"
	st.URL = "https://updated.example.com"
	if err := repo.UpdateStation(&st); err != nil {
		t.Fatalf("UpdateStation error: %v", err)
	}
	updated, err := repo.GetStationByID(st.ID)
	if err != nil {
		t.Fatalf("GetStationByID after update error: %v", err)
	}
	if updated.Name != st.Name || updated.URL != st.URL {
		t.Errorf("updated station mismatch, got %+v, want %+v", updated, st)
	}

	// Delete
	if err := repo.DeleteStation(st.ID); err != nil {
		t.Fatalf("DeleteStation error: %v", err)
	}
	_, err = repo.GetStationByID(st.ID)
	if err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
	time.Sleep(100 * time.Second)
}

func TestShowCRUD(t *testing.T) {
	repo := setupRepo(t)

	// First create a station
	st := types.Station{Name: "StationA", URL: "https://a.example.com"}
	if err := repo.CreateStation(&st); err != nil {
		t.Fatalf("CreateStation error: %v", err)
	}

	// Create Show
	s := types.Show{
		Name:         "Morning Show",
		Duration:     30 * time.Minute,
		ShowSchedule: types.ShowSchedule{Day: time.Tuesday, Hour: 9, Min: 0},
		StationID:    st.ID,
	}
	if err := repo.CreateShow(&s); err != nil {
		t.Fatalf("CreateShow error: %v", err)
	}
	if s.ID == 0 {
		t.Fatalf("expected show ID to be set, got 0")
	}

	// GetAllShows
	allShows, err := repo.GetAllShows()
	if err != nil {
		t.Fatalf("GetAllShows error: %v", err)
	}
	if len(allShows) != 1 {
		t.Errorf("expected 1 show, got %d", len(allShows))
	}

	// GetAllShowsByStation
	stationShows, err := repo.GetAllShowsByStation(st.ID)
	if err != nil {
		t.Fatalf("GetAllShowsByStation error: %v", err)
	}
	if len(stationShows) != 1 {
		t.Errorf("expected 1 station show, got %d", len(stationShows))
	}

	// GetByID
	gotS, err := repo.GetShowByID(s.ID)
	if err != nil {
		t.Fatalf("GetShowByID error: %v", err)
	}
	if gotS.Name != s.Name || gotS.StationID != s.StationID {
		t.Errorf("GetShowByID mismatch, got %+v, want %+v", gotS, s)
	}

	// Update Show
	s.Name = "Updated Show"
	s.Duration = 45 * time.Minute
	s.ShowSchedule = types.ShowSchedule{Day: time.Friday, Hour: 14, Min: 30}
	if err := repo.UpdateShow(&s); err != nil {
		t.Fatalf("UpdateShow error: %v", err)
	}
	updatedS, err := repo.GetShowByID(s.ID)
	if err != nil {
		t.Fatalf("GetShowByID after update error: %v", err)
	}
	if updatedS.Name != s.Name {
		t.Errorf("expected name %q, got %q", s.Name, updatedS.Name)
	}

	// Delete Show
	if err := repo.DeleteShow(s.ID); err != nil {
		t.Fatalf("DeleteShow error: %v", err)
	}
	_, err = repo.GetShowByID(s.ID)
	if err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}
