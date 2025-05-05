package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type RadioClient struct {
	rootFolder string
	url        string
}

func NewRadioClient(rootFolder, radioURL string) *RadioClient {
	err := os.MkdirAll(rootFolder, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	return &RadioClient{
		rootFolder: rootFolder,
		url:        radioURL,
	}
}

func (rc *RadioClient) Record(duration int, outputPath string) error {
	args := []string{
		"-i", rc.url, "-t", strconv.Itoa(duration), outputPath,
	}

	if strings.Contains(outputPath, "test") {
		args = append(args, "-y")
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		log.Println(err)

	}
	return nil
}
