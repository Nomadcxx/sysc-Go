package floydui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestTypingFullSentence tests typing a complete sentence
func TestTypingFullSentence(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	sentence := "Hello FLOYD, can you help me?"

	for _, char := range sentence {
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{char},
		}
		newModel, _ := m.Update(keyMsg)
		m = newModel.(Model)
	}

	if m.Input.Value() != sentence {
		t.Errorf("Expected %q, got %q", sentence, m.Input.Value())
	}
}

// TestBackspaceDeletesCharacter tests that backspace works
func TestBackspaceDeletesCharacter(t *testing.T) {
	m := NewModelWithSession("test_session.json")

	// Type "hello"
	for _, char := range "hello" {
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{char},
		}
		newModel, _ := m.Update(keyMsg)
		m = newModel.(Model)
	}

	if m.Input.Value() != "hello" {
		t.Fatalf("Expected 'hello', got %q", m.Input.Value())
	}

	// Press backspace twice
	for i := 0; i < 2; i++ {
		keyMsg := tea.KeyMsg{
			Type: tea.KeyBackspace,
		}
		newModel, _ := m.Update(keyMsg)
		m = newModel.(Model)
	}

	if m.Input.Value() != "hel" {
		t.Errorf("Expected 'hel' after backspace, got %q", m.Input.Value())
	}
}

// TestHistoryNavigation tests up/down arrow history navigation
func TestHistoryNavigation(t *testing.T) {
	m := NewModelWithSession("test_session.json")

	// Add some history
	m.History = []string{"first message", "second message", "third message"}
	m.HistoryIndex = len(m.History)

	// Press Up
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := m.Update(keyMsg)
	m = newModel.(Model)

	if m.Input.Value() != "third message" {
		t.Errorf("Expected 'third message', got %q", m.Input.Value())
	}

	// Press Up again
	keyMsg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.Update(keyMsg)
	m = newModel.(Model)

	if m.Input.Value() != "second message" {
		t.Errorf("Expected 'second message', got %q", m.Input.Value())
	}

	// Press Down
	keyMsg = tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ = m.Update(keyMsg)
	m = newModel.(Model)

	if m.Input.Value() != "third message" {
		t.Errorf("Expected 'third message' after down, got %q", m.Input.Value())
	}
}

// TestSpaceKey tests that space key works (space is sent as a rune, not KeySpace type)
func TestSpaceKey(t *testing.T) {
	m := NewModelWithSession("test_session.json")

	// Type "hello"
	for _, char := range "hello" {
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{char},
		}
		newModel, _ := m.Update(keyMsg)
		m = newModel.(Model)
	}

	// Press space - space is sent as a KeyRunes with space character
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{' '},
	}
	newModel, _ := m.Update(keyMsg)
	m = newModel.(Model)

	// Type "world"
	for _, char := range "world" {
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{char},
		}
		newModel, _ = m.Update(keyMsg)
		m = newModel.(Model)
	}

	expected := "hello world"
	if m.Input.Value() != expected {
		t.Errorf("Expected %q, got %q", expected, m.Input.Value())
	}
}

// TestQuestionMarkToggle tests that ? toggles help only when input is empty
func TestQuestionMarkToggle(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	initialHelpState := m.ShowHelp

	// Press ? with empty input
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'?'},
	}
	newModel, _ := m.Update(keyMsg)
	m = newModel.(Model)

	// Help should toggle
	if m.ShowHelp == initialHelpState {
		t.Error("Help should toggle when pressing ? with empty input")
	}

	// Set input to non-empty
	m.Input.SetValue("test")
	initialHelpState = m.ShowHelp

	// Press ? with non-empty input
	keyMsg = tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'?'},
	}
	newModel, _ = m.Update(keyMsg)
	m = newModel.(Model)

	// Help should NOT toggle, ? should be added to input
	if m.ShowHelp != initialHelpState {
		t.Error("Help should not toggle when pressing ? with non-empty input")
	}
	if !strings.Contains(m.Input.Value(), "?") {
		t.Error("Question mark should be added to input when input is not empty")
	}
}

// TestEnterWithSlashCommand tests that slash commands are processed
func TestEnterWithSlashCommand(t *testing.T) {
	// Add a message first
	m := NewModelWithSession("test_session.json")
	m.Messages = append(m.Messages, ChatMessage{Role: "user", Content: "test message"})

	m.Input.SetValue("/clear")

	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(keyMsg)
	m = newModel.(Model)

	// Messages should be cleared
	if len(m.Messages) != 0 {
		t.Errorf("Messages should be cleared after /clear, got %d messages", len(m.Messages))
	}
}

// TestCtrlHTogglesHelp tests Ctrl+H help toggle
func TestCtrlHTogglesHelp(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	initialHelpState := m.ShowHelp

	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlH}
	newModel, _ := m.Update(keyMsg)
	m = newModel.(Model)

	if m.ShowHelp == initialHelpState {
		t.Error("Ctrl+H should toggle help")
	}
}
