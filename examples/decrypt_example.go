package main

import (
	"fmt"
	"time"

	"github.com/Nomadcxx/sysc-Go/animations"
)

func main() {
	// Example text to decrypt (could also read from a file)
	text := "SECRET MESSAGE: THE PASSWORD IS \"OPEN SESAME\""

	// Create a decrypt effect configuration
	config := animations.DecryptConfig{
		Width:                  80,
		Height:                 24,
		Text:                   text,
		TypingSpeed:            2,
		CiphertextColors:       []string{"#008000", "#00cb00", "#00ff00"},
		FinalGradientStops:     []string{"#ff79c6", "#8be9fd"},
		FinalGradientSteps:     12,
		FinalGradientDirection: "horizontal",
	}

	// Create the decrypt effect
	decrypt := animations.NewDecryptEffect(config)

	// Run for 300 frames (about 15 seconds)
	for i := 0; i < 300; i++ {
		decrypt.Update()
		output := decrypt.Render()

		// Clear screen and print output
		fmt.Print("\033[2J\033[H")
		fmt.Print(output)

		time.Sleep(50 * time.Millisecond)
	}
}
