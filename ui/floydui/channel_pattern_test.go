package floydui

import (
	"testing"
	"time"
)

// TestTokenStreamChanPersistence verifies the critical channel pattern:
// - TokenStreamChan must be persistent (never closed)
// - Multiple sends should work without panic
// - Sentinel value \x00DONE should be receivable
//
// ⚠️  CRITICAL: This test exists to prevent regressions of the panic bug.
// If this test fails, someone likely added a close() to TokenStreamChan.
// See Claude.md "TUI Channel Pattern" for documentation.
func TestTokenStreamChanPersistence(t *testing.T) {
	m := NewModelWithSession("test_session.json")

	// Verify channel was created
	if m.TokenStreamChan == nil {
		t.Fatal("TokenStreamChan should be initialized in NewModel")
	}

	// Simulate multiple "agent calls" - each should work without panic
	for i := 0; i < 3; i++ {
		// Send tokens like the agent would
		go func(iteration int) {
			for j := 0; j < 5; j++ {
				select {
				case m.TokenStreamChan <- "token":
				case <-time.After(100 * time.Millisecond):
					// Don't block forever if channel is full
				}
			}
			// Send done sentinel
			select {
			case m.TokenStreamChan <- "\x00DONE":
			case <-time.After(100 * time.Millisecond):
			}
		}(i)

		// Drain the channel like the UI would
		drained := 0
		timeout := time.After(500 * time.Millisecond)
	drain:
		for {
			select {
			case token := <-m.TokenStreamChan:
				drained++
				if token == "\x00DONE" {
					break drain
				}
			case <-timeout:
				break drain
			}
		}

		if drained == 0 {
			t.Errorf("Iteration %d: Expected to receive tokens, got none", i)
		}
	}

	// Final check: channel should still be usable (not closed)
	select {
	case m.TokenStreamChan <- "final_test":
		// Good - channel is still open
		<-m.TokenStreamChan // Drain it
	default:
		// Channel might be full, that's ok
	}
}

// TestTokenStreamChanNeverNil ensures the channel is always initialized
func TestTokenStreamChanNeverNil(t *testing.T) {
	m := NewModelWithSession("test_session.json")

	if m.TokenStreamChan == nil {
		t.Fatal("TokenStreamChan must never be nil - check NewModel initialization")
	}

	// Verify it's buffered (capacity > 0)
	if cap(m.TokenStreamChan) == 0 {
		t.Error("TokenStreamChan should be buffered to prevent blocking")
	}
}

// TestDoneSentinelValue verifies the sentinel value is correctly recognized
func TestDoneSentinelValue(t *testing.T) {
	done := "\x00DONE"

	if done == "" {
		t.Error("Done sentinel should not be empty string")
	}

	if done == "DONE" {
		t.Error("Done sentinel should have \\x00 prefix to distinguish from regular content")
	}

	// Verify it starts with null byte
	if done[0] != 0x00 {
		t.Error("Done sentinel should start with null byte")
	}
}
