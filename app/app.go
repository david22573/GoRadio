package app

import (
	"context"
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
	r.LoadHTMLGlob("templates/**/*.tmpl")
	r.Static("/static", "static")

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

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
