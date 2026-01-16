package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Nomadcxx/sysc-Go/animations"
	"golang.org/x/term"
)

func getTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 24 // Fallback
	}
	return width, height
}

func setupTerminal() {
	fmt.Print("\033[2J")   // Clear screen
	fmt.Print("\033[H")    // Move cursor to top
	fmt.Print("\033[?25l") // Hide cursor
}

func restoreTerminal() {
	fmt.Print("\033[?25h") // Show cursor
}

func main() {
	setupTerminal()
	defer restoreTerminal()

	// Get terminal size
	width, height := getTerminalSize()

	// Create fire effect with Dracula theme
	palette := animations.GetFirePalette("dracula")
	fire := animations.NewFireEffect(width, height, palette)

	// Run animation loop
	for i := 0; i < 200; i++ { // Run for 200 frames
		fire.Update()
		output := fire.Render()

		fmt.Print("\033[H") // Move cursor to top
		fmt.Print(output)
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println("\nFire animation complete!")
}
