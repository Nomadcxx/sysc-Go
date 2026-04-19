package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Nomadcxx/sysc-Go/animations"
	"golang.org/x/term"
)

const version = "1.0.2"

const banner = `▄▀▀▀▀ █   █ ▄▀▀▀▀ ▄▀▀▀▀    ▄▀    ▄▀
 ▀▀▀▄ ▀▀▀▀█  ▀▀▀▄ █      ▄▀    ▄▀
▀▀▀▀  ▀▀▀▀▀ ▀▀▀▀   ▀▀▀▀ ▀     ▀

Terminal Animation Library
`

// wrapText wraps text to fit within the specified width
// findAssetFile searches for an asset file in multiple locations
// Priority order: user writable directories first, then system read-only paths
func findAssetFile(filename string) string {
	// Get the directory containing the binary
	exePath, err := os.Executable()
	var binaryDir string
	if err == nil {
		binaryDir = filepath.Dir(exePath)
	}

	// Priority order matches TUI for consistency:
	// 1. User-writable directories (where TUI saves exports)
	// 2. Local relative paths
	// 3. Binary-relative path
	// 4. System install paths (read-only)
	locations := []string{
		filepath.Join(os.Getenv("HOME"), "sysc-Go", "assets", filename), // User home (writable, TUI saves here)
		filepath.Join("assets", filename),                               // ./assets/ (current dir)
		filepath.Join("../assets", filename),                            // ../assets/ (parent dir, for TUI context)
		filename,                                                        // Bare filename in current directory
	}

	// Add binary-relative path if we found it
	if binaryDir != "" {
		locations = append(locations, filepath.Join(binaryDir, "assets", filename))
	}

	// Add system paths last (read-only fallback)
	locations = append(locations,
		filepath.Join("/usr/local/share/sysc-Go/assets", filename), // Local install
		filepath.Join("/usr/share/sysc-Go/assets", filename),       // System install
	)

	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// readTextFile reads text from a file with fallback to SYSC.txt
func readTextFile(file string) string {
	if file != "" {
		// Try to read from provided file
		data, readErr := os.ReadFile(file)
		if readErr == nil {
			return string(data)
		}

		// Fall back to SYSC.txt
		fallbackPath := findAssetFile("SYSC.txt")
		if fallbackPath != "" {
			data, readErr = os.ReadFile(fallbackPath)
			if readErr == nil {
				fmt.Printf("Warning: Could not read %s, using %s\n", file, fallbackPath)
				time.Sleep(1 * time.Second)
				return string(data)
			}
		}

		fmt.Printf("Error: Could not read file %s and could not find fallback SYSC.txt\n", file)
		os.Exit(1)
	}

	// No file provided, use SYSC.txt
	fallbackPath := findAssetFile("SYSC.txt")
	if fallbackPath != "" {
		data, readErr := os.ReadFile(fallbackPath)
		if readErr == nil {
			return string(data)
		}
		fmt.Printf("Error: Could not read SYSC.txt from %s\n", fallbackPath)
		os.Exit(1)
	}

	fmt.Println("Error: Could not find SYSC.txt in any asset location")
	os.Exit(1)
	return ""
}

func wrapText(text string, width int) string {
	if width <= 0 {
		width = 80
	}

	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		// If line is empty, keep it
		if strings.TrimSpace(line) == "" {
			wrappedLines = append(wrappedLines, "")
			continue
		}

		// If line fits, keep it
		if len(line) <= width {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		// Wrap long lines
		words := strings.Fields(line)
		currentLine := ""

		for _, word := range words {
			// If word itself is longer than width, break it
			if len(word) > width {
				if currentLine != "" {
					wrappedLines = append(wrappedLines, currentLine)
					currentLine = ""
				}
				// Split long word
				for len(word) > width {
					wrappedLines = append(wrappedLines, word[:width])
					word = word[width:]
				}
				currentLine = word
				continue
			}

			// Try adding word to current line
			testLine := currentLine
			if testLine != "" {
				testLine += " "
			}
			testLine += word

			if len(testLine) <= width {
				currentLine = testLine
			} else {
				// Start new line with this word
				if currentLine != "" {
					wrappedLines = append(wrappedLines, currentLine)
				}
				currentLine = word
			}
		}

		// Add remaining line
		if currentLine != "" {
			wrappedLines = append(wrappedLines, currentLine)
		}
	}

	return strings.Join(wrappedLines, "\n")
}

// setupKeyboardInterrupt sets up signal handling for Ctrl+C
// Returns a channel that will receive true when user wants to exit
func setupKeyboardInterrupt() chan bool {
	quit := make(chan bool, 1)

	// Use signal handling instead of raw mode to avoid breaking output formatting
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer signal.Stop(sigChan) // Cleanup signal handler
		select {
		case <-sigChan:
			quit <- true
		case <-quit:
			// Parent exited normally, cleanup
			return
		}
	}()

	return quit
}

