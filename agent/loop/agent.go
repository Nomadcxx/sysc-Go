package loop

import (
	"context"
	"fmt"
	"strings"

	"github.com/Nomadcxx/sysc-Go/agent"
	"github.com/Nomadcxx/sysc-Go/agent/message"
	"github.com/Nomadcxx/sysc-Go/agent/stream"
	"github.com/Nomadcxx/sysc-Go/agent/tools"
)

const (
	// DefaultMaxIterations is the maximum number of tool use iterations
	DefaultMaxIterations = 10
)

// LoopAgent handles the tool calling conversation loop
type LoopAgent struct {
	client   *agent.ProxyClient
	executor *tools.Executor
	maxIter  int
}

// NewLoopAgent creates a new tool calling loop agent
func NewLoopAgent(client *agent.ProxyClient, executor *tools.Executor) *LoopAgent {
	maxIter := DefaultMaxIterations
	if executor == nil {
		executor = tools.NewExecutor(0)
	}

	return &LoopAgent{
		client:   client,
		executor: executor,
		maxIter:  maxIter,
	}
}

// SetMaxIterations sets the maximum number of tool use iterations
func (a *LoopAgent) SetMaxIterations(n int) {
	a.maxIter = n
}

// RunToolLoop executes the full tool calling conversation loop
// It sends the initial request, processes any tool calls, and continues
// until the model is done or max iterations is reached
func (a *LoopAgent) RunToolLoop(ctx context.Context, req agent.ChatRequest) (string, error) {
	messages := req.Messages

	for iteration := 0; iteration < a.maxIter; iteration++ {
		// Send request to API
		req.Messages = messages
		chunkChan, err := a.client.StreamChat(ctx, req)
		if err != nil {
			return "", fmt.Errorf("stream chat: %w", err)
		}

		// Process the stream
		toolCalls, text, err := stream.ExtractToolUseBlocks(chunkChan)
		if err != nil {
			return "", fmt.Errorf("extract tool use: %w", err)
		}

		// If no tool calls, we're done - return the text
		if len(toolCalls) == 0 {
			return text, nil
		}

		// Execute tools
		results := a.executeTools(ctx, toolCalls)

		// Append assistant message with tool uses (converted to agent.Message)
		assistMsg := a.createAssistantMessage(toolCalls)
		messages = append(messages, convertToAgentMessage(assistMsg))

		// Append user message with tool results (converted to agent.Message)
		resultMsg := tools.CreateToolResultMessage(results)
		messages = append(messages, convertToAgentMessage(resultMsg))
	}

	return "", fmt.Errorf("max iterations (%d) exceeded", a.maxIter)
}

// RunToolLoopStreaming runs the tool loop with streaming output
// Returns a channel of text chunks and tool use events
func (a *LoopAgent) RunToolLoopStreaming(ctx context.Context, req agent.ChatRequest) (<-chan StreamingEvent, error) {
	eventChan := make(chan StreamingEvent, 100)

	go func() {
		defer close(eventChan)

		messages := req.Messages

		for iteration := 0; iteration < a.maxIter; iteration++ {
			// Send iteration start event
			eventChan <- StreamingEvent{
				Type:       EventTypeIteration,
				Iteration:  iteration,
				Text:       fmt.Sprintf("Iteration %d", iteration),
			}

			// Send request to API
			currentReq := req
			currentReq.Messages = messages

			chunkChan, err := a.client.StreamChat(ctx, currentReq)
			if err != nil {
				eventChan <- StreamingEvent{
					Type:  EventTypeError,
					Error: fmt.Errorf("stream chat: %w", err),
				}
				return
			}

			// Process stream with events
			toolCalls, text, stopReason, err := a.processStreamWithEvents(ctx, chunkChan, eventChan)
			if err != nil {
				eventChan <- StreamingEvent{
					Type:  EventTypeError,
					Error: err,
				}
				return
			}

			// If no tool calls, we're done
			if len(toolCalls) == 0 {
				eventChan <- StreamingEvent{
					Type:       EventTypeDone,
					StopReason: stopReason,
					Text:       text,
				}
				return
			}

			// Execute tools with progress events
			eventChan <- StreamingEvent{
				Type:        EventTypeToolExecuting,
				ToolCall:    fmt.Sprintf("%d tools", len(toolCalls)),
			}

			results := a.executeTools(ctx, toolCalls)

			// Send tool result events
			for i, result := range results {
				eventChan <- StreamingEvent{
					Type:        EventTypeToolResult,
					ToolName:    toolCalls[i].Name,
					ToolContent: result.Content,
					IsError:     result.IsError,
				}
			}

			// Append assistant message with tool uses (converted to agent.Message)
			assistMsg := a.createAssistantMessage(toolCalls)
			messages = append(messages, convertToAgentMessage(assistMsg))

			// Append user message with tool results (converted to agent.Message)
			resultMsg := tools.CreateToolResultMessage(results)
			messages = append(messages, convertToAgentMessage(resultMsg))
		}

		eventChan <- StreamingEvent{
			Type:  EventTypeError,
			Error: fmt.Errorf("max iterations (%d) exceeded", a.maxIter),
		}
	}()

	return eventChan, nil
}

