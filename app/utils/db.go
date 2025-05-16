package utils

import (
	"database/sql"
	"errors"

	"github.com/david22573/GoRadio/app/store"
)

// queryMultiple executes a query expected to return multiple rows.
func QueryMultiple[T any](db *sql.DB, scannerFunc func(*sql.Rows) (T, error), query string, args ...any) ([]T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []T
	for rows.Next() {
		item, err := scannerFunc(rows)
		if err != nil {
			return nil, err // Error during scan
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err // Error during iteration
	}
	return items, nil
}

// querySingle executes a query expected to return a single row.
// Returns store.ErrNotFound if no row is found.
func QuerySingle[T any](db *sql.DB, scannerFunc func(*sql.Rows) (T, error), query string, args ...any) (*T, error) {
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
		if rows.Next() { // Should not happen for a query expected to return a single row
			return nil, errors.New("database: expected single row, but got multiple")
		}
		return &item, nil
	}
	if err = rows.Err(); err != nil { // Check for errors encountered during iteration
		return nil, err
	}
	return nil, store.ErrNotFound // No rows found
}

// execInsert handles INSERT statements and returns the last inserted ID.
func ExecInsert(db *sql.DB, query string, args ...any) (int64, error) {
	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// execAffected handles UPDATE or DELETE statements and returns rows affected.
func ExecAffected(db *sql.DB, query string, args ...any) (int64, error) {
	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
