package app

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

type RadioClient struct {
	outputPath string
	url        string
}

// NewRadioClient ensures the output folder exists.
func NewRadioClient(rootFolder, radioURL string) *RadioClient {
	if err := os.MkdirAll(rootFolder, os.ModePerm); err != nil {
		log.Fatalf("failed to create root folder: %v", err)
	}
	cwd, _ := os.Getwd()
	op := filepath.Join(cwd, rootFolder)
	return &RadioClient{
		outputPath: op,
		url:        radioURL,
	}
}

// Record uses ffmpeg to record the stream for show.Duration.
func (rc *RadioClient) Record(show Show) error {
	// timestamped filename
	ts := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s_%s.aac", show.Name, ts)
	showDir := filepath.Join(rc.outputPath, show.Name)
	if err := os.MkdirAll(showDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create show folder: %w", err)
	}
	outPath := filepath.Join(showDir, filename)

	// build ffmpeg args
	args := []string{
		"-i", rc.url,
		"-t", strconv.Itoa(int(show.Duration.Seconds())),
		"-c",
		"copy",
		outPath,
	}

	currDir := exec.Command("explorer .")
	currDir.Run()

	log.Printf("▶️  Recording %s for %v → %s\n", show.Name, show.Duration, outPath)
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg exited with error: %w", err)
	}

	log.Printf("✅ Finished recording %s\n", filename)
	return nil
}
