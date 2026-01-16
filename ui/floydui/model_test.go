package floydui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestInitReturnsBlinkCommand verifies that Init() returns the textinput.Blink command
func TestInitReturnsBlinkCommand(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	cmd := m.Init()

	if cmd == nil {
		t.Fatal("Init() should return a non-nil command")
	}
}

// TestInputIsFocused verifies that the textinput is focused after model creation
func TestInputIsFocused(t *testing.T) {
	m := NewModelWithSession("test_session.json")

	if !m.Input.Focused() {
		t.Error("Input should be focused after model creation")
	}
}

// TestKeyMsgCharacterInput verifies that typing characters updates the input field
// This test demonstrates the BUG: character input does NOT work
func TestKeyMsgCharacterInput(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	initialValue := m.Input.Value()

	// Simulate typing the character 'a'
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'a'},
	}

	newModel, cmd := m.Update(keyMsg)

	// Check if the input value changed
	newValue := newModel.(Model).Input.Value()
	if newValue == initialValue {
		t.Errorf("BUG CONFIRMED: Input value did not change after typing 'a'. Got: %q, want: %q", newValue, initialValue+"a")
	} else if newValue != initialValue+"a" {
		t.Errorf("Input value changed unexpectedly. Got: %q, want: %q", newValue, initialValue+"a")
	}

	if cmd != nil {
		t.Logf("Command returned: %v", cmd)
	}
}

// TestKeyMsgEnterProcessesInput verifies that Enter key processes the input
func TestKeyMsgEnterProcessesInput(t *testing.T) {
	m := NewModelWithSession("test_session.json")

	// First, set a value directly (since typing doesn't work due to the bug)
	m.Input.SetValue("test input")

	// Now press Enter
	keyMsg := tea.KeyMsg{
		Type: tea.KeyEnter,
	}

	newModel, _ := m.Update(keyMsg)
	model := newModel.(Model)

	// After Enter, the input should be reset
	if model.Input.Value() != "" {
		t.Errorf("Input should be reset after Enter, got: %q", model.Input.Value())
	}

	// And the message should be added to Messages
	if len(model.Messages) == 0 {
		t.Error("No messages added after Enter")
	} else {
		lastMsg := model.Messages[len(model.Messages)-1]
		if lastMsg.Role != "user" {
			t.Errorf("Expected user message, got: %v", lastMsg.Role)
		}
		if lastMsg.Content != "test input" {
			t.Errorf("Expected content 'test input', got: %q", lastMsg.Content)
		}
	}
}

// TestKeyMsgCtrlCQuits verifies that Ctrl+C quits the application
func TestKeyMsgCtrlCQuits(t *testing.T) {
	m := NewModelWithSession("test_session.json")

	keyMsg := tea.KeyMsg{
		Type: tea.KeyCtrlC,
	}

	newModel, _ := m.Update(keyMsg)
	model := newModel.(Model)

	// First Ctrl+C sets ModeExitSummary (two-step exit for safety)
	if model.Mode != ModeExitSummary {
		t.Errorf("Ctrl+C should set ModeExitSummary, got: %v", model.Mode)
	}

	// Second Ctrl+C should return tea.Quit
	secondModel, cmd := model.Update(keyMsg)
	if cmd == nil {
		t.Error("Second Ctrl+C should return tea.Quit")
	}
	_ = secondModel
}

// TestKeyMsgWithSpecialCharacters tests various special keys
func TestKeyMsgWithSpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		keyType  tea.KeyType
		wantDesc string
	}{
		{"Backspace", tea.KeyBackspace, "should delete character"},
		{"Space", tea.KeySpace, "should add space"},
		{"Tab", tea.KeyTab, "should handle tab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModelWithSession("test_session.json")
			m.Input.SetValue("abc")

			keyMsg := tea.KeyMsg{
				Type: tt.keyType,
			}

			newModel, _ := m.Update(keyMsg)
			newValue := newModel.(Model).Input.Value()

			// Just log what happens - the actual behavior depends on the bug fix
			t.Logf("%s: Input value after key: %q", tt.name, newValue)
		})
	}
}

// TestMultipleCharacterInput tests typing multiple characters
func TestMultipleCharacterInput(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	testString := "hello"

	for _, char := range testString {
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{char},
		}

		newModel, _ := m.Update(keyMsg)
		m = newModel.(Model)
	}

	// This will fail until the bug is fixed
	if m.Input.Value() != testString {
		t.Errorf("BUG CONFIRMED: After typing %q, input has %q", testString, m.Input.Value())
	}
}

// TestUpdateChildComponentsNeverReachedForKeys verifies the architectural issue
// This test documents that m.Input.Update(msg) is never called for KeyMsg
func TestUpdateChildComponentsNeverReachedForKeys(t *testing.T) {
	m := NewModelWithSession("test_session.json")

	// Set a traceable value
	m.Input.SetValue("initial")

	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'x'},
	}

	// The Update function calls handleKeyMsg which returns early
	// This means the code at lines 111-112 in update.go is NEVER reached for KeyMsg
	newModel, _ := m.Update(keyMsg)

	// If child components were updated, the value would change
	// Since it doesn't, we've confirmed the architectural issue
	if newModel.(Model).Input.Value() == "initial" {
		t.Log("Confirmed: Input.Update() is not called for KeyMsg in Update()")
	}
}

// BenchmarkKeyMsgHandling benchmarks key message handling
func BenchmarkKeyMsgHandling(b *testing.B) {
	m := NewModelWithSession("test_session.json")
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'a'},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var newModel tea.Model
		newModel, _ = m.Update(keyMsg)
		m = newModel.(Model)
	}
}
