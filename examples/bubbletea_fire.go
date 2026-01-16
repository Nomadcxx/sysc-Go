//go:build bubbletea_fire

package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/Nomadcxx/sysc-Go/animations"
	"time"
)

type model struct {
	fire  *animations.FireEffect
	frame int
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	case tickMsg:
		m.frame++
		return m, tick()
	case tea.WindowSizeMsg:
		m.fire.Resize(msg.Width, msg.Height)
	}
	return m, nil
}

func (m model) View() string {
	m.fire.Update()
	return m.fire.Render()
}

func main() {
	palette := animations.GetFirePalette("dracula")
	fire := animations.NewFireEffect(80, 24, palette)

	p := tea.NewProgram(model{fire: fire, frame: 0}, tea.WithAltScreen())
	p.Run()
}