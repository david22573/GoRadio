package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/david22573/GoRadio/app/api"
	"github.com/david22573/GoRadio/app/store"
)

type App struct {
	repo       store.RadioRepository
	schedulers []*RadioScheduler
}

func NewApp(repo store.RadioRepository) *App {
	ensureDataFolder()
	app := &App{repo: repo}
	app.RegisterSchedulers()
	return app
}

func (app *App) Run() {
	// 1. Build router
	router := api.NewRouter()
	api.RegisterHandlers(router)

	// 2. Create HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// 3. Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 4. Listen for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received")

	// 5. Stop schedulers
	for _, sch := range app.schedulers {
		sch.Shutdown()
	}

	// 6. Gracefully shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Graceful shutdown complete")
}

func (app *App) RegisterSchedulers() {
	stations, err := app.repo.GetAllStations()
	if err != nil {
		log.Fatal(err)
	}
	for i := range stations {
		scheduler := NewRadioScheduler(&stations[i])
		app.schedulers = append(app.schedulers, scheduler)
	}
}

func (app *App) Shutdown() {
	for _, scheduler := range app.schedulers {
		scheduler.Shutdown()
	}
	// Clean shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Println("graceful shutdown")
}

func ensureDataFolder() {
	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		log.Fatalf("failed to create data folder: %v", err)
	}
}
