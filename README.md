# Looper

A CLI tool that downloads YouTube audio and loops it up to 30 minutes (capped, trims if needed).

## Features

- Downloads audio from YouTube videos
- Loops the song to stay under 30 minutes (removes last loop if it would exceed)
- Saves output to `output/looped_30min.mp3`
- Optional playback with `-p` flag

## Installation

### Prerequisites

Install required tools:

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install yt-dlp ffmpeg
```

**macOS:**
```bash
brew install yt-dlp ffmpeg
```

**Optional (for playback):**
```bash
# macOS
brew install mpv

# Ubuntu
sudo apt install mpv
```

### Build

```bash
go build -o looper
```

## Usage

```bash
# Create up to 30 min loop (default)
./looper "https://youtube.com/watch?v=..."

# Create shorter loop (e.g., 15 min)
./looper -d 15 "https://youtube.com/watch?v=..."

# Create and play immediately
./looper -p "https://youtube.com/watch?v=..."
```

## Example

```bash
$ ./looper "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
Downloading audio...
[youtube] Extracting URL: https://www.youtube.com/watch?v=dQw4w9WgXcQ
...
Creating 30 min loop (max)...
Song duration: 212.00 seconds
Creating 8 loops = 1696 seconds (28.27 minutes)
Looped file created: output/looped_30min.mp3
```

## How It Works

1. Downloads audio using `yt-dlp`
2. Gets duration with `ffprobe`
3. Calculates loops: `floor(target_seconds / duration)` to stay under target
4. Concatenates using `ffmpeg` concat demuxer
5. Saves to `output/looped_30min.mp3`

## Credits

Made using **Amp Code** and **BomBadil Agent (Kimi K2.5)**

