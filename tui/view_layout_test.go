package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// viewLines splits a Bubble Tea view into terminal rows. A trailing newline
// is a real extra row (the alt screen scrolls), so it is not trimmed.
func viewLines(view string) []string {
	if view == "" {
		return nil
	}
	return strings.Split(view, "\n")
}

// TestMainViewFitsAdvertisedMinimum locks the 100x30 floor advertised by the
// too-small warning. The help row must land inside the window; a view taller
// than the terminal scrolls the alt screen on every frame.
func TestMainViewFitsAdvertisedMinimum(t *testing.T) {
	sizes := []struct{ width, height int }{
		{100, 30},
		{100, 31},
		{100, 32},
		{100, 33},
		{100, 34},
		{100, 35},
		{120, 30},
		{120, 35},
		{120, 36},
		{160, 50},
	}

	for _, sz := range sizes {
		m := NewModel()
		m.width = sz.width
		m.height = sz.height

		view := m.View()
		lines := viewLines(view)
		if len(lines) > sz.height {
			t.Errorf("%dx%d: View() is %d lines (overflow +%d); canvas=%d selectors=%d guidance=%d help=%d",
				sz.width, sz.height, len(lines), len(lines)-sz.height,
				len(viewLines(m.renderCanvas())),
				len(viewLines(m.renderSelectors())),
				len(viewLines(m.renderGuidance())),
				len(viewLines(m.renderHelp())),
			)
		}

		helpIdx := -1
		for i, line := range lines {
			if strings.Contains(line, "Q Quit") {
				helpIdx = i
			}
			if w := lipgloss.Width(line); w > sz.width {
				t.Errorf("%dx%d: line %d is %d columns wide and will wrap", sz.width, sz.height, i, w)
			}
		}
		if helpIdx < 0 {
			t.Errorf("%dx%d: help row %q is missing", sz.width, sz.height, "Q Quit")
			continue
		}
		if helpIdx >= sz.height {
			t.Errorf("%dx%d: help row is on line %d, past the %d-row window", sz.width, sz.height, helpIdx, sz.height)
		}
	}
}

// TestMainViewAnimationFrameFitsMinimum covers the preview path. Animation
// frames are reprinted every tick; a frame taller than the window scrolls.
func TestMainViewAnimationFrameFitsMinimum(t *testing.T) {
	m := NewModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = updated.(Model)
	m.animationRunning = true
	m.currentAnim = m.createAnimation()
	if m.currentAnim == nil {
		t.Fatal("expected a preview animation")
	}

	view := m.View()
	lines := viewLines(view)
	if len(lines) > m.height {
		t.Fatalf("100x30 animation frame is %d lines (overflow +%d)", len(lines), len(lines)-m.height)
	}
	if !containsLine(lines, "ESC Stop animation") {
		t.Fatalf("stop hint is not inside the %d-row window", m.height)
	}
}

func containsLine(lines []string, needle string) bool {
	for _, line := range lines {
		if strings.Contains(line, needle) {
			return true
		}
	}
	return false
}
