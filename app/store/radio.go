// File: app/store/radio.go
package store

import (
	"errors" // For creating custom errors

	"github.com/david22573/GoRadio/app/types"
)

// ErrNotFound is an error returned when a requested item is not found in the store.
var ErrNotFound = errors.New("requested item not found")

// RadioRepository defines the interface for radio data storage operations.
type RadioRepository interface {
	GetAllStations() ([]types.Station, error)
	GetStationByID(id uint) (*types.Station, error)
	GetStationByName(name string) (*types.Station, error)
	CreateStation(station *types.Station) error
	UpdateStation(station *types.Station) error
	DeleteStation(id uint) error

	GetAllShows() ([]types.Show, error)
	GetAllShowsByStation(stationID uint) ([]types.Show, error)
	GetShowByID(id uint) (*types.Show, error)
	CreateShow(show *types.Show) error
	UpdateShow(show *types.Show) error
	DeleteShow(id uint) error

	Close() error
}

// RadioStore provides methods to interact with radio data, using a repository.
type RadioStore struct {
	repo RadioRepository
}

// NewRadioStore creates a new RadioStore.
func NewRadioStore(repo RadioRepository) *RadioStore {
	return &RadioStore{repo: repo}
}

// GetAllStations retrieves all stations.
func (s *RadioStore) GetAllStations() ([]types.Station, error) {
	return s.repo.GetAllStations()
}

// GetStationByID retrieves a station by its ID.
func (s *RadioStore) GetStationByID(id uint) (*types.Station, error) {
	return s.repo.GetStationByID(id)
}

func (s *RadioStore) GetStationByName(name string) (*types.Station, error) {
	return s.repo.GetStationByName(name)
}

// CreateStation creates a new station.
func (s *RadioStore) CreateStation(station *types.Station) error {
	return s.repo.CreateStation(station)
}

func (s *RadioStore) UpdateStation(station *types.Station) error {
	return s.repo.UpdateStation(station)
}

// DeleteStation deletes a station by its ID.
func (s *RadioStore) DeleteStation(id uint) error {
	return s.repo.DeleteStation(id)
}

// GetAllShows retrieves all shows.
func (s *RadioStore) GetAllShows() ([]types.Show, error) {
	return s.repo.GetAllShows()
}

// GetAllShowsByStation retrieves all shows for a given station ID.
func (s *RadioStore) GetAllShowsByStation(stationID uint) ([]types.Show, error) {
	return s.repo.GetAllShowsByStation(stationID)
}

// GetShowByID retrieves a show by its ID.
func (s *RadioStore) GetShowByID(id uint) (*types.Show, error) {
	return s.repo.GetShowByID(id)
}

// CreateShow creates a new show.
func (s *RadioStore) CreateShow(show *types.Show) error {
	return s.repo.CreateShow(show)
}

// UpdateShow updates an existing show.
func (s *RadioStore) UpdateShow(show *types.Show) error {
	return s.repo.UpdateShow(show)
}

// DeleteShow deletes a show by its ID.
func (s *RadioStore) DeleteShow(id uint) error {
	return s.repo.DeleteShow(id)
}
