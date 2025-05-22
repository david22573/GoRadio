package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/david22573/GoRadio/app/store"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
)

// App holds application state, HTTP server, and scheduled jobs.
type App struct {
	Router *gin.Engine
	Repo   store.RadioRepository

	schedulers []gocron.Scheduler
	mu         sync.Mutex
	httpSrv    *http.Server
}

// NewApp constructs an App, embedding static assets and setting up storage.
func NewApp(repo store.RadioRepository) (*App, error) {
	r := gin.Default()
	return &App{
		Router:     r,
		Repo:       repo,
		schedulers: []gocron.Scheduler{},
	}, nil
}

// AddScheduler registers a new scheduler.
func (a *App) AddScheduler(sch gocron.Scheduler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.schedulers = append(a.schedulers, sch)
}

// StartSchedulers starts all registered schedulers.
func (a *App) StartSchedulers() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, sch := range a.schedulers {
		sch.Start()
	}
}

// StopSchedulers gracefully stops all schedulers.
func (a *App) StopSchedulers() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, sch := range a.schedulers {
		sch.Shutdown()
	}
}

// Run starts the HTTP server, cron jobs, and handles graceful shutdown.
func (a *App) Run(addr string) error {
	a.Router.StaticFS("/", http.Dir("frontend/build"))
	a.Router.NoRoute(func(c *gin.Context) {
		// Only serve index.html if the request is likely for an HTML page
		// and not for a missing static asset (which StaticFS would have handled if it existed)
		// This check helps avoid serving index.html for things like missing image requests.
		// However, for a strict SPA setup where SvelteKit handles all routing,
		// always serving index.html for NoRoute might be desired.

		// Check if the request path is under our Svelte app's base path
		if strings.HasPrefix(c.Request.URL.Path, "/") {
			// Construct the path to the fallback HTML file
			fallbackFile := filepath.Join("frontend/build", "404.html") // Or your configured fallback, e.g., "200.html"

			// Check if the fallback file exists
			if _, err := os.Stat(fallbackFile); !os.IsNotExist(err) {
				c.File(fallbackFile)
				return
			}
		}

		// Default 404 if no Svelte app base path match or fallback file doesn't exist
		c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	a.Router.Static("/static", "static")
	a.Router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	RegisterRoutes(a)

	srv := &http.Server{
		Addr:    addr,
		Handler: a.Router,
	}
	a.httpSrv = srv

	a.StartSchedulers()

	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.ListenAndServe()
	}()

	fmt.Printf("🚀 Serving on http://localhost%s\n", addr)
	stop := make(chan os.Signal, 2)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Printf("Received signal %s: initiating shutdown.", sig)
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
			a.StopSchedulers()
			return err
		}
		log.Println("HTTP server exited cleanly.")
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server gracefully stopped.")
	}

	// Stop schedulers after HTTP shutdown
	a.StopSchedulers()
	log.Println("Schedulers stopped.")

	return nil
}
