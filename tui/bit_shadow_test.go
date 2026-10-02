package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func pressBit(m Model, k tea.KeyType) Model {
	updated, _ := m.handleBitEditorKeyPress(tea.KeyMsg{Type: k})
	return updated.(Model)
}

func TestBitShadowDownIncreasesOffsetY(t *testing.T) {
	m := Model{
		bitEditorMode:     true,
		bitFocusedControl: 5,
		bitShadow:         true,
		bitShadowOffsetY:  1,
	}

	m = pressBit(m, tea.KeyUp)
	if !m.bitShadow || m.bitShadowOffsetY != 0 {
		t.Fatalf("up = shadow %v Y %d, want on and 0", m.bitShadow, m.bitShadowOffsetY)
	}

	m = pressBit(m, tea.KeyUp)
	if !m.bitShadow || m.bitShadowOffsetY != -1 {
		t.Fatalf("second up = shadow %v Y %d, want on and -1", m.bitShadow, m.bitShadowOffsetY)
	}

	m = pressBit(m, tea.KeyDown)
	if !m.bitShadow {
		t.Fatal("down toggled shadow off")
	}
	if m.bitShadowOffsetY != 0 {
		t.Fatalf("down offset Y = %d, want 0", m.bitShadowOffsetY)
	}

	m.bitShadowOffsetY = 5
	m = pressBit(m, tea.KeyDown)
	if !m.bitShadow || m.bitShadowOffsetY != 5 {
		t.Fatalf("down at max = shadow %v Y %d, want on and 5", m.bitShadow, m.bitShadowOffsetY)
	}
}

func TestBitShadowEnterTogglesWithoutChangingY(t *testing.T) {
	m := Model{
		bitEditorMode:     true,
		bitFocusedControl: 5,
		bitShadow:         false,
		bitShadowOffsetY:  -2,
	}

	m = pressBit(m, tea.KeyEnter)
	if !m.bitShadow {
		t.Fatal("enter should turn shadow on")
	}
	if m.bitShadowOffsetY != -2 {
		t.Fatalf("enter changed Y to %d", m.bitShadowOffsetY)
	}

	m = pressBit(m, tea.KeyDown)
	if m.bitShadowOffsetY != -1 {
		t.Fatalf("down after enabling = %d, want -1", m.bitShadowOffsetY)
	}

	off := Model{
		bitEditorMode:     true,
		bitFocusedControl: 5,
		bitShadow:         false,
		bitShadowOffsetY:  1,
	}
	off = pressBit(off, tea.KeyDown)
	if off.bitShadow {
		t.Fatal("down should not toggle shadow on")
	}
	if off.bitShadowOffsetY != 1 {
		t.Fatalf("down while off changed Y to %d", off.bitShadowOffsetY)
	}
}
