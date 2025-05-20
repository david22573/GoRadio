package main

import (
	"github.com/david22573/GoRadio/app"
	"github.com/david22573/GoRadio/app/schedulers"
	"github.com/david22573/GoRadio/app/store/repos/sqlite"
	"github.com/david22573/GoRadio/app/types"
)

func main() {
	repo, err := sqlite.NewSQLiteRepo("radio.db")
	if err != nil {
		panic(err)
	}

	app := app.NewApp(repo)
	testStation := types.Station{
		Name: "kxlu",
		URL:  "http://ksl.com",
	}
	rs, err := schedulers.NewRadioScheduler(app, testStation, nil)
	if err != nil {
		panic(err)
	}
	app.AddSchedulers(rs)
	app.StartAllSchedulers()
	app.Run()
}
