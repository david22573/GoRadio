package app

import (
	"strconv"

	"github.com/david22573/GoRadio/app/types"
	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	app *App
}

func (a *APIHandler) RegisterAPI() {
	api := a.app.Router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
		api.GET("/stations", a.GetStations)
		api.POST("/stations", a.CreateStation)
		api.PUT("/stations/:id", a.UpdateStation)
		api.DELETE("/stations/:id", a.DeleteStation)
		api.GET("/shows", a.GetShows)
		api.POST("/shows", a.CreateShow)
		api.PUT("/shows/:id", a.UpdateShow)
		api.DELETE("/shows/:id", a.DeleteShow)
	}
}

func (a *APIHandler) GetStations(c *gin.Context) {
	stations, err := a.app.Store.GetAllStations()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"stations": stations})
}

func (a *APIHandler) CreateStation(c *gin.Context) {
	var station types.Station
	if err := c.Bind(&station); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := a.app.Store.CreateStation(station); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"message": "Station created successfully"})
}

func (a *APIHandler) UpdateStation(c *gin.Context) {
	var station types.Station
	if err := c.Bind(&station); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	station.ID = id
	if err := a.app.Store.UpdateStation(&station); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Station updated successfully"})
}

func (a *APIHandler) DeleteStation(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := a.app.Store.DeleteStation(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Station deleted successfully"})
}

func (a *APIHandler) GetShows(c *gin.Context) {
	shows, err := a.app.Store.GetAllShows()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"shows": shows})
}

func (a *APIHandler) CreateShow(c *gin.Context) {
	var show types.Show
	if err := c.Bind(&show); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := a.app.Store.CreateShow(show); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"message": "Show created successfully"})
}

func (a *APIHandler) UpdateShow(c *gin.Context) {
	var show types.Show
	if err := c.Bind(&show); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	show.ID = id
	if err := a.app.Store.UpdateShow(&show); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Show updated successfully"})
}

func (a *APIHandler) DeleteShow(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := a.app.Store.DeleteShow(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Show deleted successfully"})
}

func getUintParam(c *gin.Context, param string) (uint, error) {
	idStr := c.Param(param)
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return uint(idInt), nil
}
