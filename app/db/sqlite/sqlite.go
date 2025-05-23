// File: app/sqlite/client.go

package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/david22573/GoRadio/app/types"

	_ "modernc.org/sqlite"
)

type Client struct {
	db *sql.DB
}

func NewSQLiteClient(filename string) (*Client, error) {
	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		return nil, fmt.Errorf("error creating data folder: %w", err)
	}

	dsn := fmt.Sprintf("data/%s", filename)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Client{db: db}, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

func migrate(db *sql.DB) error {
	schema := `
    CREATE TABLE IF NOT EXISTS stations (
        id   INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        url  TEXT NOT NULL UNIQUE
    );
    CREATE TABLE IF NOT EXISTS shows (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        name       TEXT NOT NULL UNIQUE,
        duration   INTEGER NOT NULL,
        day        INTEGER NOT NULL,
        hour       INTEGER NOT NULL,
        min        INTEGER NOT NULL,
        scheduled  INTEGER NOT NULL DEFAULT 0,
        station_id INTEGER NOT NULL,
        FOREIGN KEY(station_id) REFERENCES stations(id) ON DELETE CASCADE
    );`
	_, err := db.Exec(schema)
	return err
}

// Implement db.Client methods:

func (c *Client) GetAllStations() ([]types.Station, error) {
	rows, err := c.db.Query("SELECT id, name, url FROM stations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stations []types.Station
	for rows.Next() {
		var s types.Station
		if err := rows.Scan(&s.ID, &s.Name, &s.URL); err != nil {
			return nil, err
		}
		stations = append(stations, s)
	}
	return stations, nil
}

func (c *Client) GetStationByID(id uint) (*types.Station, error) {
	var s types.Station
	err := c.db.QueryRow("SELECT id, name, url FROM stations WHERE id = ?", id).
		Scan(&s.ID, &s.Name, &s.URL)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) GetStationByName(name string) (*types.Station, error) {
	var s types.Station
	err := c.db.QueryRow("SELECT id, name, url FROM stations WHERE name = ?", name).
		Scan(&s.ID, &s.Name, &s.URL)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) CreateStation(station types.Station) error {
	_, err := c.db.Exec("INSERT INTO stations (name, url) VALUES (?, ?)", station.Name, station.URL)
	return err
}

func (c *Client) UpdateStation(show *types.Station) error {
	_, err := c.db.Exec(`UPDATE shows SET name = ?, url = ? WHERE id = ?`,
		show.Name, show.URL, show.ID)
	return err
}

func (c *Client) DeleteStation(id uint) error {
	_, err := c.db.Exec("DELETE FROM stations WHERE id = ?", id)
	return err
}

func (c *Client) GetAllShows() ([]types.Show, error) {
	rows, err := c.db.Query("SELECT id, name, duration, day, hour, min, scheduled, station_id FROM shows")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shows []types.Show
	for rows.Next() {
		var s types.Show
		var duration int64
		if err := rows.Scan(&s.ID, &s.Name, &duration, &s.Day, &s.Hour, &s.Min, &s.Scheduled, &s.StationID); err != nil {
			return nil, err
		}
		s.Duration = types.ShowDuration{Duration: time.Duration(duration) * time.Second}
		shows = append(shows, s)
	}
	return shows, nil
}

func (c *Client) GetShowByName(name string) (*types.Show, error) {
	var s types.Show
	var duration int64
	err := c.db.QueryRow("SELECT id, name, duration, day, hour, min, scheduled, station_id FROM shows WHERE name = ?", name).
		Scan(&s.ID, &s.Name, &duration, &s.Day, &s.Hour, &s.Min, &s.Scheduled, &s.StationID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	s.Duration = types.ShowDuration{Duration: time.Duration(duration) * time.Second}

	return &s, nil
}

func (c *Client) GetShowByID(id uint) (*types.Show, error) {
	var s types.Show
	var duration int64
	err := c.db.QueryRow("SELECT id, name, duration, day, hour, min, scheduled, station_id FROM shows WHERE id = ?", id).
		Scan(&s.ID, &s.Name, &duration, &s.Day, &s.Hour, &s.Min, &s.Scheduled, &s.StationID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	s.Duration = types.ShowDuration{Duration: time.Duration(duration) * time.Second}

	return &s, nil
}

func (c *Client) CreateShow(show types.Show) error {
	_, err := c.db.Exec(`INSERT INTO shows (name, duration, day, hour, min, scheduled, station_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		show.Name, int64(show.Duration.Duration/time.Second), show.Day, show.Hour, show.Min, show.Scheduled, show.StationID)
	return err
}

func (c *Client) UpdateShow(show *types.Show) error {
	_, err := c.db.Exec(`UPDATE shows SET name = ?, duration = ?, day = ?, hour = ?, min = ?, scheduled = ?, station_id = ?
		WHERE id = ?`,
		show.Name, int64(show.Duration.Duration/time.Second), show.Day, show.Hour, show.Min, show.Scheduled, show.StationID, show.ID)
	return err
}

func (c *Client) DeleteShow(id uint) error {
	_, err := c.db.Exec("DELETE FROM shows WHERE id = ?", id)
	return err
}
