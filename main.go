package main

import (
	"time"

	"github.com/david22573/GoRadio/app"
	"github.com/david22573/GoRadio/app/schedulers"
	"github.com/david22573/GoRadio/app/store/repos/sqlite"
	"github.com/go-co-op/gocron/v2"
)

func main() {
	repo, err := sqlite.NewSQLiteRepo("radio.db")
	if err != nil {
		panic(err)
	}

	app := app.NewApp(repo)
	rs, err := schedulers.NewRadioScheduler(app, "kxlu", "http://ksl.com", nil)
	if err != nil {
		panic(err)
	}
	rs.NewJob(
		gocron.WeeklyJob(
			1,
			gocron.NewWeekdays(time.Sunday),
			gocron.NewAtTimes(gocron.NewAtTime(21, 1, 0)),
		),
		gocron.NewTask(func() { println("hi") }),
	)
	rs.Start()
	app.Run()
}
