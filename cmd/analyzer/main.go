package main

import (
	"context"
	"log"
	"time"

	"github.com/david22573/GoRadio/app/audio"
	"github.com/david22573/GoRadio/app/db/sqlite"
)

func main() {
	// 1. Initialize DB
	db, err := sqlite.NewSQLiteClient("radio.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 2. Initialize Analyzer
	analyzer := audio.NewAnalyzer()

	// 3. Get all tracks
	tracks, err := db.GetAllTracks()
	if err != nil {
		log.Fatalf("Failed to get tracks: %v", err)
	}

	if len(tracks) == 0 {
		log.Println("No tracks to analyze.")
		return
	}

	log.Printf("Starting batch analysis for %d tracks...", len(tracks))

	for _, t := range tracks {
		log.Printf("Processing: %s (%s)", t.Title, t.URL)

		if t.URL == "" {
			log.Printf("  [!] Skipping: no URL")
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		features, err := analyzer.AnalyzeURL(ctx, t.URL, 30)
		cancel()
		if err != nil {
			log.Printf("  [!] Analysis failed: %v", err)
			continue
		}

		// Update vector DB keyed by the real track id
		err = db.InsertTrackVector(t.ID, features.Embedding)
		if err != nil {
			log.Printf("  [!] DB Insert failed: %v", err)
		} else {
			log.Printf("  [✓] Success: Vector stored for track %d", t.ID)
		}
	}

	log.Println("Batch analysis complete.")
}
