package schedulers

import (
	"os"
	"testing"
	"time"

	"github.com/david22573/GoRadio/app"
	"github.com/david22573/GoRadio/app/store/repos/sqlite"
	"github.com/david22573/GoRadio/app/types"
)

func setupSQLiteRepo(t *testing.T) *sqlite.SqliteRepo {
	// isolate on temp dir
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		// restore cwd + close DB
		os.Chdir(old)
	})

	repoIntf, err := sqlite.NewSQLiteRepo("test.db")
	if err != nil {
		t.Fatalf("NewSQLiteRepo: %v", err)
	}
	// repoIntf is a store.RadioRepository, but we know it's *sqliteRepo
	repo, ok := repoIntf.(*sqlite.SqliteRepo)
	if !ok {
		t.Fatalf("expected *SQLiteRepo, got %T", repoIntf)
	}
	t.Cleanup(func() {
		repo.Close()
	})
	return repo
}

func setupAppWithData(t *testing.T) (*app.App, types.Station) {
	repo := setupSQLiteRepo(t)

	// create one station
	st := types.Station{Name: "TestStation", URL: "https://ex.com"}
	if err := repo.CreateStation(&st); err != nil {
		t.Fatalf("CreateStation: %v", err)
	}

	// create two shows on different IDs/times
	shows := []types.Show{
		{
			Name:         "Show A",
			Duration:     15 * time.Minute,
			ShowSchedule: types.ShowSchedule{Day: time.Monday, Hour: 10, Min: 0},
			StationID:    st.ID,
		},
		{
			Name:         "Show B",
			Duration:     30 * time.Minute,
			ShowSchedule: types.ShowSchedule{Day: time.Tuesday, Hour: 11, Min: 30},
			StationID:    st.ID,
		},
	}
	for i := range shows {
		if err := repo.CreateShow(&shows[i]); err != nil {
			t.Fatalf("CreateShow #%d: %v", i, err)
		}
	}
	return &app.App{Repo: repo}, st
}

func TestNewAndCancelScheduler(t *testing.T) {
	app, st := setupAppWithData(t)

	// use a fixed location so tests deterministic
	loc := time.UTC
	rs, err := NewRadioScheduler(app, st, loc)
	if err != nil {
		t.Fatalf("NewRadioScheduler: %v", err)
	}

	// we should have one job per show
	if len(rs.jobs) != 2 {
		t.Errorf("jobs count = %d; want 2", len(rs.jobs))
	}
	// collect IDs present
	present := make(map[uint]struct{}, len(rs.jobs))
	for id := range rs.jobs {
		present[id] = struct{}{}
	}
	for _, want := range []uint{1, 2} {
		if _, ok := present[want]; !ok {
			t.Errorf("missing job for show ID %d", want)
		}
	}

	// Cancel a known show
	if err := rs.CancelShow(1); err != nil {
		t.Errorf("CancelShow(1) error: %v", err)
	}
	if _, still := rs.jobs[1]; still {
		t.Errorf("job 1 still present after CancelShow")
	}
	if len(rs.jobs) != 1 {
		t.Errorf("jobs count after cancel = %d; want 1", len(rs.jobs))
	}

	// Cancel unknown ID → error
	if err := rs.CancelShow(999); err == nil {
		t.Errorf("expected error CancelShow(999), got nil")
	}

	// Start() should not panic
	rs.Start()
}
