package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Nomadcxx/sysc-Go/animations"
	"golang.org/x/term"
)

const banner = `▄▀▀▀▀ █   █ ▄▀▀▀▀ ▄▀▀▀▀    ▄▀    ▄▀ 
 ▀▀▀▄ ▀▀▀▀█  ▀▀▀▄ █      ▄▀    ▄▀   
▀▀▀▀  ▀▀▀▀▀ ▀▀▀▀   ▀▀▀▀ ▀     ▀

Terminal Animation Library
`

func showHelp() {
	fmt.Print(banner)
	fmt.Println("Usage: syscgo [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -effect string")
	fmt.Println("        Animation effect (default: fire)")
	fmt.Println("        Available: fire, matrix, rain, fireworks")
	fmt.Println()
	fmt.Println("  -theme string")
	fmt.Println("        Color theme (default: dracula)")
	fmt.Println("        Available themes:")
	fmt.Println("          dracula       - Purple and pink vampiric vibes")
	fmt.Println("          gruvbox       - Retro warm colors")
	fmt.Println("          nord          - Cool arctic palette")
	fmt.Println("          tokyo-night   - Neon Tokyo nights")
	fmt.Println("          catppuccin    - Soothing pastel tones")
	fmt.Println("          material      - Google Material colors")
	fmt.Println("          solarized     - Classic precision colors")
	fmt.Println("          monochrome    - Grayscale aesthetic")
	fmt.Println("          transishardjob - Trans pride colors")
	fmt.Println()
	fmt.Println("  -duration int")
	fmt.Println("        Duration in seconds (0 = infinite, default: 10)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  syscgo -effect fire -theme dracula")
	fmt.Println("  syscgo -effect matrix -theme nord -duration 30")
	fmt.Println("  syscgo -effect fireworks -theme gruvbox -duration 0")
	fmt.Println()
}

func main() {
	effect := flag.String("effect", "fire", "Animation effect (fire, matrix, rain, fireworks)")
	theme := flag.String("theme", "dracula", "Color theme")
	duration := flag.Int("duration", 10, "Duration in seconds (0 = infinite)")
	help := flag.Bool("h", false, "Show help")
	flag.BoolVar(help, "help", false, "Show help")
	
	flag.Usage = showHelp
	flag.Parse()

	if *help {
		showHelp()
		return
	}

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
	case "matrix":
		runMatrix(width, height, *theme, frames)
	case "fireworks":
		runFireworks(width, height, *theme, frames)
	case "rain":
		runRain(width, height, *theme, frames)
	default:
		fmt.Printf("Unknown effect: %s\n", *effect)
		fmt.Println("Available: fire, matrix, rain, fireworks")
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

func runMatrix(width, height int, theme string, frames int) {
	palette := animations.GetMatrixPalette(theme)
	matrix := animations.NewMatrixEffect(width, height, palette)

	frame := 0
	for frames == 0 || frame < frames {
		matrix.Update(frame)
		output := matrix.Render()

		fmt.Print("\033[H") // Move cursor to top
		fmt.Print(output)
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runFireworks(width, height int, theme string, frames int) {
	palette := animations.GetFireworksPalette(theme)
	fireworks := animations.NewFireworksEffect(width, height, palette)

	frame := 0
	for frames == 0 || frame < frames {
		fireworks.Update(frame)
		output := fireworks.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runRain(width, height int, theme string, frames int) {
	palette := animations.GetRainPalette(theme)
	rain := animations.NewRainEffect(width, height, palette)

	frame := 0
	for frames == 0 || frame < frames {
		rain.Update(frame)
		output := rain.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

