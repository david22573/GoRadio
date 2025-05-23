package app

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/david22573/GoRadio/app/db"
	"github.com/david22573/GoRadio/app/store"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
)

//go:embed static/*
var staticFiles embed.FS

type App struct {
	Router *gin.Engine
	Store  store.RadioActions

	schedulers []gocron.Scheduler
	mu         sync.Mutex
	httpSrv    *http.Server
}

func NewApp(db db.Client) (*App, error) {
	r := gin.Default()
	return &App{
		Router: r,
		Store:  store.NewRadioStore(db),

		schedulers: []gocron.Scheduler{},
	}, nil
}

// if station, err := store.GetAllStations(); err == nil {
// 	for _, s := range station {
// 		scheduler, err := schedulers.NewRadioScheduler(app, s, nil)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		app.AddScheduler(scheduler)
// 	}
// }

func (a *App) AddScheduler(sch gocron.Scheduler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.schedulers = append(a.schedulers, sch)
}

func (a *App) StartSchedulers() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, sch := range a.schedulers {
		sch.Start()
	}
}

func (a *App) StopSchedulers() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, sch := range a.schedulers {
		sch.Shutdown()
	}
}

func (a *App) Run(addr string) error {
	a.StartSchedulers()

	r := a.Router

	staticFS, err := static.EmbedFolder(staticFiles, "static")

	if err != nil {
		return err
	}

	r.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })

	r.Use(static.Serve("/", staticFS))
	// Or if you need SPA fallback for client-side routing:
	r.NoRoute(func(c *gin.Context) {
		c.File("app/static/index.html")
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	a.httpSrv = srv

	errChan := make(chan error, 1)
	go func() {
		fmt.Printf("🚀 Serving on http://localhost%s\n", addr)
		errChan <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Printf("Received signal %s: initiating shutdown.", sig)
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
			a.StopSchedulers()
			return err
		}
		log.Println("HTTP server exited cleanly.")
	}

	a.StopSchedulers()
	log.Println("Schedulers stopped.")
	return nil
}
