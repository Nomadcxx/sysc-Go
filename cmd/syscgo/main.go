package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Nomadcxx/sysc-Go/animations"
	"golang.org/x/term"
)

func main() {
	effect := flag.String("effect", "fire", "Animation effect (fire, matrix, rain, fireworks, ticker)")
	theme := flag.String("theme", "dracula", "Color theme")
	duration := flag.Int("duration", 10, "Duration in seconds (0 = infinite)")
	flag.Parse()

	// Get terminal size
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 80, 24
	}

	// Setup terminal
	fmt.Print("\033[2J\033[H") // Clear screen
	fmt.Print("\033[?25l")     // Hide cursor
	defer fmt.Print("\033[?25h") // Show cursor on exit

	// Calculate frame count (0 = infinite)
	frames := 0
	if *duration > 0 {
		frames = *duration * 20 // 20 fps
	}

	switch *effect {
	case "fire":
		runFire(width, height, *theme, frames)
	default:
		fmt.Printf("Unknown effect: %s\n", *effect)
		fmt.Println("Available: fire, matrix, rain, fireworks, ticker")
		os.Exit(1)
	}
}

func runFire(width, height int, theme string, frames int) {
	palette := animations.GetFirePalette(theme)
	fire := animations.NewFireEffect(width, height, palette)

	frame := 0
	for frames == 0 || frame < frames {
		fire.Update(frame)
		output := fire.Render()

		fmt.Print("\033[H") // Move cursor to top
		fmt.Print(output)
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}
