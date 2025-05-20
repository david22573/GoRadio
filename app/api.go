package app

import (
	"github.com/david22573/GoRadio/app/types"
	"github.com/gin-gonic/gin"
)

type API struct {
	app *App
}

func (a *API) RegisterAPI() {
	api := a.app.Router.Group("/api")
	{
		api.GET("/stations", a.GetStations)
		api.POST("/stations", a.CreateStation)
		api.PUT("/stations/:id", a.CreateStation)
		api.DELETE("/stations/:id", nil)
		api.GET("/shows", a.GetShows)
	}
}

func (a *API) GetStations(c *gin.Context) {
	stations, err := a.app.Repo.GetAllStations()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"stations": stations})
}

func (a *API) GetShows(c *gin.Context) {
	shows, err := a.app.Repo.GetAllShows()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"shows": shows})
}

func (a *API) CreateStation(c *gin.Context) {

	name := c.PostForm("name")
	url := c.PostForm("url")

	if name == "" || url == "" {
		c.JSON(400, gin.H{"error": "name and url are required"})
		return
	}

	station := types.Station{Name: name, URL: url}
	if err := a.app.Repo.CreateStation(&station); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}
