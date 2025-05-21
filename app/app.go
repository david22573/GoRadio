package app

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/david22573/GoRadio/app/store"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
)

type assetMap map[string]struct{}

//go:embed static/*
var rawStaticFS embed.FS

// App holds application state, file map, and scheduled jobs.
type App struct {
	Router *gin.Engine
	Repo   store.RadioRepository

	assetMap   assetMap
	schedulers []gocron.Scheduler
	fs         static.ServeFileSystem

	mu sync.Mutex
}

// NewApp constructs an App, embedding static assets and setting up storage.
// It returns an error instead of calling log.Fatal.
func NewApp(repo store.RadioRepository) (*App, error) {
	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		return nil, fmt.Errorf("create data folder: %w", err)
	}

	staticFS, err := static.EmbedFolder(rawStaticFS, "static")
	if err != nil {
		return nil, fmt.Errorf("setup assets: %w", err)
	}

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	return &App{
		Router: router,
		Repo:   repo,

		schedulers: []gocron.Scheduler{},
		fs:         staticFS,
	}, nil
}

// registerStatic sets up embedded file serving with a fast map-based lookup.
func (a *App) registerStatic() {
	a.Router.Use(static.Serve("/", a.fs))
	a.Router.NoRoute(func(c *gin.Context) {
		fmt.Printf("%s doesn't exists, redirect on /\n", c.Request.URL.Path)
		c.Redirect(http.StatusMovedPermanently, "/")
	})

}

// AddScheduler registers a new pointer-based scheduler.
func (a *App) AddScheduler(sch gocron.Scheduler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.schedulers = append(a.schedulers, sch)
}

// StartAll starts all schedulers asynchronously.
func (a *App) StartAll() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, sch := range a.schedulers {
		sch.Start()
	}
}

// StopAll stops all schedulers gracefully.
func (a *App) StopAll() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, sch := range a.schedulers {
		sch.Shutdown()
	}
}

// Run starts cron jobs, the HTTP server, and gracefully handles shutdown.
func (a *App) Run(addr string) error {
	// Register static assets and routes
	a.Router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	a.registerStatic()
	// Start all cron jobs before serving
	a.StartAll()

	// Create and start HTTP server
	srv := &http.Server{
		Addr:    addr,
		Handler: a.Router,
	}
	go func() {
		log.Printf("🚀 Listening on %s...", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt or terminate signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	// Shutdown HTTP first
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP shutdown: %w", err)
	}

	// Then stop all cron jobs
	a.StopAll()
	log.Println("Server gracefully stopped")
	return nil
}

func setupAssets(staticFs embed.FS) (assetMap, error) {
	staticFS, err := fs.Sub(staticFs, "static")
	if err != nil {
		return nil, fmt.Errorf("embed static fs: %w", err)
	}

	// Preload file paths for O(1) lookup
	fileMap := make(assetMap)
	fs.WalkDir(staticFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			fileMap["/"+path] = struct{}{}
		}
		return err
	})
	return fileMap, nil
}
