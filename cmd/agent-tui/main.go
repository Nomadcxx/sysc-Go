package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Nomadcxx/sysc-Go/agenttui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Check for help flag
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		printHelp()
		return
	}

	// Parse flags
	var (
		theme string
		apiKey string
	)

	for i, arg := range os.Args[1:] {
		switch {
		case arg == "--theme" && i+1 < len(os.Args)-1:
			theme = os.Args[i+2]
		case arg == "--api-key" && i+1 < len(os.Args)-1:
			apiKey = os.Args[i+2]
		case strings.HasPrefix(arg, "--theme="):
			theme = strings.TrimPrefix(arg, "--theme=")
		case strings.HasPrefix(arg, "--api-key="):
			apiKey = strings.TrimPrefix(arg, "--api-key=")
		}
	}

	// Default theme
	if theme == "" {
		theme = "catppuccin"
	}

	// Create the TUI model
	m := agenttui.NewAgentModel()
	m.SetTheme(theme)

	// Set up the initial size
	m.SetSize(80, 24)

	// Check for API key in environment if not provided
	// Note: NewProxyClient will also try to read from ~/.claude/settings.json
	if apiKey == "" {
		apiKey = os.Getenv("GLM_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("ZHIPU_API_KEY")
		}
		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_AUTH_TOKEN")
		}
	}

	// Pass API key to model if provided
	if apiKey != "" {
		// TODO: Add SetAPIKey method to model
		// For now, NewProxyClient reads from ~/.claude/settings.json
	}

	// Create the program with options
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
		tea.WithOutput(os.Stderr), // Output to stderr for potential piping
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("FLOYD Agent TUI - GLM4.7 Interactive Console")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  agent-tui [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  --theme <name>       Set the color theme (default: catppuccin)")
	fmt.Println("  --api-key <key>      Set the GLM/Zhipu API key")
	fmt.Println("  --help               Show this help message")
	fmt.Println()
	fmt.Println("ENVIRONMENT VARIABLES:")
	fmt.Println("  GLM_API_KEY          GLM API key (Zhipu AI)")
	fmt.Println("  ZHIPU_API_KEY        Alias for GLM_API_KEY")
	fmt.Println()
	fmt.Println("AVAILABLE THEMES:")
	fmt.Println("  catppuccin, dracula, gruvbox, nord, tokyo-night,")
	fmt.Println("  material, solarized, monochrome, transishardjob,")
	fmt.Println("  rama, eldritch, dark")
	fmt.Println()
	fmt.Println("KEYBINDINGS:")
	fmt.Println("  esc                  Switch to normal mode")
	fmt.Println("  i                    Switch to insert mode (from normal)")
	fmt.Println("  enter                Send message")
	fmt.Println("  ctrl+j/k             Navigate input history")
	fmt.Println("  ctrl+l               Clear conversation")
	fmt.Println("  ctrl+r               Refresh viewport")
	fmt.Println("  ctrl+s               Toggle auto-scroll")
	fmt.Println("  ctrl+c               Quit (or cancel current operation)")
	fmt.Println("  q                    Quit (from normal mode)")
}
