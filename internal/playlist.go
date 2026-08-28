package internal

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Track struct {
	URL  string
	Path string
	Dur  float64
}

func runFFmpeg(args ...string) error {
	cmd := exec.Command("ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return err
		}
		return fmt.Errorf("%w: %s", err, msg)
	}
	return nil
}

func Normalize(src, dst string) error {
	return runFFmpeg("-y", "-i", src,
		"-ar", "44100",
		"-ac", "2",
		"-c:a", "libmp3lame", "-q:a", "0",
		dst,
	)
}

func MakeSilence(secs int, dst string) error {
	return runFFmpeg("-y",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=stereo",
		"-t", strconv.Itoa(secs),
		"-c:a", "libmp3lame", "-q:a", "0",
		dst,
	)
}

func DownloadAll(urls []string, workDir string) ([]Track, []error) {
	var tracks []Track
	var errs []error

	for i, url := range urls {
		fmt.Printf("[%d/%d] Downloading %s\n", i+1, len(urls), url)

		raw, rawDir, err := DownloadAudio(url)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", url, err))
			continue
		}

		dst := filepath.Join(workDir, fmt.Sprintf("n%02d.mp3", i+1))
		if err := Normalize(raw, dst); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", url, err))
			os.RemoveAll(rawDir)
			continue
		}
		os.RemoveAll(rawDir)

		dur, err := getDuration(dst)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", url, err))
			os.Remove(dst)
			continue
		}

		tracks = append(tracks, Track{URL: url, Path: dst, Dur: dur})
	}

	return tracks, errs
}

func BuildConcat(tracks []Track, gap string) string {
	var parts []string
	for i, t := range tracks {
		if i > 0 && gap != "" {
			parts = append(parts, gap)
		}
		parts = append(parts, filepath.Base(t.Path))
	}
	return "concat:" + strings.Join(parts, "|")
}

func Assemble(concat, outFile, dir string) error {
	// Relative paths + cmd.Dir keep the concat URL well under PATH_MAX.
	// ponytail: concat protocol caps at ~100 segments (2n-1), so ~50 songs;
	// switch to the concat demuxer if that ever matters.
	cmd := exec.Command("ffmpeg", "-y", "-i", concat, "-c", "copy", outFile)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg assemble failed: %w: %s", err, stderr.String())
	}
	return nil
}

func CreatePlaylist(urls []string, gapSecs int, outputName string) (string, error) {
	if strings.TrimSpace(outputName) == "" {
		outputName = "playlist.mp3"
	}

	workDir, err := os.MkdirTemp("", "looper-playlist-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(workDir)

	// ponytail: downloads run one at a time — 10 links is a few minutes of
	// waiting. Wrap the loop in errgroup if that ever matters.
	tracks, errs := DownloadAll(urls, workDir)
	if len(tracks) == 0 {
		return "", fmt.Errorf("no songs downloaded, %d failed: %w", len(errs), errors.Join(errs...))
	}
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", e)
	}

	gapName := ""
	if gapSecs > 0 {
		gapFile := filepath.Join(workDir, "gap.mp3")
		if err := MakeSilence(gapSecs, gapFile); err != nil {
			return "", err
		}
		gapName = filepath.Base(gapFile)
	}

	total := 0.0
	for _, t := range tracks {
		total += t.Dur
	}
	if gapSecs > 0 {
		total += float64(gapSecs * (len(tracks) - 1))
	}
	fmt.Printf("Playlist: %d songs, %.2f minutes\n", len(tracks), total/60)

	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	outFile, err := filepath.Abs(filepath.Join(outputDir, outputName))
	if err != nil {
		return "", err
	}

	if err := Assemble(BuildConcat(tracks, gapName), outFile, workDir); err != nil {
		return "", err
	}

	return outFile, nil
}

func ReadLinks(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var urls []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}
	return urls, nil
}
