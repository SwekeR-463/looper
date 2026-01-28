package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func DownloadAudio(url string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "looper-*")
	if err != nil {
		return "", err
	}
	
	outputTemplate := filepath.Join(tmpDir, "audio.%(ext)s")
	
	cmd := exec.Command("yt-dlp",
		"-x",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"-o", outputTemplate,
		url,
	)
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("yt-dlp failed: %w", err)
	}
	
	// Find the downloaded file
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}
	
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".mp3" {
			return filepath.Join(tmpDir, entry.Name()), nil
		}
	}
	
	os.RemoveAll(tmpDir)
	return "", fmt.Errorf("no mp3 file found after download")
}
