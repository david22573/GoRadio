package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

// RadioBrowserBaseURL can be overridden in tests
var RadioBrowserBaseURL string

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

func (c *searchCache) StartCleanup() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			c.mu.Lock()
			now := time.Now()
			for k, v := range c.entries {
				if now.After(v.expiresAt) {
					delete(c.entries, k)
				}
			}
			c.mu.Unlock()
		}
	}()
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

	// Enforce hard capacity limit to prevent OOM. Evict a single arbitrary
	// entry instead of clearing the whole map (which would cause a cache
	// stampede under load).
	if len(c.entries) >= 1000 {
		if _, exists := c.entries[key]; !exists {
			for k := range c.entries {
				delete(c.entries, k)
				break
			}
		}
	}

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
	a.cache.StartCleanup()

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

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "12")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 12
	}
	if limit > 48 {
		limit = 48
	}

	offset := (page - 1) * limit
	cacheKey := fmt.Sprintf("search:stations:%s:p%d:l%d", query, page, limit)

	// 1. Check local Go cache
	if cachedData, found := a.cache.get(cacheKey); found {
		c.Data(200, "application/json", cachedData)
		return
	}

	// 2. Cache miss, fetch from upstream API
	baseURL := RadioBrowserBaseURL
	if baseURL == "" {
		baseURL = "https://de1.api.radio-browser.info"
	}
	upstreamURL := fmt.Sprintf("%s/json/stations/search?name=%s&limit=%d&offset=%d&hidebroken=true", baseURL, url.QueryEscape(query), limit, offset)
	
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", upstreamURL, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create request"})
		return
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
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

	var results []json.RawMessage
	if err := json.Unmarshal(body, &results); err != nil {
		c.JSON(500, gin.H{"error": "failed to parse upstream json"})
		return
	}

	hasMore := len(results) == limit

	responseObj := gin.H{
		"page":     page,
		"limit":    limit,
		"has_more": hasMore,
		"results":  results,
	}

	respBody, err := json.Marshal(responseObj)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to encode response"})
		return
	}

	// 4. Save to cache and return
	a.cache.set(cacheKey, respBody)
	c.Data(200, "application/json", respBody)
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
					id := strings.TrimSpace(parts[1])
					if id == "" {
						continue
					}
					tracks = append(tracks, types.Track{
						Title:  strings.TrimSpace(parts[0]),
						Artist: strings.TrimSpace(parts[3]),
						URL:    "https://www.youtube.com/watch?v=" + id,
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

	// 2. Validate the URL host before handing it to yt-dlp (SSRF guard).
	if !isAllowedResolveURL(trackURL) {
		c.JSON(400, gin.H{"error": "url host not allowed"})
		return
	}

	// 3. Use yt-dlp to get the direct audio stream URL
	// -g: get URL
	// -f: bestaudio
	cmd := exec.CommandContext(c.Request.Context(), "yt-dlp", "-g", "-f", "bestaudio", trackURL)
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
		SeedTrackID uint         `json:"seed_track_id"`
		SeedTrack   *types.Track `json:"seed_track"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	seedTrackID := req.SeedTrackID

	// A seed_track object means the caller is handing us track metadata (e.g. a
	// YouTube search result whose id is 0). Ingest it as a real tracks row and
	// generate its vector so it can be seeded and found by the queue.
	if req.SeedTrack != nil {
		id, err := a.ingestSeedTrack(c.Request.Context(), req.SeedTrack)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to ingest seed track: " + err.Error()})
			return
		}
		seedTrackID = id
	} else if seedTrackID == 0 {
		c.JSON(400, gin.H{"error": "seed_track_id or seed_track is required"})
		return
	} else {
		// Legacy path: verify the referenced track exists.
		if _, err := a.app.DB.GetTrackByID(seedTrackID); err != nil {
			c.JSON(404, gin.H{"error": "seed track not found"})
			return
		}
	}

	s, err := a.sessionMgr.Create(seedTrackID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, s)
}

// seedEmbeddingDim is the fixed dimension of track embeddings (matches the
// track_vectors vec0 schema) used for the fallback vector.
const seedEmbeddingDim = 128

// ingestSeedTrack persists the given track metadata as a real tracks row (reusing
// an existing row with the same URL), generates its embedding best-effort, and
// returns the track ID. Analysis failure never blocks session creation: a zero
// vector is stored as a fallback so the session manager can always find a seed
// vector.
func (a *APIHandler) ingestSeedTrack(ctx context.Context, t *types.Track) (uint, error) {
	if t == nil || (t.URL == "" && t.Title == "") {
		return 0, fmt.Errorf("seed track metadata is empty")
	}

	// Reuse an existing row with the same URL so repeated plays of the same
	// search result don't create duplicate tracks.
	if t.URL != "" {
		if existing, err := a.app.DB.GetTrackByURL(t.URL); err == nil && existing != nil {
			// Only re-analyze when the existing track has no vector yet.
			if _, err := a.app.DB.GetTrackVector(existing.ID); err != nil {
				a.ensureTrackVector(ctx, existing.ID, existing.URL)
			}
			return existing.ID, nil
		}
	}

	id, err := a.app.DB.CreateTrack(types.Track{
		Title:  t.Title,
		Artist: t.Artist,
		URL:    t.URL,
	})
	if err != nil {
		return 0, err
	}

	a.ensureTrackVector(ctx, id, t.URL)
	return id, nil
}

// ensureTrackVector generates and stores an embedding for trackID, falling back
// to a zero vector when audio tooling is unavailable, so session creation never
// hard-fails because ffmpeg/yt-dlp are missing.
func (a *APIHandler) ensureTrackVector(ctx context.Context, trackID uint, url string) {
	if url == "" {
		a.insertFallbackVector(trackID, "no URL to analyze")
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	features, err := a.app.AudioAnalyzer.AnalyzeURL(ctx, url, 30)
	if err != nil {
		log.Printf("warning: failed to analyze seed track %d: %v", trackID, err)
		a.insertFallbackVector(trackID, "")
		return
	}
	if err := a.app.DB.InsertTrackVector(trackID, features.Embedding); err != nil {
		log.Printf("warning: failed to store vector for seed track %d: %v", trackID, err)
	}
}

// insertFallbackVector stores a zero vector so GetVectorByID succeeds even when
// audio analysis is unavailable.
func (a *APIHandler) insertFallbackVector(trackID uint, reason string) {
	if reason != "" {
		log.Printf("warning: seed track %d: %s; using fallback vector", trackID, reason)
	}
	if err := a.app.DB.InsertTrackVector(trackID, make([]float64, seedEmbeddingDim)); err != nil {
		log.Printf("warning: failed to store fallback vector for seed track %d: %v", trackID, err)
	}
}

func (a *APIHandler) GetSession(c *gin.Context) {
	id := c.Param("id")
	s, err := a.sessionMgr.Snapshot(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "session not found"})
		return
	}
	c.JSON(200, s)
}

// Queue Handlers
func (a *APIHandler) GetQueue(c *gin.Context) {
	sessionID := c.Param("sessionId")
	q := a.queueMgr.Snapshot(sessionID)
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
	q := a.queueMgr.Snapshot(sessionID)
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
	s, err := a.sessionMgr.Snapshot(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "session not found"})
		return
	}
	c.JSON(200, s.CalculateMetrics())
}

func (a *APIHandler) GetSessionJourney(c *gin.Context) {
	id := c.Param("id")
	s, err := a.sessionMgr.Snapshot(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "session not found"})
		return
	}
	c.JSON(200, gin.H{"journey": s.GetJourney()})
}

func getUintParam(c *gin.Context, param string) (uint, error) {
	idStr := c.Param(param)
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	if idInt <= 0 {
		return 0, fmt.Errorf("invalid %s: must be a positive integer", param)
	}
	return uint(idInt), nil
}

// allowedResolveHosts is the allowlist of hosts that ResolveTrack will pass to
// yt-dlp. This prevents the resolve endpoint from being used as an SSRF oracle.
var allowedResolveHosts = map[string]bool{
	"youtube.com":        true,
	"www.youtube.com":    true,
	"youtu.be":           true,
	"soundcloud.com":     true,
	"www.soundcloud.com": true,
}

// isAllowedResolveURL reports whether raw is an http(s) URL whose host is an
// allowed resolve host or one of its subdomains.
func isAllowedResolveURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if allowedResolveHosts[host] {
		return true
	}
	for allowed := range allowedResolveHosts {
		if strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}
	return false
}
