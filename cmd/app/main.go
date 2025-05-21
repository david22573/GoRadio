package main

import (
	"log"

	"github.com/david22573/GoRadio/app"
	"github.com/david22573/GoRadio/app/schedulers"
	"github.com/david22573/GoRadio/app/store/repos/sqlite"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "GoRadio.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	})
	repo, err := sqlite.NewSQLiteRepo("radio.db")
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

	app, err := app.NewApp(repo)
	if err != nil {
		log.Fatal(err)
	}
	if station, err := repo.GetAllStations(); err == nil {
		for _, s := range station {
			scheduler, err := schedulers.NewRadioScheduler(app, s, nil)
			if err != nil {
				log.Fatal(err)
			}
			app.AddScheduler(scheduler)
		}
	}

	app.Run(":8000")
}
