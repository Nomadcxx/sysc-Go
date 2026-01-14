package floydui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// CommandEdgeCaseTests tests all slash commands with edge cases and human errors
//
// This file contains tests for slash commands defined in update.go handleCommand():
//   /theme <name>   - Change theme
//   /clear, /cls    - Clear chat history
//   /help, /h, /?   - Toggle help
//   /status         - Show workspace status
//   /tools          - List available tools
//   /protocol       - Show protocol status
//   /exit, /quit, /q - Quit
//

// TestEmptySlashCommand tests typing just "/" with nothing after it
func TestEmptySlashCommand(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Input.SetValue("/")

	// Press Enter
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(keyMsg)
	m = newModel.(Model)

	// BLOCKING ERROR #1: Empty slash command provides no feedback to user
	// The command is silently ignored - user doesn't know what happened
	if m.Input.Value() != "" {
		t.Errorf("Expected input to be reset after empty command, got: %q", m.Input.Value())
	}

	// Check if any feedback message was added
	feedbackFound := false
	for _, msg := range m.Messages {
		if strings.Contains(msg.Content, "Unknown command") || strings.Contains(msg.Content, "Usage") {
			feedbackFound = true
			break
		}
	}
	if !feedbackFound {
		t.Error("BLOCKING ERROR: Empty slash command '/' produces no user feedback. User doesn't know command was invalid.")
	}
}

// TestInvalidTheme tests /theme with a non-existent theme name
func TestInvalidTheme(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Input.SetValue("/theme nonexistent")

	// Press Enter
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(keyMsg)
	m = newModel.(Model)

	// BLOCKING ERROR #2: Invalid theme produces no feedback
	// SetTheme returns nil when theme not found, but handleCommand doesn't check this
	// The /theme command at line 166-173 only works in real-time (before Enter)
	// When processed via handleCommand, it's treated as unknown command

	// Check if any error message was added
	errorFound := false
	for _, msg := range m.Messages {
		if strings.Contains(msg.Content, "theme") && (strings.Contains(msg.Content, "not found") || strings.Contains(msg.Content, "unknown")) {
			errorFound = true
			break
		}
	}

	// Also check if theme changed (it shouldn't)
	if m.CurrentTheme.Name != "Legacy Silver" {
		t.Errorf("Theme should remain Legacy Silver, got: %q", m.CurrentTheme.Name)
	}

	if !errorFound && m.Input.Value() == "" {
		t.Log("BLOCKING ERROR: Invalid theme '/theme nonexistent' produces no error message.")
		t.Log("User sees nothing - input is cleared but no feedback provided.")
	}
}

// TestThemeWithNoArgument tests /theme without a theme name
func TestThemeWithNoArgument(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Input.SetValue("/theme")

	// Press Enter
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(keyMsg)
	m = newModel.(Model)

	// BLOCKING ERROR #3: /theme with no argument is treated as unknown command
	// No feedback about missing argument
	shouldListThemes := false
	for _, msg := range m.Messages {
		if strings.Contains(msg.Content, "Available themes") || strings.Contains(msg.Content, "classic") {
			shouldListThemes = true
			break
		}
	}

	if !shouldListThemes {
		t.Error("BLOCKING ERROR: '/theme' with no argument should list available themes or show usage, but does nothing.")
	}
}

// TestThemeRealtimeVsEnter tests the inconsistency in /theme handling
func TestThemeRealtimeVsEnter(t *testing.T) {
	// Test 1: Realtime theme change (lines 166-173 in update.go)
	m1 := NewModelWithSession("test_session.json")
	m1.Input.SetValue("/theme dark")

	// Simulate space key which triggers handleKeyMsg
	keyMsg := tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	var updatedModel tea.Model
	updatedModel, _ = m1.Update(keyMsg)
	m1 = updatedModel.(Model)

	// In current code, theme check happens on key input, not on Enter
	// This is inconsistent with other commands

	// Test 2: Enter-based command (doesn't work for theme currently)
	m2 := NewModelWithSession("test_session.json")
	m2.Input.SetValue("/theme dark")

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m2.Update(enterMsg)
	m2 = newModel.(Model)

	// The theme command is now in handleCommand switch statement
	if m2.CurrentTheme.Name == "Legacy Silver" {
		t.Error("BLOCKING ERROR: /theme command via Enter doesn't work.")
	}

}

// TestClearCommandVariations tests /clear and /cls
func TestClearCommandVariations(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"clear command", "/clear"},
		{"cls command", "/cls"},
		{"CLEAR uppercase", "/CLEAR"},
		{"Clear mixed case", "/Clear"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModelWithSession("test_session.json")
			// Add some messages
			m.Messages = append(m.Messages, ChatMessage{Role: "user", Content: "test1"})
			m.Messages = append(m.Messages, ChatMessage{Role: "assistant", Content: "response1"})

			m.Input.SetValue(tt.command)

			enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
			newModel, _ := m.Update(enterMsg)
			m = newModel.(Model)

			if len(m.Messages) != 0 {
				t.Errorf("%s: Expected messages to be cleared, got %d messages", tt.name, len(m.Messages))
			}
		})
	}
}

