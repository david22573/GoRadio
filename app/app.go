package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/david22573/GoRadio/app/store"
	"github.com/go-co-op/gocron/v2"
)

type App struct {
	Repo store.RadioRepository

	schedulers []gocron.Scheduler
	mu         sync.Mutex
}

func NewApp(repo store.RadioRepository) *App {
	ensureDataFolder()
	schdulers := make([]gocron.Scheduler, 0)
	return &App{Repo: repo, schedulers: schdulers}
}

func (app *App) Run() {
	router := NewRouter()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
	log.Printf("🚀 server listening on %s", srv.Addr)

	// 4. Wait for SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
}

// AddScheduler adds a new scheduler to the app
func (a *App) AddSchedulers(schedulers ...gocron.Scheduler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, scheduler := range schedulers {
		a.schedulers = append(a.schedulers, scheduler)
	}
}

// StartAllSchedulers starts all schedulers
func (a *App) StartAllSchedulers() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, scheduler := range a.schedulers {
		scheduler.Start()
	}

	return nil
}

// ShutdownAllSchedulers stops all schedulers
func (a *App) ShutdownAllSchedulers() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var lastErr error
	for i, scheduler := range a.schedulers {
		if err := scheduler.Shutdown(); err != nil {
			lastErr = fmt.Errorf("failed to shutdown scheduler %d: %w", i, err)
			fmt.Printf("Error shutting down scheduler %d: %v\n", i, err)
		}
	}

	return lastErr
}

func (a *App) Shutdown() {
	a.ShutdownAllSchedulers()
	// Clean shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}

func ensureDataFolder() {
	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		log.Fatalf("failed to create data folder: %v", err)
	}
}
