package main

import (
	"log"

	"github.com/david22573/GoRadio/app"
	"github.com/david22573/GoRadio/app/db/sqlite"
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
	db, err := sqlite.NewSQLiteClient("radio.db")
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

	app, err := app.NewApp(db)
	if err != nil {
		log.Fatal(err)
	}

	app.Run(":8000")
}
