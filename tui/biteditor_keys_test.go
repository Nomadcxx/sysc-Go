package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func bitKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func bitNamedKey(keyType tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: keyType}
}

func applyBitKeys(m Model, keys ...tea.KeyMsg) Model {
	for _, key := range keys {
		updated, _ := m.handleBitEditorKeyPress(key)
		m = updated.(Model)
	}
	return m
}

func TestBitTextFieldInsertsNavigationLetters(t *testing.T) {
	m := NewModel()
	m.bitEditorMode = true
	m.bitFocusedControl = 0
	m.bitTextInput.Focus()
	startAlign := m.bitAlignment

	m = applyBitKeys(m, bitKey('H'), bitKey('e'), bitKey('l'), bitKey('l'), bitKey('o'))
	if got := m.bitTextInput.Value(); got != "Hello" {
		t.Fatalf("text field = %q, want Hello", got)
	}
	if m.bitAlignment != startAlign {
		t.Fatalf("alignment changed from %d to %d while typing", startAlign, m.bitAlignment)
	}

	m.bitTextInput.SetValue("")
	m = applyBitKeys(m, bitKey('h'), bitKey('j'), bitKey('k'), bitKey('l'))
	if got := m.bitTextInput.Value(); got != "hjkl" {
		t.Fatalf("text field = %q, want hjkl", got)
	}
	if m.bitFocusedControl != 0 {
		t.Fatalf("focus left the text field: %d", m.bitFocusedControl)
	}
}

func TestBitNavAliasesStayOnNonTextControls(t *testing.T) {
	m := NewModel()
	m.bitEditorMode = true
	m.bitFocusedControl = 2 // alignment
	m.bitTextInput.Blur()
	m.bitAlignment = 1

	m = applyBitKeys(m, bitKey('h'))
	if m.bitAlignment != 0 {
		t.Fatalf("h on alignment = %d, want 0", m.bitAlignment)
	}
	if got := m.bitTextInput.Value(); got != "" {
		t.Fatalf("h inserted %q while alignment was focused", got)
	}

	m = applyBitKeys(m, bitKey('l'), bitKey('l'))
	if m.bitAlignment != 2 {
		t.Fatalf("l on alignment = %d, want 2", m.bitAlignment)
	}

	m = applyBitKeys(m, bitNamedKey(tea.KeyLeft))
	if m.bitAlignment != 1 {
		t.Fatalf("left arrow on alignment = %d, want 1", m.bitAlignment)
	}
}

func TestBitTextFieldArrowsDoNotInsert(t *testing.T) {
	m := NewModel()
	m.bitEditorMode = true
	m.bitFocusedControl = 0
	m.bitTextInput.Focus()
	m.bitTextInput.SetValue("ab")
	startAlign := m.bitAlignment

	m = applyBitKeys(m,
		bitNamedKey(tea.KeyLeft),
		bitNamedKey(tea.KeyRight),
		bitNamedKey(tea.KeyUp),
		bitNamedKey(tea.KeyDown),
	)
	if got := m.bitTextInput.Value(); got != "ab" {
		t.Fatalf("arrows changed text to %q", got)
	}
	if m.bitAlignment != startAlign {
		t.Fatalf("arrows changed alignment from %d to %d", startAlign, m.bitAlignment)
	}
}
