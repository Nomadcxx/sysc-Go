package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// selectFile points the model at a file-selector entry by name.
func selectFile(t *testing.T, m Model, name string) Model {
	t.Helper()
	for i, f := range m.files {
		if f == name {
			m.selectedFile = i
			return m
		}
	}
	t.Fatalf("file %q not in selector: %v", name, m.files)
	return m
}

func key(k tea.KeyType, runes ...rune) tea.KeyMsg {
	return tea.KeyMsg{Type: k, Runes: runes}
}

// TestCustomTextReentryFocusesTextarea reproduces issue #55: Esc blurs the
// textarea, and entering Custom text again must Focus it or keystrokes are dropped.
func TestCustomTextReentryFocusesTextarea(t *testing.T) {
	m := NewModel()
	m.width = 120
	m.height = 40
	initialW, initialH := m.textarea.Width(), m.textarea.Height()
	m = selectFile(t, m, "Custom text")

	entered, _ := m.Update(key(tea.KeyEnter))
	m = entered.(Model)
	if !m.editorMode {
		t.Fatal("enter on Custom text should open the editor")
	}
	if !m.textarea.Focused() {
		t.Fatal("textarea should be focused on entry")
	}

	typed, _ := m.Update(key(tea.KeyRunes, 'A'))
	m = typed.(Model)
	if !strings.Contains(m.textarea.Value(), "A") {
		t.Fatalf("first entry should accept typing, got %q", m.textarea.Value())
	}

	left, _ := m.Update(key(tea.KeyEsc))
	m = left.(Model)
	if m.editorMode {
		t.Fatal("esc should leave the editor")
	}
	if m.textarea.Focused() {
		t.Fatal("esc should blur the textarea")
	}

	reentered, _ := m.Update(key(tea.KeyEnter))
	m = reentered.(Model)
	if !m.editorMode {
		t.Fatal("enter should reopen the editor")
	}
	if !m.textarea.Focused() {
		t.Fatal("textarea should be focused on re-entry")
	}
	if m.textarea.Width() == initialW || m.textarea.Height() == initialH {
		t.Fatalf("textarea size = %dx%d, still the %dx%d constructor default", m.textarea.Width(), m.textarea.Height(), initialW, initialH)
	}
	if m.textarea.Height() != m.height-10 {
		t.Fatalf("textarea height = %d, want %d", m.textarea.Height(), m.height-10)
	}

	retyped, _ := m.Update(key(tea.KeyRunes, 'B'))
	m = retyped.(Model)
	if !strings.Contains(m.textarea.Value(), "B") {
		t.Fatalf("re-entry should accept typing, got %q", m.textarea.Value())
	}
}
