package floydtools

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// StreamMsg represents a message from a tool's execution.
type StreamMsg struct {
	Chunk     string
	Status    string
	Done      bool
	LogAction string
	LogResult string
	LogNext   string
}

// Tool is the interface that all Floyd tools must implement.
type Tool interface {
	// Name returns the tool's identifier.
	Name() string

	// Description returns a human-readable description.
	Description() string

	// Validate returns an error if the input is not valid for this tool.
	Validate(input string) error

	// Run executes the tool with the given input.
	// It returns a function that sends StreamMsg to the provided channel.
	// The returned function is designed to be run as a goroutine.
	Run(input string) func(chan<- StreamMsg)

	// FrameDelay controls the streaming speed (for tools that output incrementally).
	// Return 0 for tools that output all at once.
	FrameDelay() time.Duration
}

// Runner is the factory function signature for creating tools.
type Runner func() Tool

// Command is the tea.Cmd type for tool execution.
type Command = tea.Cmd

// ParseInput attempts to parse the input as JSON and extract the value for the given keys.
// If it's not JSON or keys aren't found, it returns the trimmed raw input.
func ParseInput(input string, keys ...string) string {
	input = strings.TrimSpace(input)

	// Strip markdown code blocks if present
	if strings.HasPrefix(input, "```") {
		// Remove opening fence (including language identifier usually attached like ```json)
		if idx := strings.Index(input, "\n"); idx != -1 {
			input = input[idx+1:]
		} else {
			// Weird single line case, just strip first 3 chars
			input = strings.TrimPrefix(input, "```")
		}

		// Remove closing fence
		if idx := strings.LastIndex(input, "```"); idx != -1 {
			input = input[:idx]
		}

		input = strings.TrimSpace(input)
	}

	if !strings.HasPrefix(input, "{") {
		return input
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return input
	}

	for _, key := range keys {
		if val, ok := data[key]; ok {
			if s, ok := val.(string); ok {
				return s
			}
			return fmt.Sprintf("%v", val)
		}
	}

	// Fallback to searching for ANY string value if specific keys aren't found
	for _, val := range data {
		if s, ok := val.(string); ok {
			return s
		}
	}

	return input
}
