package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Nomadcxx/sysc-Go/animations"
	"golang.org/x/term"
)

const banner = `▄▀▀▀▀ █   █ ▄▀▀▀▀ ▄▀▀▀▀    ▄▀    ▄▀ 
 ▀▀▀▄ ▀▀▀▀█  ▀▀▀▄ █      ▄▀    ▄▀   
▀▀▀▀  ▀▀▀▀▀ ▀▀▀▀   ▀▀▀▀ ▀     ▀

Terminal Animation Library
`

// wrapText wraps text to fit within the specified width
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

func showHelp() {
	fmt.Print(banner)
	fmt.Println("Usage: syscgo [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -effect string")
	fmt.Println("        Animation effect (default: fire)")
	fmt.Println("        Available: fire, matrix, rain, fireworks, decrypt, pour, print, beams, aquarium")
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
	fmt.Println("  -file string")
	fmt.Println("        Text file for text-based effects (decrypt, pour, print, beams)")
	fmt.Println("        If omitted with beams effect, runs as full-screen background animation")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  syscgo -effect fire -theme dracula")
	fmt.Println("  syscgo -effect matrix -theme nord -duration 30")
	fmt.Println("  syscgo -effect fireworks -theme gruvbox -duration 0")
	fmt.Println("  syscgo -effect decrypt -theme tokyo-night -file message.txt -duration 15")
	fmt.Println("  syscgo -effect pour -theme catppuccin -duration 10")
	fmt.Println("  syscgo -effect print -theme dracula -duration 15")
	fmt.Println("  syscgo -effect beams -theme nord -duration 0")
	fmt.Println("  syscgo -effect beams -theme nord -file message.txt -duration 20")
	fmt.Println("  syscgo -effect aquarium -theme nord -duration 0")
	fmt.Println()
}

func main() {
	effect := flag.String("effect", "fire", "Animation effect (fire, matrix, rain, fireworks, decrypt)")
	theme := flag.String("theme", "dracula", "Color theme")
	duration := flag.Int("duration", 10, "Duration in seconds (0 = infinite)")
	file := flag.String("file", "", "Text file for text-based effects (decrypt, pour, print, beams)")
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
	case "matrix":
		runMatrix(width, height, *theme, frames)
	case "fireworks":
		runFireworks(width, height, *theme, frames)
	case "rain":
		runRain(width, height, *theme, frames)
	case "decrypt":
		runDecrypt(width, height, *theme, *file, frames)
	case "pour":
		runPour(width, height, *theme, *file, frames)
	case "print":
		runPrint(width, height, *theme, *file, frames)
	case "beams":
		runBeams(width, height, *theme, *file, frames)
	case "aquarium":
		runAquarium(width, height, *theme, frames)
	default:
		fmt.Printf("Unknown effect: %s\n", *effect)
		fmt.Println("Available: fire, matrix, rain, fireworks, decrypt, pour, print, beams, aquarium")
		os.Exit(1)
	}
}

