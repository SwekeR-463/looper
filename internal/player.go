package internal

import (
	"fmt"
	"os"
	"os/exec"
)

func Play(audioFile string) error {
	// Try mpv first, then ffplay
	players := [][]string{
		{"mpv", audioFile},
		{"ffplay", "-nodisp", "-autoexit", audioFile},
	}
	
	for _, player := range players {
		if _, err := exec.LookPath(player[0]); err == nil {
			cmd := exec.Command(player[0], player[1:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
	}
	
	return fmt.Errorf("no suitable player found (tried: mpv, ffplay)")
}
