package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestStaleTickIsIgnored covers Esc then Enter inside one tick: the first
// run's TickMsg is still in flight and must not re-arm a second tick loop.
func TestStaleTickIsIgnored(t *testing.T) {
	m := NewModel()
	m.width, m.height = 120, 40
	m.animations = []string{"fire"}
	m.selectedAnimation = 0

	started, firstCmd := m.startAnimation()
	if !started.animationRunning || firstCmd == nil {
		t.Fatal("animation did not start")
	}
	staleMsg := firstCmd() // the tick armed by the first run

	stopped, _ := started.Update(tea.KeyMsg{Type: tea.KeyEsc})
	restarted, secondCmd := stopped.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := restarted.(Model)

	if m2.tickGen == started.tickGen {
		t.Fatal("startAnimation did not bump tickGen")
	}
	if _, cmd := m2.Update(staleMsg); cmd != nil {
		t.Fatal("stale TickMsg re-armed the tick loop")
	}

	freshMsg := secondCmd()
	if _, cmd := m2.Update(freshMsg); cmd == nil {
		t.Fatal("fresh TickMsg did not re-arm the tick loop")
	}
}