func showHelp() {
	fmt.Print(banner)
	fmt.Println("Usage: syscgo [options]")
	fmt.Println()
	fmt.Println("Quick Start:")
	fmt.Println("  syscgo                          # Fire effect with default theme")
	fmt.Println("  syscgo -effect matrix           # Matrix rain")
	fmt.Println("  syscgo-tui                      # Interactive browser (recommended)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -effect   string   Animation effect (default: fire)")
	fmt.Println("  -theme    string   Color theme (default: dracula)")
	fmt.Println("  -duration int      Duration in seconds, 0=infinite (default: 10)")
	fmt.Println("  -file     string   Text file for text-based effects")
	fmt.Println("  -auto              Auto-size canvas (beam-text only)")
	fmt.Println("  -display           Hold at final state (beam-text only)")
	fmt.Println()
	fmt.Println("Effects:")
	fmt.Println("  fire, fire-text, matrix, matrix-art, rain, rain-art, fireworks")
	fmt.Println("  pour, print, beams, beam-text, ring-text, blackhole, aquarium, sonar")
	fmt.Println()
	fmt.Println("Themes:")
	fmt.Println("  dracula, gruvbox, nord, tokyo-night, catppuccin, material")
	fmt.Println("  solarized, monochrome, transishardjob, rama, eldritch, dark")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  syscgo -effect fire -theme nord -duration 30")
	fmt.Println("  syscgo -effect fire-text -file SYSC.txt -theme dracula -duration 0")
	fmt.Println("  syscgo -effect aquarium -theme dracula -duration 0")
	fmt.Println("  syscgo -effect beam-text -file art.txt -auto -display -theme nord")
	fmt.Println()
	fmt.Println("For more info: https://github.com/Nomadcxx/sysc-Go")
}

func main() {
	effect := flag.String("effect", "fire", "Animation effect (fire, matrix, rain, fireworks, decrypt)")
	theme := flag.String("theme", "dracula", "Color theme")
	duration := flag.Int("duration", 10, "Duration in seconds (0 = infinite)")
	file := flag.String("file", "", "Text file for text-based effects (decrypt, pour, print, beam-text)")
	auto := flag.Bool("auto", false, "Auto-size canvas to fit text (beam-text only)")
	display := flag.Bool("display", false, "Display mode: complete once and hold (beam-text only)")
	help := flag.Bool("h", false, "Show help")
	flag.BoolVar(help, "help", false, "Show help")
	showVersion := flag.Bool("version", false, "Show version")

	flag.Usage = showHelp
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *showVersion {
		fmt.Printf("syscgo version %s\n", version)
		return
	}

	// Get terminal size
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 80, 24
	}

	// Setup terminal
	fmt.Print("\033[2J\033[H")   // Clear screen
	fmt.Print("\033[?25l")       // Hide cursor
	defer fmt.Print("\033[?25h") // Show cursor on exit

	// Calculate frame count (0 = infinite)
	frames := 0
	if *duration > 0 {
		frames = *duration * 20 // 20 fps
	}

	switch *effect {
	case "fire":
		runFire(width, height, *theme, frames)
	case "fire-text":
		runFireText(width, height, *theme, *file, frames)
	case "matrix":
		runMatrix(width, height, *theme, frames)
	case "matrix-art":
		runMatrixArt(width, height, *theme, *file, frames)
	case "fireworks":
		runFireworks(width, height, *theme, frames)
	case "rain":
		runRain(width, height, *theme, frames)
	case "rain-art":
		runRainArt(width, height, *theme, *file, frames)
	case "pour":
		runPour(width, height, *theme, *file, frames)
	case "print":
		runPrint(width, height, *theme, *file, frames)
	case "beams":
		runBeams(width, height, *theme, frames)
	case "beam-text":
		runBeamText(width, height, *theme, *file, *auto, *display, frames)
	case "ring-text":
		runRingText(width, height, *theme, *file, frames)
	case "blackhole":
		runBlackhole(width, height, *theme, *file, frames)
	// WIP: blackhole-particles is currently broken (terminal scrolling issue)
	// case "blackhole-particles":
	// 	runBlackhole(width, height, *theme, "", frames)
	case "aquarium":
		runAquarium(width, height, *theme, frames)
	case "skull":
		runSkull(width, height, *theme, frames)
	case "sonar":
		runSonar(width, height, *theme, frames)
	default:
		fmt.Printf("Unknown effect: %s\n", *effect)
		fmt.Println("Available: fire, fire-text, matrix, rain, rain-art, fireworks, pour, print, beams, beam-text, ring-text, blackhole, aquarium, sonar")
		os.Exit(1)
	}
}

