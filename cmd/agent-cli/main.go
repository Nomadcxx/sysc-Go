// FLOYD CLI Test Harness - Week 1, Task 3
// A simple CLI for testing the headless Agent Engine without the TUI.
//
// Usage:
//   go run cmd/agent-cli/main.go "list all go files"
//   go run cmd/agent-cli/main.go "read README.md and summarize"
//   go run cmd/agent-cli/main.go "create hello.txt with content 'Hello World'"
//
// Environment Variables (in order of priority):
//   ANTHROPIC_AUTH_TOKEN - Primary API key
//   GLM_API_KEY          - Fallback API key
//   ZHIPU_API_KEY        - Another fallback
//
// The CLI streams events in real-time showing:
//   - Planning phase with tool selection
//   - Tool execution with results
//   - Final response

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Nomadcxx/sysc-Go/agent"
	"github.com/Nomadcxx/sysc-Go/agent/floyd"
	agentloop "github.com/Nomadcxx/sysc-Go/agent/loop"
	agenttools "github.com/Nomadcxx/sysc-Go/agent/tools"
	"github.com/Nomadcxx/sysc-Go/agent/prompt"
)

const (
	// Version of the CLI Test Harness
	version = "1.0.0"

	// Default timeout for the agent execution
	defaultTimeout = 5 * time.Minute

	// Default model to use
	defaultModel = "claude-opus-4"
)

var (
	// ANSI color codes for output
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// Event represents a CLI event for output formatting
type Event struct {
	Type    string
	Message string
	Data    map[string]interface{}
}

// CLIConfig holds configuration for the CLI
type CLIConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	Timeout     time.Duration
	Verbose     bool
	ShowTokens  bool
}

func main() {
	// Parse arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	goal := strings.TrimSpace(os.Args[1])

	// Handle help/flags
	if goal == "" || goal == "help" || goal == "--help" || goal == "-h" {
		printUsage()
		os.Exit(0)
	}

	if goal == "version" || goal == "--version" || goal == "-v" {
		fmt.Printf("FLOYD Agent CLI v%s\n", version)
		os.Exit(0)
	}

	// Load configuration
	config := loadConfig()

	// Validate API key
	if config.APIKey == "" {
		printError("No API key found. Set ANTHROPIC_AUTH_TOKEN, GLM_API_KEY, or ZHIPU_API_KEY environment variable.")
		os.Exit(1)
	}

	// Create context with cancellation
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Setup graceful shutdown on Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		printEvent("CANCEL", "Interrupted by user. Cleaning up...", nil)
		cancel()
	}()

	// Run the agent
	if err := runAgent(ctx, goal, config); err != nil {
		printError("Agent execution failed: %v", err)
		os.Exit(1)
	}
}

// runAgent executes the agent with the given goal
func runAgent(ctx context.Context, goal string, config CLIConfig) error {
	// Print header
	printHeader(goal, config)

	// Step 1: Initialize components
	printEvent("INIT", "Creating agent components...", nil)

	client := agent.NewProxyClient(config.APIKey, config.BaseURL)
	client.SetModel(config.Model)

	toolExecutor := agenttools.NewExecutor(30 * time.Second)
	loopAgent := agentloop.NewLoopAgent(client, toolExecutor)
	protocolMgr := floyd.NewProtocolManager(".floyd")

	// Step 2: Build the request
	printEvent("PLANNING", "Analyzing goal and selecting tools...", nil)

	messages := []agent.Message{
		{
			Role:    "user",
			Content: goal,
		},
	}

	req := agent.ChatRequest{
		Messages:    messages,
		MaxTokens:   8192,
		Temperature: 0.7,
	}

	// Enhance with FLOYD protocol
	req = protocolMgr.EnhancedChatRequest(req, "")

	// Show selected tools
	if len(req.Tools) > 0 {
		toolNames := make([]string, 0, len(req.Tools))
		for _, t := range req.Tools {
			toolNames = append(toolNames, t.Name)
		}
		printEvent("TOOLS", fmt.Sprintf("Available: %s", strings.Join(toolNames, ", ")), nil)
	}

	// Step 3: Execute with streaming
	printEvent("EXECUTING", "Running agent loop...", nil)

	eventChan, err := loopAgent.RunToolLoopStreaming(ctx, req)
	if err != nil {
		return fmt.Errorf("start streaming: %w", err)
	}

	// Step 4: Process events
	return processEvents(ctx, eventChan, config)
}

