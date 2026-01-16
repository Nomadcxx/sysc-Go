package floydui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestTUIStability_Loop10 verifies the TUI can be created and updated
// 10 times in a row without panicking. This catches:
// - strings.Builder copy issues
// - Channel panics
// - sync.WaitGroup copy issues
// - Any other copy-by-value problems
func TestTUIStability_Loop10(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("Iteration", func(t *testing.T) {
			// Create a new model - should not panic
			m := NewModelWithSession("test_session.json")

			// Verify critical fields are initialized
			if m.CurrentResponse == nil {
				t.Fatal("CurrentResponse should be initialized as pointer")
			}
			if m.TokenStreamChan == nil {
				t.Fatal("TokenStreamChan should be initialized")
			}
			if m.StreamWaitGroup == nil {
				t.Fatal("StreamWaitGroup should be initialized as pointer")
			}

			// Initialize - should not panic
			cmd := m.Init()
			if cmd == nil {
				t.Log("Init returned nil cmd (may be ok)")
			}

			// Simulate window resize - this triggers a copy
			newModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
			m = newModel.(Model)

			// Simulate tick - this triggers a copy
			newModel, _ = m.Update(tickMsg(time.Now()))
			m = newModel.(Model)

			// Write to CurrentResponse - this is what caused the panic before
			m.CurrentResponse.WriteString("test token ")
			m.CurrentResponse.WriteString("another token")

			// Verify content
			content := m.CurrentResponse.String()
			if !strings.Contains(content, "test token") {
				t.Errorf("Expected 'test token' in response, got: %s", content)
			}

			// Simulate another update after writing - this is the critical test
			// because it copies the Model which has a used strings.Builder
			newModel, _ = m.Update(tickMsg(time.Now()))
			m = newModel.(Model)

			// Content should still be accessible after copy
			content = m.CurrentResponse.String()
			if !strings.Contains(content, "test token") {
				t.Errorf("Content lost after Update: %s", content)
			}

			// Reset and verify
			m.CurrentResponse.Reset()
			if m.CurrentResponse.Len() != 0 {
				t.Error("Reset should clear the builder")
			}

			// Write again after reset
			m.CurrentResponse.WriteString("after reset")

			// Final update - should still work
			newModel, _ = m.Update(tickMsg(time.Now()))
			m = newModel.(Model)

			if !strings.Contains(m.CurrentResponse.String(), "after reset") {
				t.Error("Content after reset should persist")
			}

			t.Logf("Iteration %d: PASS", i+1)
		})
	}
}

// TestModelCopyStability verifies that copying the Model struct
// multiple times doesn't cause panics with any fields
func TestModelCopyStability(t *testing.T) {
	m := NewModelWithSession("test_session.json")

	// Write some data to trigger the Builder's internal state
	m.CurrentResponse.WriteString("initial content")

	// Simulate what Bubble Tea does - copy the model many times
	for i := 0; i < 100; i++ {
		// This is what Bubble Tea does internally
		copy := m

		// Using the copy should work
		_ = copy.CurrentResponse.String()
		_ = copy.CurrentResponse.Len()

		// Writing to the COPY should also work (pointer is shared)
		copy.CurrentResponse.WriteString(" more")

		// The original should see the changes too (same pointer)
		if !strings.Contains(m.CurrentResponse.String(), "more") {
			t.Fatal("Pointer should be shared between copies")
		}
	}
}

// TestStreamTokenWriteLoop simulates the streaming token write loop
func TestStreamTokenWriteLoop(t *testing.T) {
	m := NewModelWithSession("test_session.json")

	// Simulate 100 streaming tokens being written
	for i := 0; i < 100; i++ {
		// Write token
		m.CurrentResponse.WriteString("token ")

		// Simulate Update (which copies the model)
		newModel, _ := m.Update(tickMsg(time.Now()))
		m = newModel.(Model)
	}

	content := m.CurrentResponse.String()
	expected := 100 * len("token ")
	if len(content) != expected {
		t.Errorf("Expected %d chars, got %d", expected, len(content))
	}

	t.Logf("Successfully wrote 100 tokens: %d chars total", len(content))
}
