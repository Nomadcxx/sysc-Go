package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Nomadcxx/sysc-Go/agent"
	"github.com/Nomadcxx/sysc-Go/agent/floyd"
	"github.com/Nomadcxx/sysc-Go/agent/loop"
	agenttools "github.com/Nomadcxx/sysc-Go/agent/tools"
	"github.com/Nomadcxx/sysc-Go/agent/prompt"
)

func main() {
	// Load API key
	apiKey := os.Getenv("ANTHROPIC_AUTH_TOKEN")
	if apiKey == "" {
		apiKey = os.Getenv("GLM_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("ZHIPU_API_KEY")
	}
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "❌ Error: No API key found. Set ANTHROPIC_AUTH_TOKEN or GLM_API_KEY")
		os.Exit(1)
	}

	// Create client
	client := agent.NewProxyClient(apiKey, "https://api.z.ai/api/anthropic")
	client.SetModel("claude-opus-4")

	// Initialize components
	toolExecutor := agenttools.NewExecutor(30)
	protocolMgr := floyd.NewProtocolManager(".floyd")
	safety := floyd.NewSafetyEnforcer()
	loopAgent := loop.NewLoopAgent(client, toolExecutor)

	// Load system prompt
	systemPrompt := prompt.LoadSystemPrompt("")
	if systemPrompt == "" {
		systemPrompt = prompt.DefaultSystemPrompt
	}

	// Conversation
	var messages []agent.Message

	// Setup Ctrl+C handler
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\n✋ Exiting...")
		cancel()
	}()

	// Print banner
	fmt.Println("\n════════════════════════════════════════════════════════════")
	fmt.Println("  FLOYD CLI - Simple Interface")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("  Type your message and press Enter.")
	fmt.Println("  Ctrl+C to exit.")
	fmt.Println("════════════════════════════════════════════════════════════\n")

	reader := bufio.NewReader(os.Stdin)

	for {
		// Show prompt
		fmt.Print("❯ ")

		// Read input
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(input, "/") {
			handleCommand(input, protocolMgr)
			continue
		}

		// Check safety
		if err := safety.CheckAction(input); err != nil {
			fmt.Printf("🚫 SAFETY: %v\n", err)
			continue
		}

		// Add user message
		messages = append(messages, agent.Message{
			Role:    "user",
			Content: input,
		})

		// Build request
		req := agent.ChatRequest{
			Messages:    messages,
			MaxTokens:   4096,
			Temperature: 0.7,
		}

		// Add system prompt and protocol
		req = protocolMgr.EnhancedChatRequest(req, "")

		// Show thinking
		fmt.Print("🤔 Thinking...")

		// Call agent
		response, err := loopAgent.RunToolLoop(ctx, req)
		fmt.Println() // New line after "Thinking..."

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		// Show response
		fmt.Printf("\n🎸 FLOYD:\n%s\n\n", response)

		// Add to conversation
		messages = append(messages, agent.Message{
			Role:    "assistant",
			Content: response,
		})

		// Log action
		protocolMgr.LogAction("chat", "completed", "ready")
	}
}

func handleCommand(input string, pm *floyd.ProtocolManager) {
	cmd := strings.TrimPrefix(input, "/")
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	switch strings.ToLower(parts[0]) {
	case "exit", "quit", "q":
		fmt.Println("👋 Goodbye!")
		os.Exit(0)

	case "clear", "cls":
		fmt.Print("\033[H\033[2J") // Clear terminal
		fmt.Println("✅ Cleared.")

	case "help", "h", "?":
		fmt.Println(`
╔════════════════════════════════════════════════════════════╗
║                      FLOYD CLI - HELP                       ║
╚════════════════════════════════════════════════════════════╝

COMMANDS:
  /exit, /quit     - Exit FLOYD
  /clear           - Clear terminal
  /status          - Show workspace status
  /init            - Initialize .floyd/ workspace
  /tools           - List available tools

Just type your message and press Enter to chat with FLOYD.
`)

	case "status":
		status := pm.WorkspaceStatus()
		fmt.Println("\n📁 Workspace Status:")
		for k, v := range status {
			fmt.Printf("  %s: %s\n", k, v)
		}
		fmt.Println()

	case "init":
		if err := prompt.InitializeFloydDir(); err != nil {
			fmt.Printf("❌ Init failed: %v\n", err)
		} else {
			fmt.Println("✅ .floyd/ workspace initialized")
		}

	case "tools":
		tools := agenttools.ListTools()
		fmt.Printf("\n🔧 Available Tools (%d):\n", len(tools))
		for _, t := range tools {
			fmt.Printf("  • %s\n", t)
		}
		fmt.Println()

	default:
		fmt.Printf("❓ Unknown command: %s (type /? for help)\n", parts[0])
	}
}
