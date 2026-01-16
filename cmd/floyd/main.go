package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Nomadcxx/sysc-Go/ui/floydui"
)

func main() {
	p := tea.NewProgram(
		floydui.NewModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
