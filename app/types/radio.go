package types

type Station struct {
	ID   uint   `json:"id" form:"id"`
	Name string `json:"name" form:"name"`
	URL  string `json:"url" form:"url"`
}
