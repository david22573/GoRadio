package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/david22573/GoRadio/app/api"
	"github.com/david22573/GoRadio/app/store"
	sqlite "github.com/david22573/GoRadio/app/store/repos/sqilte"
)

type App struct {
	repo       store.RadioRepository
	schedulers []*RadioScheduler
}

func NewApp() *App {
	sqliteRepo, _ := sqlite.NewSQLiteRepo("./radio.db")
	scheduler := NewRadioScheduler("KXLU", "https://stream.kxlu-fm.com/kxlu")
	return &App{schedulers: []*RadioScheduler{scheduler}, repo: sqliteRepo}
}

func (app *App) Run() {
	router := api.NewRouter()
	log.Default().Fatal(router.Run(":8080"))

	// Clean shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}
