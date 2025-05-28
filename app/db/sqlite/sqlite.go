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

func (c *Client) GetAllStations() ([]types.Station, error) {
	return queryMultiple(c.db, scanStation, "SELECT id, name, url FROM stations")
}

func (c *Client) GetStationByID(id uint) (*types.Station, error) {
	return querySingle(c.db, scanStation, "SELECT id, name, url FROM stations WHERE id = ?", id)
}

func (c *Client) GetStationByName(name string) (*types.Station, error) {
	return querySingle(c.db, scanStation, "SELECT id, name, url FROM stations WHERE name = ?", name)
}

func (c *Client) CreateStation(station types.Station) error {
	_, err := execInsert(c.db, "INSERT INTO stations (name, url) VALUES (?, ?)", station.Name, station.URL)
	return err
}

func (c *Client) UpdateStation(station *types.Station) error {
	_, err := execAffected(c.db, "UPDATE stations SET name = ?, url = ? WHERE id = ?", station.Name, station.URL, station.ID)
	return err
}

func (c *Client) DeleteStation(id uint) error {
	_, err := execAffected(c.db, "DELETE FROM stations WHERE id = ?", id)
	return err
}

func (c *Client) GetAllShows() ([]types.Show, error) {
	return queryMultiple(c.db, scanShow, "SELECT id, name, duration, day, hour, min, scheduled, station_id FROM shows")
}

func (c *Client) GetShowByName(name string) (*types.Show, error) {
	return querySingle(c.db, scanShow, "SELECT id, name, duration, day, hour, min, scheduled, station_id FROM shows WHERE name = ?", name)
}

func (c *Client) GetShowByID(id uint) (*types.Show, error) {
	return querySingle(c.db, scanShow, "SELECT id, name, duration, day, hour, min, scheduled, station_id FROM shows WHERE id = ?", id)
}

func (c *Client) CreateShow(show types.Show) error {
	_, err := execInsert(c.db, `INSERT INTO shows (name, duration, day, hour, min, scheduled, station_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		show.Name, int64(show.Duration.Duration/time.Second), show.Day, show.Hour, show.Min, show.Scheduled, show.StationID)
	return err
}

func (c *Client) UpdateShow(show *types.Show) error {
	_, err := execAffected(c.db, `UPDATE shows SET name = ?, duration = ?, day = ?, hour = ?, min = ?, scheduled = ?, station_id = ?
		WHERE id = ?`,
		show.Name, int64(show.Duration.Duration/time.Second), show.Day, show.Hour, show.Min, show.Scheduled, show.StationID, show.ID)
	return err
}

func (c *Client) DeleteShow(id uint) error {
	_, err := execAffected(c.db, "DELETE FROM shows WHERE id = ?", id)
	return err
}
