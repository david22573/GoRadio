package main

import (
	"log"

	"github.com/david22573/GoRadio/app"
	sqlite "github.com/david22573/GoRadio/app/store/repos/sqilte"
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
	sqliteRepo, err := sqlite.NewSQLiteRepo("data/radio.db")
	if err != nil {
		log.Fatal(err)
	}
	app := app.NewApp(sqliteRepo)
	app.Run()
}