// processEvents handles streaming events from the agent
func processEvents(ctx context.Context, eventChan <-chan agentloop.StreamingEvent, config CLIConfig) error {
	var (
		lastType     agentloop.EventType
		tokenBuffer  strings.Builder
		toolCount    int
		resultCount  int
		totalTokens  int
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event, ok := <-eventChan:
			if !ok {
				printEvent("DONE", "Stream ended", nil)
				return nil
			}

			switch event.Type {
			case agentloop.EventTypeToken:
				// Accumulate tokens for cleaner output
				tokenBuffer.WriteString(event.Text)
				if config.Verbose || tokenBuffer.Len() > 100 {
					// Flush buffer periodically or in verbose mode
					fmt.Print(tokenBuffer.String())
					tokenBuffer.Reset()
				}

			case agentloop.EventTypeToolUse:
				toolCount++
				printEvent("TOOL_USE", fmt.Sprintf("Tool called: %s", event.ToolName), map[string]interface{}{
					"id": event.ToolID,
				})

			case agentloop.EventTypeToolExecuting:
				printEvent("TOOL_EXEC", event.ToolCall, nil)

			case agentloop.EventTypeToolResult:
				resultCount++
				status := "success"
				if event.IsError {
					status = "error"
					printEvent("TOOL_ERROR", fmt.Sprintf("%s: %s", event.ToolName, truncateString(event.ToolContent, 100)), nil)
				} else if config.Verbose {
					printEvent("TOOL_RESULT", fmt.Sprintf("%s: %s", event.ToolName, truncateString(event.ToolContent, 100)), nil)
				} else {
					// Just show success indicator
					fmt.Printf("   ✓ %s completed\n", event.ToolName)
				}
				mapSet(event.Data, "status", status)

			case agentloop.EventTypeIteration:
				if event.Iteration > 0 {
					fmt.Printf("\n   [Iteration %d]\n", event.Iteration)
				}

			case agentloop.EventTypeContentBlockDone:
				// Flush any remaining tokens
				if tokenBuffer.Len() > 0 {
					fmt.Print(tokenBuffer.String())
					tokenBuffer.Reset()
				}

			case agentloop.EventTypeDone:
				// Flush remaining tokens
				if tokenBuffer.Len() > 0 {
					fmt.Print(tokenBuffer.String())
					tokenBuffer.Reset()
				}

				// Print final response section header
				fmt.Printf("\n")
				printEvent("RESPONSE", "", nil)
				fmt.Printf("%s%s%s\n", colorCyan, event.Text, colorReset)

				// Print summary
				printSummary(toolCount, resultCount, event.StopReason)

			case agentloop.EventTypeError:
				// Flush remaining tokens
				if tokenBuffer.Len() > 0 {
					fmt.Print(tokenBuffer.String())
					tokenBuffer.Reset()
				}
				return fmt.Errorf("agent error: %w", event.Error)
			}

			lastType = event.Type
		}
	}
}

// printHeader displays the CLI header
func printHeader(goal string, config CLIConfig) {
	fmt.Printf("\n%s╔════════════════════════════════════════════════════════════╗%s\n", colorCyan, colorReset)
	fmt.Printf("%s║%s  FLOYD Agent Engine - CLI Test Harness v%s%s           %s║%s\n", colorCyan, colorReset, version, colorCyan, colorReset, colorCyan, colorReset)
	fmt.Printf("%s╚════════════════════════════════════════════════════════════╝%s\n", colorCyan, colorReset)
	fmt.Printf("\n%sGoal:%s %s\n", colorYellow, colorReset, truncateString(goal, 60))
	fmt.Printf("%sModel:%s %s\n", colorGray, colorReset, config.Model)
	fmt.Printf("%sTimeout:%s %s\n\n", colorGray, colorReset, config.Timeout)
}

