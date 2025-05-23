package store

import (
	"github.com/david22573/GoRadio/app/db"
	"github.com/david22573/GoRadio/app/types"
)

type radioStore struct {
	dbClient db.Client
}

func NewRadioStore(client db.Client) *radioStore {
	return &radioStore{dbClient: client}
}

func (s *radioStore) GetAllStations() ([]types.Station, error) {
	return s.dbClient.GetAllStations()
}

func (s *radioStore) GetStationByID(id uint) (*types.Station, error) {
	return s.dbClient.GetStationByID(id)
}

func (s *radioStore) GetStationByName(name string) (*types.Station, error) {
	return s.dbClient.GetStationByName(name)
}

func (s *radioStore) CreateStation(station types.Station) error {
	return s.dbClient.CreateStation(station)
}

func (s *radioStore) UpdateStation(station *types.Station) error {
	return s.dbClient.UpdateStation(station)
}

func (s *radioStore) DeleteStation(id uint) error {
	return s.dbClient.DeleteStation(id)
}

func (s *radioStore) GetAllShows() ([]types.Show, error) {
	return s.dbClient.GetAllShows()
}

func (s *radioStore) GetShowByID(id uint) (*types.Show, error) {
	return s.dbClient.GetShowByID(id)
}
func (s *radioStore) GetShowByName(name string) (*types.Show, error) {
	return s.dbClient.GetShowByName(name)
}

func (s *radioStore) CreateShow(show types.Show) error {
	return s.dbClient.CreateShow(show)
}

func (s *radioStore) UpdateShow(show *types.Show) error {
	return s.dbClient.UpdateShow(show)
}

func (s *radioStore) DeleteShow(id uint) error {
	return s.dbClient.DeleteShow(id)
}

func (s *radioStore) Close() error {
	return s.dbClient.Close()
}