func runFire(width, height int, theme string, frames int) {
	palette := animations.GetFirePalette(theme)
	fire := animations.NewFireEffect(width, height, palette)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		fire.Update()
		output := fire.Render()

		fmt.Print("\033[H") // Move cursor to top
		fmt.Print(output)
		os.Stdout.Sync() // Flush output buffer immediately
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runFireText(width, height int, theme string, file string, frames int) {
	// Get theme palette for fire
	palette := animations.GetFirePalette(theme)

	// Read text from file or use default SYSC.txt
	text := readTextFile(file)

	// Create fire-text effect
	fireText := animations.NewFireTextEffect(width, height, palette, text)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		fireText.Update()
		output := fireText.Render()

		fmt.Print("\033[H") // Move cursor to top
		fmt.Print(output)
		os.Stdout.Sync() // Flush output buffer immediately
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runMatrix(width, height int, theme string, frames int) {
	palette := animations.GetMatrixPalette(theme)
	matrix := animations.NewMatrixEffect(width, height, palette)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		matrix.Update()
		output := matrix.Render()

		fmt.Print("\033[H") // Move cursor to top
		fmt.Print(output)
		os.Stdout.Sync() // Flush output buffer immediately
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runMatrixArt(width, height int, theme string, file string, frames int) {
	// Get theme palette for matrix
	palette := animations.GetMatrixPalette(theme)

	// Read text from file or use default SYSC.txt
	text := readTextFile(file)

	// Create matrix-art effect
	matrixArt := animations.NewMatrixArtEffect(width, height, palette, text)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		matrixArt.Update()
		output := matrixArt.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		os.Stdout.Sync() // Flush output buffer immediately
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runFireworks(width, height int, theme string, frames int) {
	palette := animations.GetFireworksPalette(theme)
	fireworks := animations.NewFireworksEffect(width, height, palette)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		fireworks.Update()
		output := fireworks.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		os.Stdout.Sync() // Flush output buffer immediately
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runRain(width, height int, theme string, frames int) {
	palette := animations.GetRainPalette(theme)
	rain := animations.NewRainEffect(width, height, palette)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		rain.Update()
		output := rain.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		os.Stdout.Sync() // Flush output buffer immediately
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runRainArt(width, height int, theme string, file string, frames int) {
	// Get theme palette for rain
	palette := animations.GetRainPalette(theme)

	// Read text from file or use default SYSC.txt
	text := readTextFile(file)

	// Create rain-art effect
	rainArt := animations.NewRainArtEffect(width, height, palette, text)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		rainArt.Update()
		output := rainArt.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		os.Stdout.Sync() // Flush output buffer immediately
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runPour(width, height int, theme string, file string, frames int) {
	// Get theme colors for pour effect
	var gradientStops []string

	switch theme {
	case "dracula":
		gradientStops = []string{"#ff79c6", "#bd93f9", "#ffffff"}
	case "gruvbox":
		gradientStops = []string{"#fe8019", "#fabd2f", "#ffffff"}
	case "nord":
		gradientStops = []string{"#88c0d0", "#81a1c1", "#ffffff"}
	case "tokyo-night":
		gradientStops = []string{"#9ece6a", "#e0af68", "#ffffff"}
	case "catppuccin":
		gradientStops = []string{"#cba6f7", "#f5c2e7", "#ffffff"}
	case "material":
		gradientStops = []string{"#03dac6", "#bb86fc", "#ffffff"}
	case "solarized":
		gradientStops = []string{"#268bd2", "#2aa198", "#ffffff"}
	case "monochrome":
		gradientStops = []string{"#808080", "#c0c0c0", "#ffffff"}
	case "transishardjob":
		gradientStops = []string{"#55cdfc", "#f7a8b8", "#ffffff"}
	case "rama":
		gradientStops = []string{"#ef233c", "#d90429", "#edf2f4"}
	case "eldritch":
		gradientStops = []string{"#37f499", "#04d1f9", "#ebfafa"}
	case "dark":
		gradientStops = []string{"#ffffff", "#cccccc", "#ffffff"}
	default:
		gradientStops = []string{"#8A008A", "#00D1FF", "#FFFFFF"}
	}

	// Read text from file or use default
	text := "POUR EFFECT\nDEMO TEXT\nTHIRD LINE"
	if file != "" {
		data, err := os.ReadFile(file)
		if err == nil {
			text = string(data)
		}
	}

	// Don't wrap text - ASCII art needs to be preserved as-is
	// The pour effect will handle centering

	// Create pour effect with sample text centered in terminal
	config := animations.PourConfig{
		Width:                  width,
		Height:                 height,
		Text:                   text,
		PourDirection:          "down",
		PourSpeed:              3,
		MovementSpeed:          0.2,
		EasingFunction:         "easeIn", // Options: "easeIn", "easeOut", "easeInOut"
		Gap:                    1,
		StartingColor:          "#ffffff",
		FinalGradientStops:     gradientStops,
		FinalGradientSteps:     12,
		FinalGradientFrames:    5,
		FinalGradientDirection: "horizontal",
		Auto:                   false, // CLI uses full terminal width/height
		Display:                false, // CLI loops continuously
		HoldFrames:             100,   // ~5 seconds at 20fps
	}

	pour := animations.NewPourEffect(config)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		pour.Update()
		output := pour.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		os.Stdout.Sync() // Flush output buffer immediately
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runPrint(width, height int, theme string, file string, frames int) {
	// Get theme colors for print effect
	var gradientStops []string

	switch theme {
	case "dracula":
		gradientStops = []string{"#ff79c6", "#bd93f9", "#8be9fd"}
	case "gruvbox":
		gradientStops = []string{"#fe8019", "#fabd2f", "#b8bb26"}
	case "nord":
		gradientStops = []string{"#88c0d0", "#81a1c1", "#5e81ac"}
	case "tokyo-night":
		gradientStops = []string{"#9ece6a", "#e0af68", "#bb9af7"}
	case "catppuccin":
		gradientStops = []string{"#cba6f7", "#f5c2e7", "#f5e0dc"}
	case "material":
		gradientStops = []string{"#03dac6", "#bb86fc", "#cf6679"}
	case "solarized":
		gradientStops = []string{"#268bd2", "#2aa198", "#859900"}
	case "monochrome":
		gradientStops = []string{"#808080", "#c0c0c0", "#ffffff"}
	case "transishardjob":
		gradientStops = []string{"#55cdfc", "#f7a8b8", "#ffffff"}
	case "rama":
		gradientStops = []string{"#ef233c", "#d90429", "#edf2f4"}
	case "eldritch":
		gradientStops = []string{"#37f499", "#04d1f9", "#ebfafa"}
	case "dark":
		gradientStops = []string{"#ffffff", "#cccccc", "#ffffff"}
	default:
		gradientStops = []string{"#8A008A", "#00D1FF", "#FFFFFF"}
	}

	// Read text from file or use default
	text := "PRINT EFFECT\nDEMO TEXT\nTHIRD LINE"
	if file != "" {
		data, err := os.ReadFile(file)
		if err == nil {
			text = string(data)
		}
	}

	// Don't wrap text - ASCII art needs to be preserved as-is
	// The print effect will handle centering

	// Create print effect configuration
	config := animations.PrintConfig{
		Width:           width,
		Height:          height,
		Text:            text,
		FramesPerChar:   1, // Print every frame for smooth animation
		PrintSpeed:      2, // 2 characters per update
		PrintHeadSymbol: "█",
		TrailSymbols:    []string{"░", "▒", "▓"},
		GradientStops:   gradientStops,
		Auto:            false, // CLI uses full terminal width/height
		Display:         false, // CLI loops continuously
		HoldFrames:      100,   // ~5 seconds at 20fps
	}

	print := animations.NewPrintEffect(config)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		print.Update()
		output := print.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		os.Stdout.Sync() // Flush output buffer immediately
		time.Sleep(30 * time.Millisecond)
		frame++
	}
}

func runBeams(width, height int, theme string, frames int) {
	// Get theme colors for beams background effect
	var beamGradientStops []string
	var finalGradientStops []string

	switch theme {
	case "dracula":
		beamGradientStops = []string{"#ffffff", "#8be9fd", "#bd93f9"}
		finalGradientStops = []string{"#6272a4", "#bd93f9", "#f8f8f2"}
	case "gruvbox":
		beamGradientStops = []string{"#ffffff", "#fabd2f", "#fe8019"}
		finalGradientStops = []string{"#504945", "#fabd2f", "#ebdbb2"}
	case "nord":
		beamGradientStops = []string{"#ffffff", "#88c0d0", "#81a1c1"}
		finalGradientStops = []string{"#434c5e", "#88c0d0", "#eceff4"}
	case "tokyo-night":
		beamGradientStops = []string{"#ffffff", "#7dcfff", "#bb9af7"}
		finalGradientStops = []string{"#414868", "#7aa2f7", "#c0caf5"}
	case "catppuccin":
		beamGradientStops = []string{"#ffffff", "#89dceb", "#cba6f7"}
		finalGradientStops = []string{"#45475a", "#cba6f7", "#cdd6f4"}
	case "material":
		beamGradientStops = []string{"#ffffff", "#89ddff", "#bb86fc"}
		finalGradientStops = []string{"#546e7a", "#89ddff", "#eceff1"}
	case "solarized":
		beamGradientStops = []string{"#ffffff", "#2aa198", "#268bd2"}
		finalGradientStops = []string{"#586e75", "#2aa198", "#fdf6e3"}
	case "monochrome":
		beamGradientStops = []string{"#ffffff", "#c0c0c0", "#808080"}
		finalGradientStops = []string{"#3a3a3a", "#9a9a9a", "#ffffff"}
	case "transishardjob":
		beamGradientStops = []string{"#ffffff", "#55cdfc", "#f7a8b8"}
		finalGradientStops = []string{"#55cdfc", "#f7a8b8", "#ffffff"}
	case "rama":
		beamGradientStops = []string{"#ffffff", "#ef233c", "#d90429"}
		finalGradientStops = []string{"#8d99ae", "#ef233c", "#edf2f4"}
	case "eldritch":
		beamGradientStops = []string{"#ffffff", "#37f499", "#04d1f9"}
		finalGradientStops = []string{"#7081d0", "#37f499", "#ebfafa"}
	case "dark":
		beamGradientStops = []string{"#ffffff", "#cccccc", "#999999"}
		finalGradientStops = []string{"#333333", "#ffffff", "#ffffff"}
	default:
		beamGradientStops = []string{"#ffffff", "#00D1FF", "#8A008A"}
		finalGradientStops = []string{"#4A4A4A", "#00D1FF", "#FFFFFF"}
	}

	// Create beams background effect configuration
	config := animations.BeamsConfig{
		Width:                width,
		Height:               height,
		BeamRowSymbols:       []rune{'▂', '▁', '_'},
		BeamColumnSymbols:    []rune{'▌', '▍', '▎', '▏'},
		BeamDelay:            2,
		BeamRowSpeedRange:    [2]int{20, 80},
		BeamColumnSpeedRange: [2]int{15, 30},
		BeamGradientStops:    beamGradientStops,
		BeamGradientSteps:    5,
		BeamGradientFrames:   1,
		FinalGradientStops:   finalGradientStops,
		FinalGradientSteps:   8,
		FinalGradientFrames:  1,
		FinalWipeSpeed:       3,
	}

	beams := animations.NewBeamsEffect(config)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		beams.Update()
		output := beams.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		os.Stdout.Sync() // Flush output buffer immediately
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runBeamText(width, height int, theme string, file string, auto bool, display bool, frames int) {
	// Get theme colors for beam text effect
	var beamGradientStops []string
	var finalGradientStops []string

	switch theme {
	case "dracula":
		beamGradientStops = []string{"#ffffff", "#8be9fd", "#bd93f9"}
		finalGradientStops = []string{"#6272a4", "#bd93f9", "#f8f8f2"}
	case "gruvbox":
		beamGradientStops = []string{"#ffffff", "#fabd2f", "#fe8019"}
		finalGradientStops = []string{"#504945", "#fabd2f", "#ebdbb2"}
	case "nord":
		beamGradientStops = []string{"#ffffff", "#88c0d0", "#81a1c1"}
		finalGradientStops = []string{"#434c5e", "#88c0d0", "#eceff4"}
	case "tokyo-night":
		beamGradientStops = []string{"#ffffff", "#7dcfff", "#bb9af7"}
		finalGradientStops = []string{"#414868", "#7aa2f7", "#c0caf5"}
	case "catppuccin":
		beamGradientStops = []string{"#ffffff", "#89dceb", "#cba6f7"}
		finalGradientStops = []string{"#45475a", "#cba6f7", "#cdd6f4"}
	case "material":
		beamGradientStops = []string{"#ffffff", "#89ddff", "#bb86fc"}
		finalGradientStops = []string{"#546e7a", "#89ddff", "#eceff1"}
	case "solarized":
		beamGradientStops = []string{"#ffffff", "#2aa198", "#268bd2"}
		finalGradientStops = []string{"#586e75", "#2aa198", "#fdf6e3"}
	case "monochrome":
		beamGradientStops = []string{"#ffffff", "#c0c0c0", "#808080"}
		finalGradientStops = []string{"#3a3a3a", "#9a9a9a", "#ffffff"}
	case "transishardjob":
		beamGradientStops = []string{"#ffffff", "#55cdfc", "#f7a8b8"}
		finalGradientStops = []string{"#55cdfc", "#f7a8b8", "#ffffff"}
	case "rama":
		beamGradientStops = []string{"#ffffff", "#ef233c", "#d90429"}
		finalGradientStops = []string{"#8d99ae", "#ef233c", "#edf2f4"}
	case "eldritch":
		beamGradientStops = []string{"#ffffff", "#37f499", "#04d1f9"}
		finalGradientStops = []string{"#7081d0", "#37f499", "#ebfafa"}
	case "dark":
		beamGradientStops = []string{"#ffffff", "#cccccc", "#999999"}
		finalGradientStops = []string{"#333333", "#ffffff", "#ffffff"}
	default:
		beamGradientStops = []string{"#ffffff", "#00D1FF", "#8A008A"}
		finalGradientStops = []string{"#4A4A4A", "#00D1FF", "#FFFFFF"}
	}

	// Read text from file
	text := readTextFile(file)
	if text == "" {
		fmt.Println("beam-text effect requires -file flag")
		os.Exit(1)
	}

	// Don't wrap text - ASCII art needs to be preserved as-is
	// The beam-text effect will handle sizing based on auto flag

	// Create beam text effect configuration
	config := animations.BeamTextConfig{
		Width:                width,
		Height:               height,
		Text:                 text,
		Auto:                 auto,
		Display:              display,
		BeamRowSymbols:       []rune{'▂', '▁', '_'},
		BeamColumnSymbols:    []rune{'▌', '▍', '▎', '▏'},
		BeamDelay:            2,
		BeamRowSpeedRange:    [2]int{20, 80},
		BeamColumnSpeedRange: [2]int{15, 30},
		BeamGradientStops:    beamGradientStops,
		BeamGradientSteps:    5,
		BeamGradientFrames:   1,
		FinalGradientStops:   finalGradientStops,
		FinalGradientSteps:   8,
		FinalGradientFrames:  1,
		FinalWipeSpeed:       3,
	}

	beamText := animations.NewBeamTextEffect(config)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	// When display mode is enabled, ignore duration and run until completion
	// This allows the multi-phase beam-text animation to reach its final "hold" state
	effectiveFrames := frames
	if display {
		effectiveFrames = 0
	}

	frame := 0
	for effectiveFrames == 0 || frame < effectiveFrames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		beamText.Update()
		output := beamText.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		os.Stdout.Sync() // Flush output buffer immediately
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runRingText(width, height int, theme string, file string, frames int) {
	// Get theme colors for ring text effect
	var ringColors []string
	var finalGradientStops []string

	switch theme {
	case "dracula":
		ringColors = []string{"#bd93f9", "#ff79c6", "#f1fa8c", "#8be9fd", "#50fa7b", "#ffb86c"}
		finalGradientStops = []string{"#6272a4", "#bd93f9", "#f8f8f2"}
	case "gruvbox":
		ringColors = []string{"#fabd2f", "#fe8019", "#b8bb26", "#83a598", "#d3869b", "#fb4934"}
		finalGradientStops = []string{"#504945", "#fabd2f", "#ebdbb2"}
	case "nord":
		ringColors = []string{"#88c0d0", "#81a1c1", "#5e81ac", "#8fbcbb", "#b48ead", "#a3be8c"}
		finalGradientStops = []string{"#434c5e", "#88c0d0", "#eceff4"}
	case "tokyo-night":
		ringColors = []string{"#7dcfff", "#bb9af7", "#9ece6a", "#7aa2f7", "#ff9e64", "#f7768e"}
		finalGradientStops = []string{"#414868", "#7aa2f7", "#c0caf5"}
	case "catppuccin":
		ringColors = []string{"#cba6f7", "#f5c2e7", "#a6e3a1", "#89b4fa", "#f38ba8", "#fab387"}
		finalGradientStops = []string{"#45475a", "#cba6f7", "#cdd6f4"}
	case "material":
		ringColors = []string{"#bb86fc", "#03dac6", "#cf6679", "#89ddff", "#ffcb6b", "#c3e88d"}
		finalGradientStops = []string{"#546e7a", "#89ddff", "#eceff1"}
	case "solarized":
		ringColors = []string{"#268bd2", "#2aa198", "#859900", "#cb4b16", "#d33682", "#6c71c4"}
		finalGradientStops = []string{"#586e75", "#2aa198", "#fdf6e3"}
	case "monochrome":
		ringColors = []string{"#ffffff", "#e0e0e0", "#c0c0c0", "#a0a0a0", "#808080", "#606060"}
		finalGradientStops = []string{"#3a3a3a", "#9a9a9a", "#ffffff"}
	case "transishardjob":
		ringColors = []string{"#55cdfc", "#f7a8b8", "#ffffff", "#f7a8b8", "#55cdfc", "#ffffff"}
		finalGradientStops = []string{"#55cdfc", "#f7a8b8", "#ffffff"}
	case "rama":
		ringColors = []string{"#ef233c", "#d90429", "#8d99ae", "#edf2f4", "#ef233c", "#d90429"}
		finalGradientStops = []string{"#8d99ae", "#ef233c", "#edf2f4"}
	case "eldritch":
		ringColors = []string{"#37f499", "#04d1f9", "#a48cf2", "#f265b5", "#f16c75", "#f7c67f"}
		finalGradientStops = []string{"#7081d0", "#37f499", "#ebfafa"}
	case "dark":
		ringColors = []string{"#ffffff", "#cccccc", "#999999", "#666666", "#999999", "#ffffff"}
		finalGradientStops = []string{"#333333", "#ffffff", "#ffffff"}
	default:
		ringColors = []string{"#bd93f9", "#ff79c6", "#f1fa8c", "#8be9fd", "#50fa7b", "#ffb86c"}
		finalGradientStops = []string{"#4A4A4A", "#00D1FF", "#FFFFFF"}
	}

	// Read text from file or use default SYSC.txt
	text := readTextFile(file)

	// Create ring text effect configuration (TTE-like parameters with theme-sensitive gradients)
	config := animations.RingTextConfig{
		Width:               width,
		Height:              height,
		Text:                text,
		RingColors:          ringColors,
		RingGap:             0.1,                      // Like TTE default
		SpinSpeedRange:      [2]float64{0.025, 0.075}, // Min-max range like TTE (0.25-1.0 mapped to radians)
		SpinDuration:        200,                      // Frames per spin rotation
		DisperseDuration:    200,                      // Frames in dispersed state
		SpinDisperseCycles:  3,                        // 3 cycles like TTE default
		TransitionFrames:    60,                       // Transition between states (reduced for faster animation)
		StaticFrames:        30,                       // Initial static display (reduced to start ring animation sooner)
		FinalGradientStops:  finalGradientStops,
		FinalGradientSteps:  12,
		StaticGradientStops: ringColors,                    // Use ring colors for static gradient
		StaticGradientDir:   animations.GradientHorizontal, // Left-to-right gradient
	}

	ringText := animations.NewRingTextEffect(config)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		ringText.Update()
		output := ringText.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		os.Stdout.Sync() // Flush output buffer immediately
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runBlackhole(width, height int, theme string, file string, frames int) {
	// Get theme colors for blackhole effect
	var starColors []string
	var blackholeColor string

	switch theme {
	case "dracula":
		starColors = []string{"#bd93f9", "#ff79c6", "#f1fa8c", "#8be9fd", "#50fa7b", "#ffb86c"}
		blackholeColor = "#f8f8f2"
	case "gruvbox":
		starColors = []string{"#fabd2f", "#fe8019", "#b8bb26", "#83a598", "#d3869b", "#fb4934"}
		blackholeColor = "#ebdbb2"
	case "nord":
		starColors = []string{"#88c0d0", "#81a1c1", "#5e81ac", "#8fbcbb", "#b48ead", "#a3be8c"}
		blackholeColor = "#eceff4"
	case "tokyo-night":
		starColors = []string{"#7dcfff", "#bb9af7", "#9ece6a", "#7aa2f7", "#f7768e", "#e0af68"}
		blackholeColor = "#c0caf5"
	case "catppuccin":
		starColors = []string{"#cba6f7", "#f5c2e7", "#a6e3a1", "#89dceb", "#fab387", "#f38ba8"}
		blackholeColor = "#cdd6f4"
	case "material":
		starColors = []string{"#bb86fc", "#03dac6", "#cf6679", "#89ddff", "#c3e88d", "#ffcb6b"}
		blackholeColor = "#eceff1"
	case "solarized":
		starColors = []string{"#268bd2", "#2aa198", "#859900", "#cb4b16", "#6c71c4", "#b58900"}
		blackholeColor = "#fdf6e3"
	case "monochrome":
		starColors = []string{"#ffffff", "#c0c0c0", "#808080", "#9a9a9a", "#bababa", "#dadada"}
		blackholeColor = "#ffffff"
	case "transishardjob":
		starColors = []string{"#55cdfc", "#f7a8b8", "#ffffff", "#f7a8b8", "#55cdfc", "#ffffff"}
		blackholeColor = "#ffffff"
	case "rama":
		starColors = []string{"#ef233c", "#d90429", "#8d99ae", "#edf2f4", "#ef233c", "#d90429"}
		blackholeColor = "#edf2f4"
	case "eldritch":
		starColors = []string{"#37f499", "#04d1f9", "#a48cf2", "#f265b5", "#f16c75", "#f7c67f"}
		blackholeColor = "#ebfafa"
	case "dark":
		starColors = []string{"#ffffff", "#cccccc", "#999999", "#666666", "#999999", "#ffffff"}
		blackholeColor = "#ffffff"
	default:
		starColors = []string{"#ffffff", "#ffd700", "#ff6b6b", "#4ecdc4", "#95e1d3", "#f38181"}
		blackholeColor = "#ffffff"
	}

	// Read text from file
	// If file is empty string, use empty text (triggers particle generation)
	// Otherwise read from file or use default assets/SYSC.txt
	var text string

	if file == "" {
		// Empty file means generate random particles (no text)
		text = ""
	} else {
		// Try to read from provided file
		data, readErr := os.ReadFile(file)
		if readErr == nil {
			text = string(data)
		} else {
			// Fall back to assets/SYSC.txt
			data, readErr = os.ReadFile("assets/SYSC.txt")
			if readErr == nil {
				text = string(data)
				fmt.Printf("Warning: Could not read %s, using assets/SYSC.txt\n", file)
				time.Sleep(1 * time.Second)
			} else {
				fmt.Printf("Error: Could not read file %s or assets/SYSC.txt\n", file)
				os.Exit(1)
			}
		}
	}

	// Create blackhole effect configuration
	config := animations.BlackholeConfig{
		Width:               width,
		Height:              height,
		Text:                text,
		BlackholeColor:      blackholeColor,
		StarColors:          starColors,
		FinalGradientStops:  starColors, // Use same gradient as start
		FinalGradientSteps:  12,
		FinalGradientDir:    animations.GradientHorizontal, // Match start direction
		StaticGradientStops: starColors,
		StaticGradientDir:   animations.GradientHorizontal,
		FormingFrames:       10,
		ConsumingFrames:     60,
		CollapsingFrames:    50,
		ExplodingFrames:     100,
		ReturningFrames:     120,
		StaticFrames:        30,
	}

	blackhole := animations.NewBlackholeEffect(config)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		blackhole.Update()
		output := blackhole.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		os.Stdout.Sync() // Flush output buffer immediately
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runAquarium(width, height int, theme string, frames int) {
	// Theme-specific colors for aquarium
	var fishColors []string
	var waterColors []string
	var seaweedColors []string
	var bubbleColor string
	var diverColor string
	var boatColor string
	var mermaidColor string
	var anchorColor string

	switch theme {
	case "dracula":
		fishColors = []string{"#ff79c6", "#bd93f9", "#8be9fd", "#50fa7b", "#ffb86c"}
		waterColors = []string{"#6272a4", "#c2b280"}
		seaweedColors = []string{"#44475a", "#50fa7b", "#8be9fd"}
		bubbleColor = "#8be9fd"
		diverColor = "#f8f8f2"
		boatColor = "#ffb86c"
		mermaidColor = "#ff79c6"
		anchorColor = "#6272a4"
	case "gruvbox":
		fishColors = []string{"#fe8019", "#fabd2f", "#b8bb26", "#83a598", "#d3869b"}
		waterColors = []string{"#458588", "#d79921"}
		seaweedColors = []string{"#3c3836", "#98971a", "#b8bb26"}
		bubbleColor = "#83a598"
		diverColor = "#ebdbb2"
		boatColor = "#fabd2f"
		mermaidColor = "#d3869b"
		anchorColor = "#504945"
	case "nord":
		fishColors = []string{"#88c0d0", "#81a1c1", "#5e81ac", "#8fbcbb", "#b48ead"}
		waterColors = []string{"#5e81ac", "#d08770"}
		seaweedColors = []string{"#2e3440", "#a3be8c", "#8fbcbb"}
		bubbleColor = "#88c0d0"
		diverColor = "#eceff4"
		boatColor = "#d08770"
		mermaidColor = "#b48ead"
		anchorColor = "#4c566a"
	case "tokyo-night":
		fishColors = []string{"#7aa2f7", "#bb9af7", "#7dcfff", "#9ece6a", "#f7768e"}
		waterColors = []string{"#7aa2f7", "#e0af68"}
		seaweedColors = []string{"#1a1b26", "#9ece6a", "#7dcfff"}
		bubbleColor = "#7dcfff"
		diverColor = "#c0caf5"
		boatColor = "#e0af68"
		mermaidColor = "#bb9af7"
		anchorColor = "#414868"
	case "catppuccin":
		fishColors = []string{"#f5c2e7", "#cba6f7", "#89dceb", "#a6e3a1", "#fab387"}
		waterColors = []string{"#89b4fa", "#f9e2af"}
		seaweedColors = []string{"#1e1e2e", "#a6e3a1", "#94e2d5"}
		bubbleColor = "#89dceb"
		diverColor = "#cdd6f4"
		boatColor = "#fab387"
		mermaidColor = "#f5c2e7"
		anchorColor = "#45475a"
	case "material":
		fishColors = []string{"#82aaff", "#c792ea", "#89ddff", "#c3e88d", "#f78c6c"}
		waterColors = []string{"#82aaff", "#ffcb6b"}
		seaweedColors = []string{"#263238", "#c3e88d", "#89ddff"}
		bubbleColor = "#89ddff"
		diverColor = "#eceff1"
		boatColor = "#ffcb6b"
		mermaidColor = "#c792ea"
		anchorColor = "#37474f"
	case "solarized":
		fishColors = []string{"#268bd2", "#2aa198", "#859900", "#cb4b16", "#6c71c4"}
		waterColors = []string{"#268bd2", "#b58900"}
		seaweedColors = []string{"#002b36", "#859900", "#2aa198"}
		bubbleColor = "#2aa198"
		diverColor = "#fdf6e3"
		boatColor = "#cb4b16"
		mermaidColor = "#d33682"
		anchorColor = "#073642"
	case "monochrome":
		fishColors = []string{"#9a9a9a", "#bababa", "#dadada", "#c0c0c0", "#808080"}
		waterColors = []string{"#5a5a5a", "#8a8a8a"}
		seaweedColors = []string{"#1a1a1a", "#5a5a5a", "#7a7a7a"}
		bubbleColor = "#c0c0c0"
		diverColor = "#ffffff"
		boatColor = "#9a9a9a"
		mermaidColor = "#bababa"
		anchorColor = "#3a3a3a"
	case "transishardjob":
		fishColors = []string{"#55cdfc", "#f7a8b8", "#ffffff", "#f7a8b8", "#55cdfc"}
		waterColors = []string{"#55cdfc", "#f7a8b8"}
		seaweedColors = []string{"#1a1a1a", "#55cdfc", "#f7a8b8"}
		bubbleColor = "#ffffff"
		diverColor = "#ffffff"
		boatColor = "#f7a8b8"
		mermaidColor = "#f7a8b8"
		anchorColor = "#55cdfc"
	case "rama":
		fishColors = []string{"#ef233c", "#d90429", "#8d99ae", "#edf2f4", "#ef233c"}
		waterColors = []string{"#8d99ae", "#ef233c"}
		seaweedColors = []string{"#2b2d42", "#8d99ae", "#ef233c"}
		bubbleColor = "#edf2f4"
		diverColor = "#edf2f4"
		boatColor = "#ef233c"
		mermaidColor = "#d90429"
		anchorColor = "#8d99ae"
	case "eldritch":
		fishColors = []string{"#37f499", "#04d1f9", "#a48cf2", "#f265b5", "#f16c75"}
		waterColors = []string{"#7081d0", "#a48cf2"}
		seaweedColors = []string{"#212337", "#37f499", "#04d1f9"}
		bubbleColor = "#04d1f9"
		diverColor = "#ebfafa"
		boatColor = "#f7c67f"
		mermaidColor = "#f265b5"
		anchorColor = "#292e42"
	case "dark":
		fishColors = []string{"#ffffff", "#cccccc", "#999999", "#ffffff", "#cccccc"}
		waterColors = []string{"#666666", "#999999"}
		seaweedColors = []string{"#000000", "#333333", "#666666"}
		bubbleColor = "#ffffff"
		diverColor = "#ffffff"
		boatColor = "#cccccc"
		mermaidColor = "#ffffff"
		anchorColor = "#333333"
	default:
		fishColors = []string{"#00ffff", "#ff00ff", "#ffff00", "#00ff00", "#ff8000"}
		waterColors = []string{"#4a9eff", "#c2b280"}
		seaweedColors = []string{"#001a1a", "#00ff00", "#00ffff"}
		bubbleColor = "#00ffff"
		diverColor = "#ffffff"
		boatColor = "#ff8000"
		mermaidColor = "#ff00ff"
		anchorColor = "#808080"
	}

	config := animations.AquariumConfig{
		Width:         width,
		Height:        height,
		FishColors:    fishColors,
		WaterColors:   waterColors,
		SeaweedColors: seaweedColors,
		BubbleColor:   bubbleColor,
		DiverColor:    diverColor,
		BoatColor:     boatColor,
		MermaidColor:  mermaidColor,
		AnchorColor:   anchorColor,
	}

	aquarium := animations.NewAquariumEffect(config)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		// Check for user exit
		select {
		case <-quit:
			return
		default:
		}

		aquarium.Update()
		output := aquarium.Render()

		fmt.Print("[H")
		fmt.Print(output)
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runSonar(width, height int, theme string, frames int) {
	palette := animations.GetSkullPalette(theme)
	sonar := animations.NewSonarEffect(width, height, palette, theme)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		select {
		case <-quit:
			return
		default:
		}

		sonar.Update()
		output := sonar.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		os.Stdout.Sync()
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runSkull(width, height int, theme string, frames int) {
	palette := animations.GetSkullPalette(theme)
	skull := animations.NewSkullEffect(width, height, palette, theme)

	quit := setupKeyboardInterrupt()
	defer close(quit)

	frame := 0
	for frames == 0 || frame < frames {
		select {
		case <-quit:
			return
		default:
		}

		skull.Update()
		output := skull.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		os.Stdout.Sync()
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

