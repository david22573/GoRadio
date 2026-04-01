package main

import (
	"log"

	"github.com/david22573/GoRadio/app"
	"github.com/david22573/GoRadio/app/db/sqlite"
	"github.com/david22573/GoRadio/app/schedulers"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "GoRadio.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	})

	db, err := sqlite.NewSQLiteClient("radio.db")
	if err != nil {
		log.Fatal(err)
	}

	application, err := app.NewApp(db)
	if err != nil {
		log.Fatal(err)
	}

	stations, err := db.GetAllStations()
	if err == nil {
		for _, s := range stations {
			scheduler, err := schedulers.NewRadioScheduler(application, s, nil)
			if err != nil {
				log.Printf("Failed to create scheduler for station %s: %v", s.Name, err)
				continue
			}
			application.AddScheduler(scheduler)
		}
	} else {
		log.Printf("Could not load stations for scheduling: %v", err)
	}

	if err := application.Run(":8000"); err != nil {
		log.Fatal(err)
	}
}