// printEvent displays a formatted event
func printEvent(eventType, message string, data map[string]interface{}) {
	var icon, color string

	switch eventType {
	case "INIT":
		icon, color = "", colorBlue
	case "PLANNING":
		icon, color = "", colorPurple
	case "TOOLS":
		icon, color = "", colorGray
	case "EXECUTING":
		icon, color = "", colorGreen
	case "TOOL_USE":
		icon, color = "  ", colorYellow
	case "TOOL_EXEC":
		icon, color = "  ", colorGray
	case "TOOL_RESULT":
		icon, color = "  ", colorGreen
	case "TOOL_ERROR":
		icon, color = "  ", colorRed
	case "RESPONSE":
		icon, color = "", colorCyan
	case "DONE":
		icon, color = "", colorGreen
	case "CANCEL":
		icon, color = "", colorYellow
	case "ERROR":
		icon, color = "", colorRed
	default:
		icon, color = "", colorReset
	}

	if icon != "" {
		fmt.Printf("%s%s%s %s%s\n", color, icon, colorReset, message, colorReset)
	} else if eventType != "" && message != "" {
		fmt.Printf("%s[%s]%s %s\n", color, eventType, colorReset, message)
	} else if eventType == "RESPONSE" {
		fmt.Printf("%s─── Response ───%s\n", color, colorReset)
	}
}

// printSummary displays execution summary
func printSummary(toolCount, resultCount int, stopReason string) {
	fmt.Printf("\n%s╔════════════════════════════════════════════════════════════╗%s\n", colorGreen, colorReset)
	fmt.Printf("%s║%s  Execution Summary%s                                           %s║%s\n", colorGreen, colorReset, colorGreen, colorReset, colorGreen, colorReset)
	fmt.Printf("%s╠════════════════════════════════════════════════════════════╣%s\n", colorGreen, colorReset)
	fmt.Printf("%s║%s  Tools called:     %s%-3d%s                                      %s║%s\n", colorGreen, colorReset, colorYellow, toolCount, colorReset, colorGreen, colorReset)
	fmt.Printf("%s║%s  Tools completed:  %s%-3d%s                                      %s║%s\n", colorGreen, colorReset, colorGreen, resultCount, colorReset, colorGreen, colorReset)
	fmt.Printf("%s║%s  Stop reason:      %s%-15s%s                            %s║%s\n", colorGreen, colorReset, colorCyan, stopReason, colorReset, colorGreen, colorReset)
	fmt.Printf("%s╚════════════════════════════════════════════════════════════╝%s\n", colorGreen, colorReset)
	fmt.Println()
}

// printError displays an error message
func printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s[ERROR]%s %s\n", colorRed, colorReset, fmt.Sprintf(format, args...))
}

// printUsage displays usage information
func printUsage() {
	fmt.Printf(`
FLOYD Agent CLI - Test Harness v%s

USAGE:
  go run cmd/agent-cli/main.go "<your goal>"
  ./agent-cli "<your goal>"

EXAMPLES:
  ./agent-cli "list all go files in the current directory"
  ./agent-cli "read the README and summarize it"
  ./agent-cli "create a file called hello.txt with content 'Hello World'"
  ./agent-cli "find all functions that return error in ./agent"

FLAGS:
  -v, --verbose     Show verbose output including all tool results
  -h, --help        Show this help message
  --version         Show version information

ENVIRONMENT VARIABLES:
  ANTHROPIC_AUTH_TOKEN    Primary API key (recommended)
  GLM_API_KEY            Fallback API key
  ZHIPU_API_KEY          Another fallback API key

OUTPUT FORMAT:
  The CLI streams events in real-time showing:
    - Planning phase with selected tools
    - Tool execution with results
    - Final response from the agent

`, version)
}

// loadConfig loads configuration from environment variables
func loadConfig() CLIConfig {
	return CLIConfig{
		APIKey:     getAPIKey(),
		BaseURL:    getEnv("FLOYD_API_BASE_URL", "https://api.z.ai/api/anthropic"),
		Model:      getEnv("FLOYD_MODEL", defaultModel),
		Timeout:    defaultTimeout,
		Verbose:    getEnvBool("FLOYD_VERBOSE", false),
		ShowTokens: getEnvBool("FLOYD_SHOW_TOKENS", false),
	}
}

// getAPIKey retrieves the API key from environment variables
func getAPIKey() string {
	if key := os.Getenv("ANTHROPIC_AUTH_TOKEN"); key != "" {
		return key
	}
	if key := os.Getenv("GLM_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("ZHIPU_API_KEY"); key != "" {
		return key
	}
	return ""
}

// getEnv retrieves an environment variable with a fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvBool retrieves a boolean environment variable
func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value == "1" || strings.ToLower(value) == "true" || strings.ToLower(value) == "yes"
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// mapSet sets a value in a map (helper for missing key)
func mapSet(m map[string]interface{}, key string, value interface{}) {
	if m == nil {
		m = make(map[string]interface{})
	}
	m[key] = value
}
