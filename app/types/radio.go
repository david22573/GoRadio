package types

type Station struct {
	ID   uint   `json:"id" form:"id"`
	Name string `json:"name" form:"name"`
	URL  string `json:"url" form:"url"`
}

type Show struct {
	ID        uint     `json:"id" form:"id"`
	Name      string   `json:"name" form:"name"`
	Day       uint     `json:"day" form:"day"`   // 0 = Sunday
	Hour      uint     `json:"hour" form:"hour"` // 0–23
	Min       uint     `json:"min" form:"min"`   // 0–59
	Duration  Duration `json:"duration" form:"duration"`
	Scheduled bool     `json:"scheduled" form:"scheduled"`
	StationID uint     `json:"station_id" form:"station_id"`
}

func (s Show) StartTime() (uint, uint, uint) {
	return s.Hour, s.Min, 0
}
