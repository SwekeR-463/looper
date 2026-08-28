package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"looper/internal"
)

var rootCmd = &cobra.Command{
	Use:   "looper [youtube-url...]",
	Short: "Create a looped audio file, or a serial playlist, from YouTube links",
	Long:  `Downloads YouTube audio and either loops a single song for a target duration, or joins several songs serially with a silence gap between them.`,
	Args: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		if file != "" {
			return nil
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := cmd.Flags().GetString("file")
		if file != "" && len(args) > 0 {
			fmt.Fprintln(os.Stderr, "Error: pass links as arguments or with -f <file>, not both")
			os.Exit(1)
		}

		urls := args
		if file != "" {
			var err error
			urls, err = internal.ReadLinks(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading links: %v\n", err)
				os.Exit(1)
			}
		}
		if len(urls) == 0 {
			fmt.Fprintln(os.Stderr, "Error: no links given")
			os.Exit(1)
		}

		outputName := "looped_30min.mp3"
		if len(urls) > 1 {
			outputName = "playlist.mp3"
		}
		if cmd.Flags().Changed("output") {
			outputName, _ = cmd.Flags().GetString("output")
		}

		var outFile string
		var err error

		if len(urls) == 1 {
			outFile, err = runLoop(cmd, urls[0], outputName)
		} else {
			outFile, err = runPlaylist(cmd, urls, outputName)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Created: %s\n", outFile)

		play, _ := cmd.Flags().GetBool("play")
		if play {
			fmt.Println("Playing...")
			if err := internal.Play(outFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error playing: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func runLoop(cmd *cobra.Command, url, outputName string) (string, error) {
	duration, _ := cmd.Flags().GetInt("duration")
	if duration <= 0 {
		return "", fmt.Errorf("duration must be greater than 0")
	}

	fmt.Println("Downloading audio...")
	audioFile, tmpDir, err := internal.DownloadAudio(url)
	if err != nil {
		return "", fmt.Errorf("downloading: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Creating %d min loop (max)...\n", duration)
	return internal.CreateLoop(audioFile, duration, outputName)
}

func runPlaylist(cmd *cobra.Command, urls []string, outputName string) (string, error) {
	if cmd.Flags().Changed("duration") {
		fmt.Fprintln(os.Stderr, "Warning: -d is ignored in playlist mode, length is the sum of the songs")
	}

	gap, _ := cmd.Flags().GetInt("gap")
	if gap < 0 {
		return "", fmt.Errorf("gap must be 0 or greater")
	}

	return internal.CreatePlaylist(urls, gap, outputName)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("play", "p", false, "Play the audio after creation")
	rootCmd.Flags().IntP("duration", "d", 30, "Target duration in minutes, loop mode only (default: 30, max: 30)")
	rootCmd.Flags().StringP("output", "o", "looped_30min.mp3", "Output file name")
	rootCmd.Flags().StringP("file", "f", "", "Read links from a file, one per line (# comments allowed)")
	rootCmd.Flags().IntP("gap", "g", 2, "Silence between songs in seconds, playlist mode only")
}
