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

func CreateLoop(audioFile string, targetMinutes int, outputName string) (string, error) {
	if strings.TrimSpace(outputName) == "" {
		outputName = "looped_30min.mp3"
	}
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
	outputFile, err := filepath.Abs(filepath.Join(outputDir, outputName))
	if err != nil {
		return "", err
	}
	
	// mp3 supports lossless byte-level concatenation via the ffmpeg concat
	// protocol — no timestamp math, unlike the concat demuxer which emits
	// non-monotonic DTS warnings on mp3's negative encoder-delay start.
	var concat strings.Builder
	concat.WriteString("concat:")
	for i := 0; i < loops; i++ {
		if i > 0 {
			concat.WriteByte('|')
		}
		concat.WriteString(filepath.Base(audioFile))
	}
	
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", concat.String(),
		"-c", "copy",
		outputFile,
	)
	// Relative paths keep the concat URL well under PATH_MAX.
	// ponytail: concat protocol caps at ~100 segments (PATH_MAX/len(name));
	// songs under ~20s at 30min target would exceed it — switch to concat
	// demuxer + re-encode if that ever matters.
	cmd.Dir = filepath.Dir(audioFile)
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}
	
	return outputFile, nil
}
