package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/david22573/GoRadio/app/types"

	_ "github.com/asg017/sqlite-vec-go-bindings/ncruces"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type Client struct {
	db *sql.DB
}

func NewSQLiteClient(filename string) (*Client, error) {
	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		return nil, fmt.Errorf("error creating data folder: %w", err)
	}

	dsn := fmt.Sprintf("data/%s", filename)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Performance Optimization: Connection Pooling
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return &Client{db: db}, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) GetDB() *sql.DB {
	return c.db
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

func (c *Client) GetTracksByStation(stationID uint) ([]types.Track, error) {
	return queryMultiple(c.db, scanTrack, "SELECT id, station_id, title, artist, url, duration, analyzed_at FROM tracks WHERE station_id = ?", stationID)
}

func (c *Client) GetTrackByID(id uint) (*types.Track, error) {
	return querySingle(c.db, scanTrack, "SELECT id, station_id, title, artist, url, duration, analyzed_at FROM tracks WHERE id = ?", id)
}

func (c *Client) CreateTrack(track types.Track) (uint, error) {
	id, err := execInsert(c.db, "INSERT INTO tracks (station_id, title, artist, url, duration, analyzed_at) VALUES (?, ?, ?, ?, ?, ?)",
		track.StationID, track.Title, track.Artist, track.URL, track.Duration, track.AnalyzedAt)
	return uint(id), err
}

func (c *Client) CreateTracksBatch(tracks []types.Track) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO tracks (station_id, title, artist, url, duration, analyzed_at) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range tracks {
		if _, err := stmt.Exec(t.StationID, t.Title, t.Artist, t.URL, t.Duration, t.AnalyzedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (c *Client) GetRandomTrack(excludeIDs []uint) (*types.Track, error) {
	query := "SELECT id, station_id, title, artist, url, duration, analyzed_at FROM tracks"
	var args []interface{}

	if len(excludeIDs) > 0 {
		query += " WHERE id NOT IN ("
		for i, id := range excludeIDs {
			if i > 0 {
				query += ","
			}
			query += "?"
			args = append(args, id)
		}
		query += ")"
	}

	query += " ORDER BY RANDOM() LIMIT 1"
	return querySingle(c.db, scanTrack, query, args...)
}

func (c *Client) SearchTracks(query string) ([]types.Track, error) {
	q := "%" + query + "%"
	sql := "SELECT id, station_id, title, artist, url, duration, analyzed_at FROM tracks WHERE title LIKE ? OR artist LIKE ? LIMIT 20"
	return queryMultiple(c.db, scanTrack, sql, q, q)
}
