package sqlite

import (
	"database/sql"
	"time"

	"github.com/david22573/GoRadio/app/types"
)

// --- Scanner Functions ---

// scanShowInternal scans a show from the current sql.Rows.
// Assumes columns are: id, name, duration, day, hour, min, station_id
func scanShowInternal(rows *sql.Rows) (types.Show, error) {
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
