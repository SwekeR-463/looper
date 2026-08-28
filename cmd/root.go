package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"looper/internal"
)

var rootCmd = &cobra.Command{
	Use:   "looper [youtube-url]",
	Short: "Create a looped audio file from YouTube",
	Long:  `Downloads a YouTube video's audio and loops it for at least the specified duration, completing the final full loop.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		
		duration, _ := cmd.Flags().GetInt("duration")
		if duration <= 0 {
			fmt.Fprintf(os.Stderr, "Duration must be greater than 0\n")
			os.Exit(1)
		}
		
		outputName, _ := cmd.Flags().GetString("output")
		
		fmt.Println("Downloading audio...")
		audioFile, err := internal.DownloadAudio(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading: %v\n", err)
			os.Exit(1)
		}
		defer os.Remove(audioFile)
		
		fmt.Printf("Creating %d min loop (max)...\n", duration)
		loopedFile, err := internal.CreateLoop(audioFile, duration, outputName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating loop: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("Looped file created: %s\n", loopedFile)
		
		play, _ := cmd.Flags().GetBool("play")
		if play {
			fmt.Println("Playing...")
			if err := internal.Play(loopedFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error playing: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("play", "p", false, "Play the looped audio after creation")
	rootCmd.Flags().IntP("duration", "d", 30, "Target duration in minutes (default: 30, max: 30)")
	rootCmd.Flags().StringP("output", "o", "looped_30min.mp3", "Output file name")
}
