package main

import (
	"fmt"
	"time"

	"github.com/Nomadcxx/sysc-Go/animations"
)

func main() {
	// Get terminal size (hardcoded for now)
	width, height := 80, 24

	// Create fire effect with Dracula theme
	palette := animations.GetFirePalette("dracula")
	fire := animations.NewFireEffect(width, height, palette)

	// Run animation loop
	fmt.Print("\033[2J\033[H") // Clear screen
	fmt.Print("\033[?25l")     // Hide cursor

	for i := 0; i < 200; i++ { // Run for 200 frames
		fire.Update(i)
		output := fire.Render()

		fmt.Print("\033[H") // Move cursor to top
		fmt.Print(output)
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Print("\033[?25h") // Show cursor
	fmt.Println("\nFire animation complete!")
}