func runFire(width, height int, theme string, frames int) {
	palette := animations.GetFirePalette(theme)
	fire := animations.NewFireEffect(width, height, palette)

	frame := 0
	for frames == 0 || frame < frames {
		fire.Update()
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
		matrix.Update()
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
		fireworks.Update()
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
		rain.Update()
		output := rain.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
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
	
	// Wrap text to fit terminal width (leave margin for centering)
	text = wrapText(text, width-10)
	
	// Create pour effect with sample text centered in terminal
	config := animations.PourConfig{
		Width:                  width,
		Height:                 height,
		Text:                   text,
		PourDirection:          "down",
		PourSpeed:              3,
		MovementSpeed:          0.2,
		Gap:                    1,
		StartingColor:          "#ffffff",
		FinalGradientStops:     gradientStops,
		FinalGradientSteps:     12,
		FinalGradientFrames:    5,
		FinalGradientDirection: "horizontal",
	}
	
	pour := animations.NewPourEffect(config)

	frame := 0
	for frames == 0 || frame < frames {
		pour.Update()
		output := pour.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
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
	
	// Wrap text to fit terminal width (leave margin for centering)
	text = wrapText(text, width-10)
	
	// Create print effect configuration
	config := animations.PrintConfig{
		Width:           width,
		Height:          height,
		Text:            text,
		CharDelay:       30 * time.Millisecond,
		PrintSpeed:      2,
		PrintHeadSymbol: "█",
		TrailSymbols:    []string{"░", "▒", "▓"},
		GradientStops:   gradientStops,
	}
	
	print := animations.NewPrintEffect(config)

	frame := 0
	for frames == 0 || frame < frames {
		print.Update()
		output := print.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		time.Sleep(30 * time.Millisecond)
		frame++
	}
}

func runBeams(width, height int, theme string, file string, frames int) {
	// Get theme colors for beams effect
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
	default:
		beamGradientStops = []string{"#ffffff", "#00D1FF", "#8A008A"}
		finalGradientStops = []string{"#4A4A4A", "#00D1FF", "#FFFFFF"}
	}

	// Read text from file if provided (empty = background mode)
	text := ""
	if file != "" {
		data, err := os.ReadFile(file)
		if err == nil {
			// Wrap text to fit terminal width (leave margin for centering)
			text = wrapText(string(data), width-10)
		}
	}

	// Create beams effect configuration
	config := animations.BeamsConfig{
		Width:                width,
		Height:               height,
		Text:                 text,
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

	frame := 0
	for frames == 0 || frame < frames {
		beams.Update()
		output := beams.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}

func runDecrypt(width, height int, theme string, file string, frames int) {
	// Get theme colors for decrypt effect
	var ciphertextColors []string
	var gradientStops []string

	switch theme {
	case "dracula":
		ciphertextColors = []string{"#008000", "#00cb00", "#00ff00"}
		gradientStops = []string{"#ff79c6"}
	case "gruvbox":
		ciphertextColors = []string{"#008000", "#00cb00", "#00ff00"}
		gradientStops = []string{"#fe8019"}
	case "nord":
		ciphertextColors = []string{"#008000", "#00cb00", "#00ff00"}
		gradientStops = []string{"#88c0d0"}
	case "tokyo-night":
		ciphertextColors = []string{"#008000", "#00cb00", "#00ff00"}
		gradientStops = []string{"#9ece6a"}
	case "catppuccin":
		ciphertextColors = []string{"#008000", "#00cb00", "#00ff00"}
		gradientStops = []string{"#cba6f7"}
	case "material":
		ciphertextColors = []string{"#008000", "#00cb00", "#00ff00"}
		gradientStops = []string{"#03dac6"}
	case "solarized":
		ciphertextColors = []string{"#008000", "#00cb00", "#00ff00"}
		gradientStops = []string{"#268bd2"}
	case "monochrome":
		ciphertextColors = []string{"#808080", "#a0a0a0", "#c0c0c0"}
		gradientStops = []string{"#ffffff"}
	case "transishardjob":
		ciphertextColors = []string{"#008000", "#00cb00", "#00ff00"}
		gradientStops = []string{"#55cdfc"}
	default:
		ciphertextColors = []string{"#008000", "#00cb00", "#00ff00"}
		gradientStops = []string{"#eda000"}
	}

	// Read text from file or use default
	text := "DECRYPT ME"
	if file != "" {
		data, err := os.ReadFile(file)
		if err == nil {
			text = string(data)
		}
	}
	
	// Wrap text to fit terminal width (leave margin for centering)
	text = wrapText(text, width-10)

	// Create decrypt effect with sample text centered in terminal
	config := animations.DecryptConfig{
		Width:                  width,
		Height:                 height,
		Text:                   text,
		Palette:                []string{}, // Not used in decrypt effect
		TypingSpeed:            2,          // Slower for better visibility
		CiphertextColors:       ciphertextColors,
		FinalGradientStops:     gradientStops,
		FinalGradientSteps:     12,
		FinalGradientDirection: "vertical",
	}

	decrypt := animations.NewDecryptEffect(config)

	frame := 0
	for frames == 0 || frame < frames {
		decrypt.Update()
		output := decrypt.Render()

		fmt.Print("\033[H")
		fmt.Print(output)
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

	frame := 0
	for frames == 0 || frame < frames {
		aquarium.Update()
		output := aquarium.Render()

		fmt.Print("[H")
		fmt.Print(output)
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}
