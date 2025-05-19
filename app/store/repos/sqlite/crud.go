package sqlite

import (
	"github.com/david22573/GoRadio/app/store"
	"github.com/david22573/GoRadio/app/types"
)

// GetAllStations returns all stations.
func (r *sqliteRepo) GetAllStations() ([]types.Station, error) {
	query := `SELECT id, name, url FROM stations ORDER BY name`
	return queryMultiple(r.db, scanStation, query)
}

// GetStationByName loads a station by its primary name.
func (r *sqliteRepo) GetStationByName(name string) (*types.Station, error) {
	query := `SELECT id, name, url FROM stations WHERE name = ?`
	return querySingle(r.db, scanStation, query, name)
}

// GetStationByID loads a station by its primary key.
func (r *sqliteRepo) GetStationByID(id int) (*types.Station, error) {
	query := `SELECT id, name, url FROM stations WHERE id = ?`
	return querySingle(r.db, scanStation, query, id)
}

// CreateStation inserts a new station.
func (r *sqliteRepo) CreateStation(station *types.Station) error {
	query := `INSERT INTO stations (name, url) VALUES (?, ?)`
	id, err := execInsert(r.db, query, station.Name, station.URL)
	if err != nil {
		// You might want to check for specific errors, e.g., UNIQUE constraint violations
		return err
	}
	station.ID = int(id)
	return nil
}

// UpdateStation modifies an existing station.
func (r *sqliteRepo) UpdateStation(station *types.Station) error {
	query := `UPDATE stations SET name = ?, url = ? WHERE id = ?`
	_, err := execAffected(r.db, query, station.Name, station.URL, station.ID)
	if err != nil {
		// You might want to check for specific errors, e.g., UNIQUE constraint violations
		return err
	}
	return nil
}

// DeleteStation removes a station by ID.
func (r *sqliteRepo) DeleteStation(id int) error {
	query := `DELETE FROM stations WHERE id = ?`
	_, err := execAffected(r.db, query, id)
	return err
}

// GetAllShows returns every show in the DB.
func (r *sqliteRepo) GetAllShows() ([]types.Show, error) {
	query := `SELECT id, name, duration, day, hour, min, station_id FROM shows ORDER BY station_id, day, hour, min`
	return queryMultiple(r.db, scanShow, query)
}

// GetAllShowsByStation returns shows filtered by station ID.
func (r *sqliteRepo) GetAllShowsByStation(stationID int) ([]types.Show, error) {
	query := `SELECT id, name, duration, day, hour, min, station_id FROM shows WHERE station_id = ? ORDER BY day, hour, min`
	return queryMultiple(r.db, scanShow, query, stationID)
}

// GetShowByID loads a single show.
func (r *sqliteRepo) GetShowByID(id int) (*types.Show, error) {
	query := `SELECT id, name, duration, day, hour, min, station_id FROM shows WHERE id = ?`
	return querySingle(r.db, scanShow, query, id)
}

// GetStationByName loads a station by its primary name.
func (r *sqliteRepo) GetShowByName(name string) (*types.Show, error) {
	query := `SELECT id, name, duration, day, hour, min, station_id FROM shows WHERE id = ?`
	return querySingle(r.db, scanShow, query, name)
}

// CreateShow inserts a new show.
func (r *sqliteRepo) CreateShow(s *types.Show) error {
	query := `INSERT INTO shows (name, duration, day, hour, min, station_id) VALUES (?, ?, ?, ?, ?, ?)`
	id, err := execInsert(r.db, query,
		s.Name,
		int64(s.Duration.Seconds()),
		int(s.Day), // s.Day from types.ShowSchedule (embedded)
		s.Hour,     // s.Hour from types.ShowSchedule
		s.Min,      // s.Min from types.ShowSchedule
		s.StationID,
	)
	if err != nil {
		return err
	}
	s.ID = int(id)
	return nil
}

// UpdateShow modifies an existing show.
func (r *sqliteRepo) UpdateShow(s *types.Show) error {
	query := `UPDATE shows SET name = ?, duration = ?, day = ?, hour = ?, min = ?, station_id = ? WHERE id = ?`
	affected, err := execAffected(r.db, query,
		s.Name,
		int64(s.Duration.Seconds()),
		int(s.Day),
		s.Hour,
		s.Min,
		s.StationID,
		s.ID, // s.ID for WHERE clause
	)
	if err != nil {
		return err
	}
	if affected == 0 {
		return store.ErrNotFound // No rows updated, implies ID not found
	}
	return nil
}

// DeleteShow removes a show by ID.
func (r *sqliteRepo) DeleteShow(id int) error {
	query := `DELETE FROM shows WHERE id = ?`
	affected, err := execAffected(r.db, query, id)
	if err != nil {
		return err
	}
	if affected == 0 {
		return store.ErrNotFound // No rows deleted, implies ID not found
	}
	return nil
}
