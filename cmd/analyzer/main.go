package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
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

	// 3. Get all stations
	stations, err := db.GetAllStations()
	if err != nil {
		log.Fatalf("Failed to get stations: %v", err)
	}

	log.Printf("Starting batch analysis for %d stations...", len(stations))

	for _, s := range stations {
		log.Printf("Processing: %s (%s)", s.Name, s.URL)

		// Create temp file for sample
		tempFile := fmt.Sprintf("data/sample_%d.wav", s.ID)
		
		// Sample 30 seconds using FFmpeg
		err := sampleStream(s.URL, tempFile, 30)
		if err != nil {
			log.Printf("  [!] Sampling failed: %v", err)
			continue
		}

		// Extract features
		features, err := analyzer.AnalyzeFile(tempFile)
		if err != nil {
			log.Printf("  [!] Analysis failed: %v", err)
			os.Remove(tempFile)
			continue
		}

		// Update vector DB
		err = db.InsertTrackVector(s.ID, features.Embedding)
		if err != nil {
			log.Printf("  [!] DB Insert failed: %v", err)
		} else {
			log.Printf("  [✓] Success: Vector stored")
		}

		// Cleanup
		os.Remove(tempFile)
	}

	log.Println("Batch analysis complete.")
}

// sampleStream uses ffmpeg to capture N seconds of a stream and save as WAV
func sampleStream(url, outPath string, seconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(seconds+10)*time.Second)
	defer cancel()

	// ffmpeg -i [url] -t [seconds] -acodec pcm_s16le -ar 44100 -ac 1 [outPath]
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y", // Overwrite
		"-i", url,
		"-t", fmt.Sprintf("%d", seconds),
		"-acodec", "pcm_s16le",
		"-ar", "44100",
		"-ac", "1",
		outPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %w (output: %s)", err, string(output))
	}

	return nil
}
