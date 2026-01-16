package agent

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestClientCanConnect(t *testing.T) {
	client := NewProxyClient("", "")

	if client.apiKey == "" {
		t.Fatal("No API key found - check ~/.claude/settings.json or ANTHROPIC_AUTH_TOKEN env var")
	}

	t.Logf("API key found: %s...", client.apiKey[:10])

	// Try a simple request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := ChatRequest{
		Messages:  []Message{{Role: "user", Content: "hi"}},
		MaxTokens: 10,
	}

	chunkChan, err := client.StreamChat(ctx, req)
	if err != nil {
		t.Fatalf("StreamChat failed: %v", err)
	}

	// Read a few chunks
	chunkCount := 0
	timeout := time.After(5 * time.Second)
	for chunkCount < 5 {
		select {
		case chunk, ok := <-chunkChan:
			if !ok {
				t.Log("Stream closed")
				return
			}
			if chunk.Error != nil {
				t.Fatalf("Received error chunk: %v", chunk.Error)
			}
			chunkCount++
			t.Logf("Chunk %d: token=%q done=%v", chunkCount, chunk.Token, chunk.Done)
			if chunk.Done {
				t.Log("Stream completed normally")
				return
			}
		case <-timeout:
			t.Fatalf("Timeout waiting for chunks (received %d)", chunkCount)
		}
	}

	t.Log("Successfully connected and received chunks")
}

func TestAPIKeyLocation(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	settingsPath := homeDir + "/.claude/settings.json"

	t.Logf("Looking for API key in: %s", settingsPath)

	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Logf("File does not exist: %s", settingsPath)
	} else {
		t.Logf("File exists: %s", settingsPath)
	}

	// Check env var
	envKey := os.Getenv("ANTHROPIC_AUTH_TOKEN")
	if envKey != "" {
		t.Logf("ANTHROPIC_AUTH_TOKEN env var is set (length: %d)", len(envKey))
	} else {
		t.Log("ANTHROPIC_AUTH_TOKEN env var is NOT set")
	}
}
