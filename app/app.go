package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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
	router := api.NewRouter()
	api.RegisterHandlers(router)
	for _, scheduler := range app.schedulers {
		scheduler.Start()
	}
	log.Default().Fatal(router.Run(":8080"))
}

func (app *App) RegisterSchedulers() {
	stations, err := app.repo.GetAllStations()
	if err != nil {
		log.Fatal(err)
	}
	for _, station := range stations {
		scheduler := NewRadioScheduler(&station)
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
