package app

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/david22573/GoRadio/app/store"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
)

//go:embed static/*
var rawStaticFS embed.FS

// App holds application state and scheduled jobs
type App struct {
	Router     *gin.Engine
	Repo       store.RadioRepository
	schedulers []gocron.Scheduler
	mu         sync.Mutex
	fs         fs.FS
}

// NewApp constructs an App, embedding static assets and setting up storage
func NewApp(repo store.RadioRepository) *App {
	ensureDataFolder()

	// Prepare embedded FS rooted at "static"
	staticFS, err := fs.Sub(rawStaticFS, "static")
	if err != nil {
		log.Fatalf("failed to sub static FS: %v", err)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	return &App{
		Router:     router,
		Repo:       repo,
		schedulers: make([]gocron.Scheduler, 0),
		fs:         staticFS,
	}
}

// Run starts the HTTP server and listens for shutdown signals
func (a *App) Run(addr string) {
	// Serve embedded static files
	a.Router.StaticFS("/", http.FS(a.fs)) // mounts all files from embedded FS at /

	// SPA fallback: if no file matches, serve index.html
	a.Router.NoRoute(func(c *gin.Context) {
		// Try to serve the requested file
		if existsInFS(a.fs, c.Request.URL.Path[1:]) {
			c.File(c.Request.URL.Path)
			return
		}
		// Otherwise fallback
		c.FileFromFS("index.html", http.FS(a.fs))
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: a.Router,
	}

	// Start in background
	go func() {
		log.Printf("🚀 Listening on %s...", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	a.StartAllSchedulers()
	defer a.ShutdownAllSchedulers()
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}
	log.Println("Server gracefully stopped")
}

// AddSchedulers registers new schedulers to the app
func (a *App) AddScheduler(schs gocron.Scheduler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.schedulers = append(a.schedulers, schs)
}

// StartAllSchedulers starts all registered schedulers
func (a *App) StartAllSchedulers() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, sch := range a.schedulers {
		sch.Start()
	}
}

// ShutdownAllSchedulers stops all registered schedulers
func (a *App) ShutdownAllSchedulers() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, sch := range a.schedulers {
		if err := sch.Shutdown(); err != nil {
			log.Printf("error shutting down scheduler: %v", err)
		}
	}
}

func ensureDataFolder() {
	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		log.Fatalf("failed to create data folder: %v", err)
	}
}

// existsInFS checks for a path in the embedded FS
func existsInFS(fsys fs.FS, path string) bool {
	if path == "" {
		return false
	}
	_, err := fsys.Open(path)
	return err == nil
}
