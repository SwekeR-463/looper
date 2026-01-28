package internal

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)



func getDuration(audioFile string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioFile,
	)
	
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}
	
	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}
	
	return duration, nil
}

func CreateLoop(audioFile string, targetMinutes int) (string, error) {
	duration, err := getDuration(audioFile)
	if err != nil {
		return "", err
	}
	
	// Cap at 30 minutes max
	if targetMinutes > 30 {
		targetMinutes = 30
	}
	
	targetSeconds := targetMinutes * 60
	
	// Calculate number of loops (floor to stay under target)
	loops := int(math.Floor(float64(targetSeconds) / duration))
	if loops < 1 {
		loops = 1
	}
	totalDuration := loops * int(duration)
	
	fmt.Printf("Song duration: %.2f seconds\n", duration)
	fmt.Printf("Creating %d loops = %d seconds (%.2f minutes)\n", loops, totalDuration, float64(totalDuration)/60)
	
	// Create output directory
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Generate output filename
	outputFile := filepath.Join(outputDir, "looped_30min.mp3")
	
	// Use ffmpeg to concatenate the file with itself
	// Method: use concat demuxer with a file list
	tmpDir := filepath.Dir(audioFile)
	listFile := filepath.Join(tmpDir, "concat_list.txt")
	var listContent strings.Builder
	for i := 0; i < loops; i++ {
		listContent.WriteString(fmt.Sprintf("file '%s'\n", audioFile))
	}
	
	if err := os.WriteFile(listFile, []byte(listContent.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write concat list: %w", err)
	}
	defer os.Remove(listFile)
	
	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", listFile,
		"-c", "copy",
		outputFile,
	)
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}
	
	return outputFile, nil
}
