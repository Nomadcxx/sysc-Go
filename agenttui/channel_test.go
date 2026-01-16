package agenttui

import (
	"testing"
)

// TestChannelCloseBehavior tests that when a channel closes, waitForStream handles it correctly
func TestChannelCloseBehavior(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(t.Name()+"/CloseIteration_test", func(t *testing.T) {
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

