package db

import "github.com/david22573/GoRadio/app/types"

// Client defines all the methods the store layer can use to interact with the DB
type Client interface {
	GetAllStations() ([]types.Station, error)
	GetStationByID(id uint) (*types.Station, error)
	GetStationByName(name string) (*types.Station, error)
	CreateStation(station types.Station) error
	UpdateStation(station *types.Station) error
	DeleteStation(id uint) error

	GetAllShows() ([]types.Show, error)
	GetShowByID(id uint) (*types.Show, error)
	GetShowByName(name string) (*types.Show, error)
	CreateShow(show types.Show) error
	UpdateShow(show *types.Show) error
	DeleteShow(id uint) error

	// Tracks
	GetTrackByID(id uint) (*types.Track, error)
	GetTracksByStation(stationID uint) ([]types.Track, error)
	CreateTrack(track types.Track) (uint, error)
	CreateTracksBatch(tracks []types.Track) error
	GetRandomTrack(excludeIDs []uint) (*types.Track, error)

	// Vectors
	InsertTrackVector(trackID uint, embedding []float64) error
	GetTrackVector(trackID uint) ([]float64, error)
	FindNearestNeighbors(trackID uint, k int, excludeIDs []uint) ([]types.Track, []float64, error)
	FindDistantNeighbors(trackID uint, minDistance, maxDistance float64, k int) ([]types.Track, error)

	// Sessions
	CreateSession(session *types.SessionState) error
	GetSession(id string) (*types.SessionState, error)
	UpdateSession(session *types.SessionState) error
	DeleteSession(id string) error

	// Session events
	RecordPlayEvent(sessionID string, event types.PlayEvent) error
	RecordSkipEvent(sessionID string, event types.SkipEvent) error
	GetSessionEvents(sessionID string) ([]types.PlayEvent, []types.SkipEvent, error)

	// Close releases any open resources (like DB connections)
	Close() error
}
