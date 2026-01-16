//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Nomadcxx/sysc-Go/agent"
	"github.com/Nomadcxx/sysc-Go/agent/floyd"
	"github.com/Nomadcxx/sysc-Go/agent/loop"
	"github.com/Nomadcxx/sysc-Go/agent/tools"
)

func main() {
	fmt.Println("=== FLOYD Agent Integration Test ===")
	fmt.Println()

	// Create client
	client := agent.NewProxyClient("", "")
	if client == nil {
		fmt.Println("ERROR: Failed to create client")
		return
	}
	fmt.Println("✓ Client created")

	// Create tool executor
	executor := tools.NewExecutor(5 * time.Minute)
	fmt.Println("✓ Tool executor created")

	// Create loop agent
	loopAgent := loop.NewLoopAgent(client, executor)
	fmt.Println("✓ Loop agent created")

	// Create protocol manager
	protocolMgr := floyd.NewProtocolManager(".floyd")
	fmt.Println("✓ Protocol manager created")

	// Build request
	messages := []agent.Message{
		{Role: "user", Content: "Say hello in exactly 5 words."},
	}

	req := agent.ChatRequest{
		Messages:    messages,
		MaxTokens:   100,
		Temperature: 0.7,
		Stream:      true,
	}

	// Inject protocol
	req = protocolMgr.EnhancedChatRequest(req, "")
	fmt.Printf("✓ Request built with %d messages\n", len(req.Messages))

	// Run streaming
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println()
	fmt.Println("Starting streaming request...")

	eventChan, err := loopAgent.RunToolLoopStreaming(ctx, req)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	fmt.Println("✓ Stream started")
	fmt.Println()
	fmt.Print("Response: ")

	tokenCount := 0
	for event := range eventChan {
		switch event.Type {
		case loop.EventTypeToken:
			fmt.Print(event.Text)
			tokenCount++
		case loop.EventTypeToolUse:
			fmt.Printf("\n[Tool: %s]\n", event.ToolName)
		case loop.EventTypeDone:
			fmt.Println()
			fmt.Println()
			fmt.Printf("✓ Done! Received %d tokens\n", tokenCount)
			return
		case loop.EventTypeError:
			fmt.Printf("\nERROR: %v\n", event.Error)
			return
		}
	}

	fmt.Println()
	fmt.Printf("✓ Stream ended. Received %d tokens\n", tokenCount)
}
