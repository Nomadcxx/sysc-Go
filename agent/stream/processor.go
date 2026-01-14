package stream

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Nomadcxx/sysc-Go/agent"
)

// Processor handles processing of streaming chunks from Anthropic API
type Processor struct {
	accumulatedText strings.Builder
	toolCalls       []agent.ToolCall
	activeTool      *activeToolCall
}

type activeToolCall struct {
	ID        string
	Name      string
	InputJSON strings.Builder
}

// NewProcessor creates a new stream processor
func NewProcessor() *Processor {
	return &Processor{
		toolCalls: make([]agent.ToolCall, 0),
	}
}

// ProcessChunk processes a single stream chunk
// Returns true if the stream is complete
func (p *Processor) ProcessChunk(chunk agent.StreamChunk) bool {
	// Handle text tokens
	if chunk.Token != "" {
		p.accumulatedText.WriteString(chunk.Token)
	}

	// Handle tool use completion
	if chunk.ToolUseComplete && chunk.ToolCall != nil {
		p.toolCalls = append(p.toolCalls, *chunk.ToolCall)
	}

	// Stream is done when StopReason is set or Done flag is true
	return chunk.Done || chunk.StopReason != ""
}

// GetText returns all accumulated text content
func (p *Processor) GetText() string {
	return p.accumulatedText.String()
}

// GetToolCalls returns all completed tool calls
func (p *Processor) GetToolCalls() []agent.ToolCall {
	return p.toolCalls
}

// HasToolCalls returns true if any tool calls were made
func (p *Processor) HasToolCalls() bool {
	return len(p.toolCalls) > 0
}

// GetStopReason returns the stop reason if stream ended
func (p *Processor) GetStopReason() string {
	// This would be tracked by the caller
	return ""
}

// Reset clears all accumulated state
func (p *Processor) Reset() {
	p.accumulatedText.Reset()
	p.toolCalls = make([]agent.ToolCall, 0)
	p.activeTool = nil
}

// AccumulateToolInput accumulates partial JSON deltas for tool input
// Used when reconstructing streaming tool calls
func AccumulateToolInput(deltas []string) (map[string]any, error) {
	fullJSON := strings.Join(deltas, "")

	var result map[string]any
	if err := json.Unmarshal([]byte(fullJSON), &result); err != nil {
		return nil, fmt.Errorf("parse accumulated tool input: %w", err)
	}

	return result, nil
}

// ExtractTextContent extracts all text from a stream channel
// This is a convenience function for simple streaming use cases
func ExtractTextContent(ch <-chan agent.StreamChunk) string {
	var sb strings.Builder

	for chunk := range ch {
		if chunk.Error != nil {
			return ""
		}
		if chunk.Token != "" {
			sb.WriteString(chunk.Token)
		}
		if chunk.Done {
			break
		}
	}

	return sb.String()
}

// ExtractToolUseBlocks extracts all tool_use blocks from stream chunks
// Returns the complete tool calls and the full text response
func ExtractToolUseBlocks(ch <-chan agent.StreamChunk) ([]agent.ToolCall, string, error) {
	processor := NewProcessor()

	for chunk := range ch {
		if chunk.Error != nil {
			return nil, "", fmt.Errorf("stream error: %w", chunk.Error)
		}

		processor.ProcessChunk(chunk)

		if chunk.Done {
			break
		}
	}

	return processor.GetToolCalls(), processor.GetText(), nil
}

// DetectToolUseEnd checks if the stream has ended with tool_use as stop reason
func DetectToolUseEnd(ch <-chan agent.StreamChunk) bool {
	for chunk := range ch {
		if chunk.StopReason == "tool_use" {
			return true
		}
		if chunk.Done {
			return chunk.StopReason == "tool_use"
		}
	}
	return false
}

// ProcessToolUseBlocks is a higher-level function that processes a stream
// and returns either tool calls or text, depending on what the model generated
func ProcessToolUseBlocks(ch <-chan agent.StreamChunk) <-chan ProcessedResult {
	resultChan := make(chan ProcessedResult, 1)

	go func() {
		defer close(resultChan)

		processor := NewProcessor()

		for chunk := range ch {
			if chunk.Error != nil {
				resultChan <- ProcessedResult{Error: chunk.Error}
				return
			}

			done := processor.ProcessChunk(chunk)

			if done {
				break
			}
		}

		resultChan <- ProcessedResult{
			ToolCalls: processor.GetToolCalls(),
			Text:      processor.GetText(),
			StopReason: processor.GetStopReason(),
		}
	}()

	return resultChan
}

// ProcessedResult represents the final result of processing a stream
type ProcessedResult struct {
	ToolCalls  []agent.ToolCall
	Text       string
	StopReason string
	Error      error
}

// IsToolUse returns true if the result contains tool calls
func (r ProcessedResult) IsToolUse() bool {
	return len(r.ToolCalls) > 0
}

// IsTextOnly returns true if the result is text only (no tool calls)
func (r ProcessedResult) IsTextOnly() bool {
	return len(r.ToolCalls) == 0 && r.Text != ""
}

// GetFirstToolCall returns the first tool call, or nil if none
func (r ProcessedResult) GetFirstToolCall() *agent.ToolCall {
	if len(r.ToolCalls) == 0 {
		return nil
	}
	return &r.ToolCalls[0]
}
