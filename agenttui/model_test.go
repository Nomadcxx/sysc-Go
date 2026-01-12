package agenttui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestStreamingFlow tests the complete streaming flow
func TestStreamingFlow(t *testing.T) {
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

// TestChannelCloseBehavior tests that when a channel closes, waitForStream handles it correctly
func TestChannelCloseBehavior(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("CloseIteration_%d", i), func(t *testing.T) {
			m := NewAgentModel()

			// Create a channel and immediately close it
			ch := make(chan StreamChunkMsg, 10)
			close(ch)

			// Set it as the streamChan
			m.streamChanMu.Lock()
			m.streamChan = ch
			m.streamChan = ch // Set twice to be sure
			m.streamChanMu.Unlock()
			m.streaming = true

			// Run waitForStream multiple times (simulating in-flight calls)
			var doneCount int
			for j := 0; j < 5; j++ {
				cmd := m.waitForStream()
				msg := cmd()
				if chunk, ok := msg.(StreamChunkMsg); ok {
					if chunk.Done {
						doneCount++
					}
				}
			}

			// Only the first one should get Done: true, rest should get nil channel
			t.Logf("Got %d Done messages from 5 waitForStream calls on closed channel", doneCount)

			if doneCount > 1 {
				t.Errorf("Expected at most 1 Done message, got %d - multiple waitForStream calls are reading the closed channel", doneCount)
			}

			// After first close detection, streamChan should be nil
			m.streamChanMu.Lock()
			if m.streamChan != nil {
				t.Error("Expected streamChan to be nil after closed channel detection")
			}
			m.streamChanMu.Unlock()
		})
	}
}

// TestNoPanicOnClosedChannel tests that we don't panic when accessing a closed channel
func TestNoPanicOnClosedChannel(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic detected: %v", r)
		}
	}()

	m := NewAgentModel()

	// Create and close a channel
	ch := make(chan StreamChunkMsg, 10)
	close(ch)

	m.streamChanMu.Lock()
	m.streamChan = ch
	m.streamChanMu.Unlock()
	m.streaming = true

	// This should not panic
	for i := 0; i < 10; i++ {
		cmd := m.waitForStream()
		_ = cmd()
	}
}
