package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestTooSmallWarningQuitWhileEditing(t *testing.T) {
	t.Run("custom text q quits", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 20
		m.editorMode = true
		m.textarea.Focus()

		view := m.View()
		if !strings.Contains(view, "Press Q to quit.") {
			t.Fatalf("expected too-small warning, got %q", view)
		}
		if strings.Contains(view, "Enter your ASCII art here...") {
			t.Fatal("editor view rendered under the too-small warning")
		}

		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if !isQuit(cmd) {
			t.Fatal("q should quit while the too-small warning is shown")
		}
		got := updated.(Model)
		if strings.Contains(got.textarea.Value(), "q") {
			t.Fatalf("q was typed into the editor: %q", got.textarea.Value())
		}
	})

	t.Run("custom text shift-q esc and ctrl+c quit", func(t *testing.T) {
		for _, k := range []tea.KeyMsg{
			{Type: tea.KeyRunes, Runes: []rune{'Q'}},
			{Type: tea.KeyEsc},
			{Type: tea.KeyCtrlC},
		} {
			m := NewModel()
			m.width = 40
			m.height = 10
			m.editorMode = true
			_, cmd := m.Update(k)
			if !isQuit(cmd) {
				t.Fatalf("%s should quit while the too-small warning is shown", k.String())
			}
		}
	})

	t.Run("custom text other keys are not inserted", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 20
		m.editorMode = true
		m.textarea.Focus()
		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
		if isQuit(cmd) {
			t.Fatal("x should not quit")
		}
		got := updated.(Model)
		if strings.Contains(got.textarea.Value(), "x") {
			t.Fatalf("x was typed into the hidden editor: %q", got.textarea.Value())
		}
		if !got.editorMode {
			t.Fatal("non-quit key should leave editor mode set")
		}
	})

	t.Run("bit editor q quits", func(t *testing.T) {
		m := NewModel()
		m.width = 80
		m.height = 20
		m.bitEditorMode = true
		m.bitFocusedControl = 2

		view := m.View()
		if !strings.Contains(view, "Press Q to quit.") {
			t.Fatalf("expected too-small warning, got %q", view)
		}

		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if !isQuit(cmd) {
			t.Fatal("q should quit from the BIT editor while the warning is shown")
		}
		got := updated.(Model)
		if got.bitTextInput.Value() != "" {
			t.Fatalf("q was typed into the BIT editor: %q", got.bitTextInput.Value())
		}
	})

	t.Run("large editor still types q", func(t *testing.T) {
		m := NewModel()
		m.width = 120
		m.height = 40
		m.editorMode = true
		m.textarea.Focus()
		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if isQuit(cmd) {
			t.Fatal("q should not quit a large custom text editor")
		}
		got := updated.(Model)
		if !strings.Contains(got.textarea.Value(), "q") {
			t.Fatalf("q should be typed in a large editor, got %q", got.textarea.Value())
		}
	})

	t.Run("unsized terminal is not the warning", func(t *testing.T) {
		m := NewModel()
		m.editorMode = true
		m.textarea.Focus()
		if m.View() != "Loading..." {
			t.Fatalf("unsized view = %q, want Loading...", m.View())
		}
		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if isQuit(cmd) {
			t.Fatal("q should not quit before the first window size")
		}
		got := updated.(Model)
		if !strings.Contains(got.textarea.Value(), "q") {
			t.Fatalf("q should reach the editor before sizing, got %q", got.textarea.Value())
		}
	})
}
