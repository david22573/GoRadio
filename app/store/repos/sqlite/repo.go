package sqlite

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"

	"github.com/david22573/GoRadio/app/store"
)

type sqliteRepo struct {
	db *sql.DB
}

// NewSQLiteRepo opens (and migrates) the database and returns a RadioRepository.
func NewSQLiteRepo(filename string) (store.RadioRepository, error) {
	_ = os.MkdirAll("data", os.ModePerm)
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
	return &sqliteRepo{db: db}, nil
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
        name       TEXT NOT NULL,
        duration   INTEGER NOT NULL,  -- stored as seconds
        day        INTEGER NOT NULL,  -- time.Weekday (0–6)
        hour       INTEGER NOT NULL,
        min        INTEGER NOT NULL,
        station_id INTEGER NOT NULL,
        FOREIGN KEY(station_id) REFERENCES stations(id) ON DELETE CASCADE
    );
    `
	_, err := db.Exec(schema)
	return err
}