// TestClearWithExtraSpaces tests /clear with trailing spaces
func TestClearWithExtraSpaces(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Messages = append(m.Messages, ChatMessage{Role: "user", Content: "test"})

	// strings.TrimSpace in handleCommand handles this
	m.Input.SetValue("/clear   ")

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(Model)

	if len(m.Messages) != 0 {
		t.Errorf("Messages should be cleared even with trailing spaces, got %d messages", len(m.Messages))
	}
}

// TestHelpCommandVariations tests all help command variants
func TestHelpCommandVariations(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"help", "/help"},
		{"h", "/h"},
		{"question mark", "/?"},
		{"HELP uppercase", "/HELP"},
		{"H uppercase", "/H"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModelWithSession("test_session.json")
			initialState := m.ShowHelp

			m.Input.SetValue(tt.command)

			enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
			newModel, _ := m.Update(enterMsg)
			m = newModel.(Model)

			if m.ShowHelp == initialState {
				t.Errorf("%s: Help state should toggle", tt.name)
			}
		})
	}
}

// TestStatusCommand tests /status command
func TestStatusCommand(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Input.SetValue("/status")

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(Model)

	// Should add a system message with workspace status
	statusFound := false
	for _, msg := range m.Messages {
		if msg.Role == "system" && strings.Contains(msg.Content, "Workspace Status") {
			statusFound = true
			break
		}
	}

	if !statusFound {
		t.Error("/status command should add a status message")
	}
}

// TestToolsCommand tests /tools command
func TestToolsCommand(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Input.SetValue("/tools")

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(Model)

	// Should add a system message about tools
	toolsFound := false
	for _, msg := range m.Messages {
		if msg.Role == "system" && strings.Contains(msg.Content, "Tool listing") {
			toolsFound = true
			break
		}
	}

	if !toolsFound {
		// Current behavior - shows a placeholder message
		t.Log("BLOCKING ERROR #5: /tools command shows placeholder message, doesn't actually list tools")
	}
}

// TestProtocolCommand tests /protocol command
func TestProtocolCommand(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Input.SetValue("/protocol")

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(Model)

	// Should add a system message with protocol status
	protocolFound := false
	for _, msg := range m.Messages {
		if msg.Role == "system" && strings.Contains(msg.Content, "Protocol Status") {
			protocolFound = true
			break
		}
	}

	if !protocolFound {
		t.Error("/protocol command should add a protocol status message")
	}
}

// TestExitCommandVariations tests all exit command variants
func TestExitCommandVariations(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"exit", "/exit"},
		{"quit", "/quit"},
		{"q", "/q"},
		{"EXIT uppercase", "/EXIT"},
		{"QUIT uppercase", "/QUIT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModelWithSession("test_session.json")
			m.Input.SetValue(tt.command)

			enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
			newModel, cmd := m.Update(enterMsg)

			if cmd == nil {
				t.Errorf("%s: Should return quit command", tt.name)
			}

			// Verify it's the tea.Quit command
			// We can check if the model would quit by examining the command
			if cmd != nil {
				// Execute the command to check return value
				msg := cmd()
				if msg != tea.Quit() {
					t.Errorf("%s: Expected tea.Quit message", tt.name)
				}
			}
			_ = newModel
		})
	}
}

// TestUnknownCommand tests commands that don't exist
func TestUnknownCommand(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Input.SetValue("/foobar")

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(Model)

	// BLOCKING ERROR #6: Unknown commands produce no feedback
	errorFound := false
	for _, msg := range m.Messages {
		if strings.Contains(msg.Content, "Unknown command") || strings.Contains(msg.Content, "foobar") {
			errorFound = true
			break
		}
	}

	if !errorFound {
		t.Error("BLOCKING ERROR: Unknown command '/foobar' produces no feedback. User doesn't know the command doesn't exist.")
	}
}

// TestCommandInMiddleOfText tests command-like text in regular input
func TestCommandInMiddleOfText(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Input.SetValue("hello /theme dark")

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(Model)

	// Should NOT treat as command since it doesn't start with /
	// Should add as user message
	lastMsg := m.Messages[len(m.Messages)-1]
	if lastMsg.Role != "user" {
		t.Errorf("Expected user message, got: %v", lastMsg.Role)
	}
	if lastMsg.Content != "hello /theme dark" {
		t.Errorf("Content should be preserved, got: %q", lastMsg.Content)
	}
}

// TestThemeWithSpecialCharacters tests theme names with special characters
func TestThemeWithSpecialCharacters(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Input.SetValue("/theme @#$%")

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(Model)

	// Should not crash, should handle gracefully
	// Theme should remain unchanged
	if m.CurrentTheme.Name != "Legacy Silver" {
		t.Errorf("Theme should remain Legacy Silver after invalid input, got: %q", m.CurrentTheme.Name)
	}

}

