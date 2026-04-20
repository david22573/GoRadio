package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/david22573/GoRadio/app/queue"
	"github.com/david22573/GoRadio/app/session"
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
	app        *App
	cache      *searchCache
	sessionMgr *session.Manager
	queueMgr   *queue.Manager
}

func (a *APIHandler) RegisterAPI() {
	// Initialize cache with a 15-minute TTL for search queries
	a.cache = newSearchCache(15 * time.Minute)

	// Wire managers from app
	a.sessionMgr = a.app.SessionMgr
	a.queueMgr = a.app.QueueMgr

	api := a.app.Router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
		api.GET("/search", a.SearchStations)
		api.GET("/tracks/search", a.SearchTracks)
		api.GET("/tracks/resolve", a.ResolveTrack)

		api.GET("/stations", a.GetStations)
		api.POST("/stations", a.CreateStation)
		api.PUT("/stations/:id", a.UpdateStation)
		api.DELETE("/stations/:id", a.DeleteStation)

		api.GET("/shows", a.GetShows)
		api.POST("/shows", a.CreateShow)
		api.PUT("/shows/:id", a.UpdateShow)
		api.DELETE("/shows/:id", a.DeleteShow)

		// Session management
		api.POST("/sessions", a.CreateSession)
		api.GET("/sessions/:id", a.GetSession)

		// Queue management
		api.GET("/queue/:sessionId", a.GetQueue)
		api.POST("/queue/:sessionId/advance", a.AdvanceQueue)
		api.GET("/queue/:sessionId/upcoming", a.GetUpcoming)

		// Events
		api.POST("/events/play", a.RecordPlayEvent)
		api.POST("/events/skip", a.RecordSkipEvent)

		// Analytics
		api.GET("/sessions/:id/metrics", a.GetSessionMetrics)
		api.GET("/sessions/:id/journey", a.GetSessionJourney)
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

func (a *APIHandler) SearchTracks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(400, gin.H{"error": "search query 'q' is required"})
		return
	}

	// 1. Check local Go cache for global results
	cacheKey := "tracks:" + query
	if cachedData, found := a.cache.get(cacheKey); found {
		var cachedTracks []types.Track
		if err := json.Unmarshal(cachedData, &cachedTracks); err == nil {
			c.JSON(200, gin.H{"tracks": cachedTracks})
			return
		}
	}

	// 2. Query Local DB
	tracks, _ := a.app.DB.SearchTracks(query)

	// 3. If local results are low, bridge to global YouTube (via yt-dlp)
	if len(tracks) < 5 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Search YouTube: ytsearch10:query
		// Output format: title || id || duration || uploader
		cmd := exec.CommandContext(ctx, "yt-dlp", 
			"ytsearch10:"+query, 
			"--print", "%(title)s || %(id)s || %(duration)s || %(uploader)s",
			"--no-playlist")
		
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				parts := strings.Split(line, " || ")
				if len(parts) >= 4 {
					tracks = append(tracks, types.Track{
						Title:  strings.TrimSpace(parts[0]),
						Artist: strings.TrimSpace(parts[3]),
						URL:    "https://www.youtube.com/watch?v=" + strings.TrimSpace(parts[1]),
					})
				}
			}
		}
	}

	// 4. Update cache if we found anything
	if len(tracks) > 0 {
		if data, err := json.Marshal(tracks); err == nil {
			a.cache.set(cacheKey, data)
		}
	}

	c.JSON(200, gin.H{"tracks": tracks})
}

func (a *APIHandler) ResolveTrack(c *gin.Context) {
	trackURL := c.Query("url")
	if trackURL == "" {
		c.JSON(400, gin.H{"error": "url parameter is required"})
		return
	}

	// 1. Check if it needs resolution (YouTube etc)
	needsResolve := strings.Contains(trackURL, "youtube.com") ||
		strings.Contains(trackURL, "youtu.be") ||
		strings.Contains(trackURL, "soundcloud.com")

	if !needsResolve {
		c.JSON(200, gin.H{"url": trackURL})
		return
	}

	// 2. Use yt-dlp to get the direct audio stream URL
	// -g: get URL
	// -f: bestaudio
	cmd := exec.Command("yt-dlp", "-g", "-f", "bestaudio", trackURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to resolve stream URL", "details": string(output)})
		return
	}

	resolvedURL := strings.TrimSpace(string(output))
	c.JSON(200, gin.H{"url": resolvedURL})
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

// Session Handlers
func (a *APIHandler) CreateSession(c *gin.Context) {
	var req struct {
		SeedTrackID uint `json:"seed_track_id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Validation: Verify track exists
	_, err := a.app.DB.GetTrackByID(req.SeedTrackID)
	if err != nil {
		c.JSON(404, gin.H{"error": "seed track not found"})
		return
	}

	s, err := a.sessionMgr.Create(req.SeedTrackID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, s)
}

func (a *APIHandler) GetSession(c *gin.Context) {
	id := c.Param("id")
	s, err := a.sessionMgr.GetSession(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "session not found"})
		return
	}
	c.JSON(200, s)
}

// Queue Handlers
func (a *APIHandler) GetQueue(c *gin.Context) {
	sessionID := c.Param("sessionId")
	q, err := a.queueMgr.GetQueue(sessionID)
	if err != nil {
		c.JSON(404, gin.H{"error": "queue not found"})
		return
	}
	c.JSON(200, q)
}

func (a *APIHandler) AdvanceQueue(c *gin.Context) {
	sessionID := c.Param("sessionId")

	nextTrack, mode, err := a.queueMgr.Advance(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"track": nextTrack,
		"mode":  mode, // "exploitation" or "exploration"
	})
}

func (a *APIHandler) GetUpcoming(c *gin.Context) {
	sessionID := c.Param("sessionId")
	q, err := a.queueMgr.GetQueue(sessionID)
	if err != nil {
		c.JSON(404, gin.H{"error": "queue not found"})
		return
	}
	c.JSON(200, gin.H{"upcoming": q.Upcoming})
}

// Event Handlers
func (a *APIHandler) RecordPlayEvent(c *gin.Context) {
	var req struct {
		SessionID  string  `json:"session_id"`
		TrackID    uint    `json:"track_id"`
		Completion float64 `json:"completion"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := a.sessionMgr.RecordPlay(req.SessionID, req.TrackID, req.Completion); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "recorded"})
}

func (a *APIHandler) RecordSkipEvent(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id"`
		TrackID   uint   `json:"track_id"`
		PlayedFor int    `json:"played_for"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := a.sessionMgr.RecordSkip(req.SessionID, req.TrackID, req.PlayedFor); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "recorded"})
}

func (a *APIHandler) GetSessionMetrics(c *gin.Context) {
	id := c.Param("id")
	metrics, err := a.sessionMgr.GetMetrics(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "session not found"})
		return
	}
	c.JSON(200, metrics)
}

func (a *APIHandler) GetSessionJourney(c *gin.Context) {
	id := c.Param("id")
	journey, err := a.sessionMgr.GetJourney(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "session not found"})
		return
	}
	c.JSON(200, gin.H{"journey": journey})
}

func getUintParam(c *gin.Context, param string) (uint, error) {
	idStr := c.Param(param)
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return uint(idInt), nil
}
