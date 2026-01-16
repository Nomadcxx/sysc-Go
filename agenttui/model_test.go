//go:build integration

package agenttui

import (
	"fmt"
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestStreamingFlow tests the complete streaming flow
func TestStreamingFlow(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=1 to run")
	}

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("Iteration_%d", i), func(t *testing.T) {
			testSingleStream(t)
		})
	}
}

func testSingleStream(t *testing.T) {
	m := NewAgentModel()
	m.SetSize(80, 24)

	// Track all messages received
	var messages []tea.Msg
	done := make(chan bool, 1)

	// Simulate sending a message
	submitCmd := m.SubmitMessage("hello")
	if submitCmd == nil {
		t.Fatal("SubmitMessage returned nil command")
	}

	// Start a goroutine to process messages
	go func() {
		currentModel := m
		currentCmd := submitCmd
		messageCount := 0
		maxMessages := 500 // Safety limit to catch infinite loops

		for messageCount < maxMessages {
			// Execute current command
			if currentCmd == nil {
				break
			}

			msg := currentCmd()
			if msg == nil {
				break
			}

			messageCount++
			messages = append(messages, msg)

			// Update model
			newModel, newCmd := currentModel.Update(msg)
			currentModel = newModel.(AgentModel)
			currentCmd = newCmd

			// Check if streaming is done
			if !currentModel.streaming && messageCount > 5 {
				// Stream ended gracefully
				done <- true
				return
			}
		}

		// If we get here, either maxMessages reached or stream ended
		if currentModel.streaming {
			t.Errorf("Stream still active after %d messages", messageCount)
		}
		done <- true
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		// Test completed
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out - possible infinite loop")
	}

	// Validate expectations
	t.Logf("Received %d messages", len(messages))

	// Find messages by type
	var chunkCount, doneCount, tickCount int
	for _, msg := range messages {
		switch m := msg.(type) {
		case StreamChunkMsg:
			chunkCount++
			if m.Done {
				doneCount++
			}
		case TickMsg:
			tickCount++
		}
	}

	t.Logf("Message breakdown: chunks=%d (done=%d), ticks=%d", chunkCount, doneCount, tickCount)

	// Expectations:
	// 1. At least one Done message (stream completed)
	if doneCount == 0 && chunkCount > 0 {
		t.Error("Expected at least one Done message when chunks were received")
	}

	// 2. Not an excessive number of ticks (the loop of death would generate thousands)
	if tickCount > 200 {
		t.Errorf("Too many tick messages (%d), possible infinite loop", tickCount)
	}
}

