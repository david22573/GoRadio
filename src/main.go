package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Create scheduler + client
	rs := NewRadioScheduler("/radio", "http://kxlu.streamguys1.com/kxlu-lo")

	// Define your shows
	shows := []Show{
		NewShow(
			"KXLU",
			ShowSchedule{
				Day:  time.Saturday,
				Hour: 18,
				Min:  01,
			},
			1*time.Hour,
		),
	}

	// Schedule them
	rs.AddShows(shows...)

	// Start scheduler in background
	rs.Start()
	log.Println("Scheduler started. Waiting for jobs...")

	// Clean shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down scheduler...")
	rs.Shutdown()
	fmt.Println("Goodbye!")
}
