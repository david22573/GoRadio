package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/david22573/GoRadio/app/store"
	"github.com/david22573/GoRadio/app/types"
)

// QueryMultiple executes a query expected to return multiple rows.
func queryMultiple[T any](
	db *sql.DB,
	scannerFunc func(*sql.Rows) (T, error),
	query string, args ...any,
) ([]T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []T
	for rows.Next() {
		item, err := scannerFunc(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// querySingle executes a query expected to return a single row.
// Returns store.ErrNotFound if no row is found.
func querySingle[T any](
	db *sql.DB,
	scannerFunc func(*sql.Rows) (T, error),
	query string, args ...any,
) (*T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		item, err := scannerFunc(rows)
		if err != nil {
			return nil, err
		}
		if rows.Next() {
			return nil, errors.New("database: expected single row, but got multiple")
		}
		return &item, nil
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return nil, store.ErrNotFound
}

// execInsert handles INSERT statements and returns the last inserted ID.
func execInsert(db *sql.DB, query string, args ...any) (int64, error) {
	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// execAffected handles UPDATE or DELETE statements and returns rows affected.
func execAffected(db *sql.DB, query string, args ...any) (int64, error) {
	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// --- Scanner Functions ---

// scanShow scans a show from the current sql.Rows.
// Assumes columns are: id, name, duration, day, hour, min, station_id
func scanShow(rows *sql.Rows) (types.Show, error) {
	var s types.Show
	var secs int64
	var weekday int
	var hour, min uint

	err := rows.Scan(
		&s.ID,
		&s.Name,
		&secs,
		&weekday,
		&hour,
		&min,
		&s.Scheduled,
		&s.StationID,
	)
	if err != nil {
		return types.Show{}, err
	}
	s.Duration = time.Duration(secs) * time.Second
	s.ShowSchedule = types.ShowSchedule{
		Day:  time.Weekday(weekday),
		Hour: hour,
		Min:  min,
	}
	return s, nil
}

// scanStation scans a station from the current sql.Rows.
// Assumes columns are: id, name, url
func scanStation(rows *sql.Rows) (types.Station, error) {
	var st types.Station
	if err := rows.Scan(&st.ID, &st.Name, &st.URL); err != nil {
		return types.Station{}, err
	}
	return st, nil
}
