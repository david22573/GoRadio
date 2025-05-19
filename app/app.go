package app

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/david22573/GoRadio/app/store"
	"github.com/go-co-op/gocron/v2"
)

type App struct {
	repo       store.RadioRepository
	schedulers []gocron.Scheduler
}

func NewApp(repo store.RadioRepository) *App {
	ensureDataFolder()
	app := &App{repo: repo}
	return app
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
