package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Nomadcxx/sysc-Go/agent"
	"github.com/Nomadcxx/sysc-Go/agent/message"
	"github.com/Nomadcxx/sysc-Go/tui/floydtools"
)

// Executor handles tool execution from LLM tool use requests
type Executor struct {
	timeout time.Duration
}

// NewExecutor creates a new tool executor
func NewExecutor(timeout time.Duration) *Executor {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Executor{
		timeout: timeout,
	}
}

// ExecuteToolCall executes a single tool call and returns the result
func (e *Executor) ExecuteToolCall(ctx context.Context, toolCall agent.ToolCall) (message.ToolResultBlock, error) {
	// Get the tool from global registry
	tool, err := floydtools.Get(toolCall.Name)
	if err != nil {
		return message.ToolResultBlock{
			ToolUseID: toolCall.ID,
			Content:   fmt.Sprintf("Error getting tool: %v", err),
			IsError:   true,
		}, nil
	}

	// Prepare input - convert InputMap to JSON string
	inputStr := toolCall.Input
	if inputStr == "" && toolCall.InputMap != nil {
		inputBytes, err := json.Marshal(toolCall.InputMap)
		if err != nil {
			return message.ToolResultBlock{
				ToolUseID: toolCall.ID,
				Content:   fmt.Sprintf("Error marshaling tool input: %v", err),
				IsError:   true,
			}, nil
		}
		inputStr = string(inputBytes)
	}

	// Validate input
	if err := tool.Validate(inputStr); err != nil {
		return message.ToolResultBlock{
			ToolUseID: toolCall.ID,
			Content:   fmt.Sprintf("Validation error: %v", err),
			IsError:   true,
		}, nil
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	result, err := e.executeTool(ctx, tool, inputStr)
	if err != nil {
		return message.ToolResultBlock{
			ToolUseID: toolCall.ID,
			Content:   fmt.Sprintf("Execution error: %v", err),
			IsError:   true,
		}, nil
	}

	return message.ToolResultBlock{
		ToolUseID: toolCall.ID,
		Content:   result,
		IsError:   false,
	}, nil
}

// executeTool runs the tool and captures its output
func (e *Executor) executeTool(ctx context.Context, tool floydtools.Tool, input string) (string, error) {
	// Create a channel for the tool's stream messages
	msgChan := make(chan floydtools.StreamMsg, 100)

	// Start the tool's goroutine
	runFunc := tool.Run(input)
	go runFunc(msgChan)

	// Collect output
	var output strings.Builder
	var errMsg string
	done := false

	for !done {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("tool execution timed out or cancelled: %w", ctx.Err())

		case msg, ok := <-msgChan:
			if !ok {
				done = true
				break
			}

			// Append chunk content
			if msg.Chunk != "" {
				output.WriteString(msg.Chunk)
			}

			// Include status in output when it contains useful info
			if msg.Status != "" && !msg.Done {
				output.WriteString(msg.Status)
				output.WriteString("\n")
			}

			// Check for completion
			if msg.Done {
				done = true
			}

			// Check for error in status - multiple patterns
			if strings.HasPrefix(msg.Status, "error:") {
				errMsg = strings.TrimPrefix(msg.Status, "error:")
			} else if strings.Contains(msg.Status, "failed:") || strings.Contains(msg.Status, "Error") {
				errMsg = msg.Status
			}
		}
	}

	result := output.String()

	// Return error if any
	if errMsg != "" {
		return result, fmt.Errorf("%s", errMsg)
	}

	return result, nil
}

// ExecuteToolCalls executes multiple tool calls in parallel
// Returns results in the same order as the input tool calls
func (e *Executor) ExecuteToolCalls(ctx context.Context, toolCalls []agent.ToolCall) []message.ToolResultBlock {
	results := make([]message.ToolResultBlock, len(toolCalls))
	var wg sync.WaitGroup

	for i, toolCall := range toolCalls {
		wg.Add(1)
		go func(idx int, tc agent.ToolCall) {
			defer wg.Done()

			result, err := e.ExecuteToolCall(ctx, tc)
			if err != nil {
				result = message.ToolResultBlock{
					ToolUseID: tc.ID,
					Content:   fmt.Sprintf("Tool execution failed: %v", err),
					IsError:   true,
				}
			}

			results[idx] = result
		}(i, toolCall)
	}

	wg.Wait()
	return results
}

// CreateToolResultMessage creates a user message containing tool results
// This is ready to be sent back to the API
func CreateToolResultMessage(results []message.ToolResultBlock) message.Message {
	contentBlocks := make([]message.ContentBlock, len(results))

	for i, result := range results {
		contentBlocks[i] = message.ContentBlock{
			Type:      "tool_result",
			ToolUseID: result.ToolUseID,
			Content:   result.Content,
			IsError:   result.IsError,
		}
	}

	return message.Message{
		Role:    "user",
		Content: contentBlocks,
	}
}

// ParseToolInput parses tool input from JSON string
func ParseToolInput(inputStr string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(inputStr), &result); err != nil {
		return nil, fmt.Errorf("parse tool input: %w", err)
	}
	return result, nil
}

// FormatToolInput formats tool input as JSON string
func FormatToolInput(input map[string]any) (string, error) {
	result, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("format tool input: %w", err)
	}
	return string(result), nil
}

// ListTools returns all available tool names
func ListTools() []string {
	return floydtools.List()
}

// GetToolDescriptions returns descriptions for all tools
func GetToolDescriptions() map[string]string {
	tools := floydtools.AllTools()
	descriptions := make(map[string]string, len(tools))

	for _, tool := range tools {
		descriptions[tool.Name()] = tool.Description()
	}

	return descriptions
}