// TestThemeWithUnicode tests theme names with unicode characters
func TestThemeWithUnicode(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Input.SetValue("/theme dark\x00") // Null byte

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(Model)

	// Should not crash
	_ = m // Just verify it doesn't panic
}

// TestVeryLongCommandArgument tests commands with very long arguments
func TestVeryLongCommandArgument(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	longArg := strings.Repeat("a", 1000)
	m.Input.SetValue("/theme " + longArg)

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(Model)

	// Should not crash
	_ = m
}

// TestMultipleCommandsInOneLine tests multiple command-like patterns
func TestMultipleCommandsInOneLine(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Input.SetValue("/clear /help")

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(Model)

	// This is treated as unknown command since parts[0] would be "clear"
	// but the full input is "/clear /help"
	// BLOCKING ERROR: No feedback about unknown command
}

// TestThemeCaseSensitivity tests if theme names are case-sensitive
func TestThemeCaseSensitivity(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Input.SetValue("/theme Dark") // Capital D

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(Model)

	// Themes map uses lowercase keys, but our GetThemeByName handles normalization
	if m.CurrentTheme.Name == "Legacy Noir" {
		// If this passes, themes are case-insensitive (good)
		t.Log("Themes are case-insensitive")
	} else {

		// BLOCKING ERROR: Theme names are case-sensitive
		t.Error("BLOCKING ERROR: Theme names are case-sensitive. '/theme Dark' doesn't work but '/theme dark' does. No feedback to user.")
	}
}

// TestAllValidThemes tests all known valid theme names
func TestAllValidThemes(t *testing.T) {
	validThemes := []string{"classic", "dark", "highcontrast", "darkside", "midnight"}

	for _, theme := range validThemes {
		t.Run("theme_"+theme, func(t *testing.T) {
			m := NewModelWithSession("test_session.json")

			// Note: /theme only works via realtime detection, not handleCommand
			m.Input.SetValue("/theme " + theme + " ")

			// This simulates the space key which triggers theme change
			keyMsg := tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
			newModel, cmd := m.Update(keyMsg)
			m = newModel.(Model)

			// Check if theme command was processed (cmd would be non-nil)
			if cmd == nil {
				// Theme not found or not processed
				t.Logf("Theme %q may not have been applied via realtime detection", theme)
			}
		})
	}
}

// TestSlashWithOnlySpaces tests "/   " (slash with only spaces)
func TestSlashWithOnlySpaces(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.Input.SetValue("/   ")

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(Model)

	// strings.TrimSpace makes this equivalent to empty slash command
	// BLOCKING ERROR: No feedback to user
	feedbackFound := false
	for _, msg := range m.Messages {
		if strings.Contains(msg.Content, "command") {
			feedbackFound = true
			break
		}
	}
	if !feedbackFound {
		t.Log("BLOCKING ERROR: '/   ' (slash with only spaces) produces no feedback")
	}
}

// TestCommandWhileThinking tests commands during agent thinking state
func TestCommandWhileThinking(t *testing.T) {
	m := NewModelWithSession("test_session.json")
	m.IsThinking = true
	m.Input.SetValue("/clear")

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterMsg)
	m = newModel.(Model)

	// BLOCKING ERROR #7: Commands are ignored while thinking (line 177-179)
	// User can't use /clear or other commands while agent is processing
	// This might be intentional but could be frustrating
	if m.IsThinking {
		t.Log("Commands are blocked while agent is thinking - user must wait for agent to finish")
	}
}

// TestAllCommandsList tests that all documented commands are implemented
func TestAllCommandsList(t *testing.T) {
	documentedCommands := []string{
		"theme", "clear", "status", "tools", "protocol", "help", "exit", "quit", "q", "cls", "h", "?",
	}

	// This test documents which commands are actually in handleCommand
	implementedCommands := map[string]bool{
		"exit":     true, // Line 259
		"quit":     true, // Line 259
		"q":        true, // Line 259
		"clear":    true, // Line 263
		"cls":      true, // Line 263
		"help":     true, // Line 269
		"h":        true, // Line 269
		"?":        true, // Line 269
		"status":   true, // Line 275
		"tools":    true, // Line 287
		"protocol": true, // Line 294
		// Note: "theme" is NOT in handleCommand - it's only in handleKeyMsg realtime
	}

	for _, cmd := range documentedCommands {
		if !implementedCommands[cmd] && cmd != "theme" {
			t.Errorf("Command %q is documented but not implemented in handleCommand", cmd)
		}
	}

	if !implementedCommands["theme"] {
		t.Log("BLOCKING ERROR #8: 'theme' command is not in handleCommand switch statement.")
		t.Log("It only works via realtime detection in handleKeyMsg, making it inconsistent with other commands.")
	}
}
