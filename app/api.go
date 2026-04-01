package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/david22573/GoRadio/app/types"
	"github.com/gin-gonic/gin"
)

// SearchCache implementation
type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

type searchCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

func newSearchCache(ttl time.Duration) *searchCache {
	return &searchCache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

func (c *searchCache) get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, found := c.entries[key]
	if !found || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.data, true
}

func (c *searchCache) set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	}
}

type APIHandler struct {
	app   *App
	cache *searchCache
}

func (a *APIHandler) RegisterAPI() {
	// Initialize cache with a 15-minute TTL for search queries
	a.cache = newSearchCache(15 * time.Minute)

	api := a.app.Router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
		api.GET("/search", a.SearchStations) // New Route

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

func (a *APIHandler) SearchStations(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(400, gin.H{"error": "search query 'q' is required"})
		return
	}

	// 1. Check local Go cache
	if cachedData, found := a.cache.get(query); found {
		c.Data(200, "application/json", cachedData)
		return
	}

	// 2. Cache miss, fetch from upstream API
	upstreamURL := fmt.Sprintf("https://de1.api.radio-browser.info/json/stations/search?name=%s&limit=12&hidebroken=true", url.QueryEscape(query))
	resp, err := http.Get(upstreamURL)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to contact radio-browser api"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "upstream api returned an error"})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to read upstream response"})
		return
	}

	// 3. Verify it's valid JSON before caching
	if !json.Valid(body) {
		c.JSON(500, gin.H{"error": "upstream api returned invalid json"})
		return
	}

	// 4. Save to cache and return
	a.cache.set(query, body)
	c.Data(200, "application/json", body)
}

func (a *APIHandler) GetStations(c *gin.Context) {
	stations, err := a.app.DB.GetAllStations()
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
	if err := a.app.DB.CreateStation(station); err != nil {
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
	if err := a.app.DB.UpdateStation(&station); err != nil {
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
	if err := a.app.DB.DeleteStation(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Station deleted successfully"})
}

func (a *APIHandler) GetShows(c *gin.Context) {
	shows, err := a.app.DB.GetAllShows()
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
	if err := a.app.DB.CreateShow(show); err != nil {
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
	if err := a.app.DB.UpdateShow(&show); err != nil {
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
	if err := a.app.DB.DeleteShow(id); err != nil {
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
