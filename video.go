package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

type ffProbeOutput struct {
	Streams []struct {
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	} `json:"streams"`
}

func getVideoAspectRatioPrefix(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	buffer := bytes.NewBuffer(nil)
	cmd.Stdout = buffer
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf("ffprobe command failed: %w", err)
		return "", err
	}

	output := ffProbeOutput{}
	if err := json.Unmarshal(buffer.Bytes(), &output); err != nil {
		err := fmt.Errorf("failed to unmarshal ffprobe output: %w", err)
		return "", err
	}

	width := output.Streams[0].Width
	height := output.Streams[0].Height

	if width == 0 || height == 0 {
		return "", fmt.Errorf("invalid video dimensions: width %.3f, height %.3f", width, height)
	}

	ratio := width / height

	tolerance := 0.1

	horizontalRatio := 16.0 / 9.0

	if ratio > horizontalRatio-tolerance && ratio < horizontalRatio+tolerance {
		return "landscape", nil
	}

	verticalRatio := 9.0 / 16.0

	if ratio > verticalRatio-tolerance && ratio < verticalRatio+tolerance {
		return "portrait", nil
	}

	return "other", nil
}

func processVideoForFastStart(filePath string) (string, error) {
	outputFilePath := filePath + ".processing"

	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputFilePath)
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf("ffmpeg command failed: %w", err)
		return "", err
	}

	return outputFilePath, nil
}
