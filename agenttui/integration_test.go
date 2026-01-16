//go:build integration

package agenttui

import (
	"fmt"
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestActualAPIStream tests the real API call
func TestActualAPIStream(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=1 to run")
	}

	m := NewAgentModel()

	// Check if client is configured
	if m.client == nil {
		t.Skip("No client configured - set ANTHROPIC_AUTH_TOKEN, GLM_API_KEY, or ZHIPU_API_KEY")
	}

	t.Log("Client initialized, testing API stream...")

	done := make(chan bool, 1)
	var streamChunks []string

	go func() {
		// Simulate sending a message
		submitCmd := m.SubmitMessage("say hello")
		if submitCmd == nil {
			t.Error("SubmitMessage returned nil")
			done <- true
			return
		}

		// Process messages in a loop
		currentCmd := submitCmd
		iterations := 0
		maxIterations := 200

		for iterations < maxIterations {
			if currentCmd == nil {
				t.Log("Command is nil, breaking")
				break
			}

			msg := currentCmd()
			if msg == nil {
				t.Log("Message is nil, breaking")
				break
			}

			t.Logf("Iter %d: msg type %T", iterations+1, msg)

			// Handle BatchMsg specially
			if batch, ok := msg.(tea.BatchMsg); ok {
				// Execute all commands in the batch
				var firstMsg tea.Msg
				for _, cmd := range batch {
					if cmd != nil {
						m := cmd()
						if m != nil && firstMsg == nil {
							firstMsg = m
						}
					}
				}
				msg = firstMsg
				if msg == nil {
					iterations++
					continue
				}
				t.Logf("  Batch first msg type: %T", msg)
			}

			// Check for StreamChunkMsg with tokens
			shouldBreak := false
			if chunk, ok := msg.(StreamChunkMsg); ok {
				if chunk.Token != "" {
					streamChunks = append(streamChunks, chunk.Token)
					t.Logf("  Got token: %q", chunk.Token)
				}
				if chunk.Done {
					t.Logf("  Stream done!")
					shouldBreak = true
				}
				if chunk.Error != nil {
					t.Logf("  Stream error: %v", chunk.Error)
					shouldBreak = true
				}
			}

			// Update model
			newModel, newCmd := m.Update(msg)
			m = newModel.(AgentModel)
			currentCmd = newCmd

			// Break if we got Done or Error
			if shouldBreak {
				break
			}

			// Check if we're done
			if !m.streaming && iterations > 5 {
				t.Logf("Streaming ended at iteration %d", iterations)
				break
			}

			iterations++
		}

		t.Logf("Final state: streaming=%v, chunks=%d", m.streaming, len(streamChunks))
		t.Logf("Stream content: %q", fmt.Sprint(streamChunks))

		if len(streamChunks) == 0 {
			t.Error("No stream chunks received!")
		}

		done <- true
	}()

	select {
	case <-done:
		// Test completed
	case <-time.After(15 * time.Second):
		t.Error("Test timed out")
	}
}