// processStreamWithEvents processes a stream and sends events
func (a *LoopAgent) processStreamWithEvents(ctx context.Context, chunkChan <-chan agent.StreamChunk, eventChan chan<- StreamingEvent) ([]agent.ToolCall, string, string, error) {
	processor := stream.NewProcessor()
	var fullText strings.Builder

	for chunk := range chunkChan {
		if chunk.Error != nil {
			return nil, "", "", chunk.Error
		}

		// Send text tokens
		if chunk.Token != "" {
			fullText.WriteString(chunk.Token)
			eventChan <- StreamingEvent{
				Type: EventTypeToken,
				Text: chunk.Token,
			}
		}

		// Check for tool use completion
		if chunk.ToolUseComplete && chunk.ToolCall != nil {
			eventChan <- StreamingEvent{
				Type:     EventTypeToolUse,
				ToolName: chunk.ToolCall.Name,
				ToolID:   chunk.ToolCall.ID,
			}
		}

		// Check if done
		if chunk.Done {
			// Collect final state
			processor.ProcessChunk(chunk)

			eventChan <- StreamingEvent{
				Type:       EventTypeContentBlockDone,
				StopReason: chunk.StopReason,
			}

			return processor.GetToolCalls(), fullText.String(), chunk.StopReason, nil
		}

		processor.ProcessChunk(chunk)
	}

	return processor.GetToolCalls(), fullText.String(), "", nil
}

// executeTools executes all tool calls in parallel
func (a *LoopAgent) executeTools(ctx context.Context, toolCalls []agent.ToolCall) []message.ToolResultBlock {
	return a.executor.ExecuteToolCalls(ctx, toolCalls)
}

// createAssistantMessage creates an assistant message containing tool use blocks
func (a *LoopAgent) createAssistantMessage(toolCalls []agent.ToolCall) message.Message {
	blocks := make([]message.ContentBlock, 0, len(toolCalls))

	for _, tc := range toolCalls {
		blocks = append(blocks, message.ContentBlock{
			Type:  "tool_use",
			ID:    tc.ID,
			Name:  tc.Name,
			Input: tc.InputMap,
		})
	}

	return message.Message{
		Role:    "assistant",
		Content: blocks,
	}
}

// convertToAgentMessage converts a message.Message to agent.Message
func convertToAgentMessage(msg message.Message) agent.Message {
	return agent.Message{
		Role:    msg.Role,
		Content: msg.Content,
	}
}

// convertToAgentMessages converts a slice of message.Message to agent.Message
func convertToAgentMessages(messages []message.Message) []agent.Message {
	result := make([]agent.Message, len(messages))
	for i, msg := range messages {
		result[i] = convertToAgentMessage(msg)
	}
	return result
}

// StreamingEvent represents an event from the streaming tool loop
type StreamingEvent struct {
	Type        EventType
	Text        string
	Iteration   int
	Error       error
	ToolName    string
	ToolID      string
	ToolCall    string
	ToolContent string
	IsError     bool
	StopReason  string
}

// EventType represents the type of streaming event
type EventType string

const (
	EventTypeToken           EventType = "token"
	EventTypeToolUse         EventType = "tool_use"
	EventTypeToolExecuting   EventType = "tool_executing"
	EventTypeToolResult      EventType = "tool_result"
	EventTypeIteration       EventType = "iteration"
	EventTypeDone            EventType = "done"
	EventTypeError           EventType = "error"
	EventTypeContentBlockDone EventType = "content_block_done"
)

// IsDone returns true if the event indicates the stream is done
func (e StreamingEvent) IsDone() bool {
	return e.Type == EventTypeDone || e.Type == EventTypeError
}

// HasError returns true if the event contains an error
func (e StreamingEvent) HasError() bool {
	return e.Type == EventTypeError && e.Error != nil
}
