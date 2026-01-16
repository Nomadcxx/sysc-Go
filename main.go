package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Nomadcxx/sysc-Go/ui/floydui"
)

func main() {
	// Create the FLOYD UI model
	model := floydui.NewModel()

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithFPS(30), // For animations
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}