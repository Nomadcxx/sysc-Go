package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Nomadcxx/sysc-Go/agent/message"
)

// GLMClient is the interface for GLM API client (via Anthropic-compatible proxy)
type GLMClient interface {
	StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
	Chat(ctx context.Context, req ChatRequest) (string, error)
	Cancel(streamID string)
}

// ChatRequest represents a chat request using Anthropic API format
// This works with api.z.ai proxy which translates to GLM
type ChatRequest struct {
	Messages      []Message              `json:"messages"`
	Model         string                 `json:"model"`
	MaxTokens     int                    `json:"max_tokens,omitempty"`
	Temperature   float64                `json:"temperature,omitempty"`
	TopP          float64                `json:"top_p,omitempty"`
	TopK          int                    `json:"top_k,omitempty"`
	Stream        bool                   `json:"stream,omitempty"`
	System        string                 `json:"system,omitempty"` // Simple string system prompt (backward compat)
	SystemBlocks  []message.ContentBlock `json:"-"`                // Content blocks for cache control (preferred)
	Tools         []Tool                 `json:"tools,omitempty"`
	ToolChoice    any                    `json:"tool_choice,omitempty"`    // "auto", "any", "none", or specific tool
	StopSequences []string               `json:"stop_sequences,omitempty"` // Sequences that trigger stop

	// Beta features for experimental API functionality
	// Sent as anthropic-beta header, not in request body
	Beta []string `json:"-"` // e.g., "computer-use-20250124", "prompt-caching-2024-01-01"
}

// MarshalJSON implements custom JSON marshaling for ChatRequest
// This handles the SystemBlocks field properly
func (r ChatRequest) MarshalJSON() ([]byte, error) {
	type Alias ChatRequest

	// If SystemBlocks are provided, serialize them as system
	if len(r.SystemBlocks) > 0 {
		// Create a map that we'll manually serialize
		obj := map[string]any{
			"messages":   r.Messages,
			"model":      r.Model,
			"max_tokens": r.MaxTokens,
			"system":     r.SystemBlocks,
		}
		if r.Temperature > 0 {
			obj["temperature"] = r.Temperature
		}
		if r.TopP > 0 {
			obj["top_p"] = r.TopP
		}
		if r.TopK > 0 {
			obj["top_k"] = r.TopK
		}
		if r.Tools != nil {
			obj["tools"] = r.Tools
		}
		if r.ToolChoice != nil {
			obj["tool_choice"] = r.ToolChoice
		}
		if r.StopSequences != nil {
			obj["stop_sequences"] = r.StopSequences
		}
		return json.Marshal(obj)
	}

	// Otherwise use standard marshaling
	return json.Marshal((Alias)(r))
}

// Message represents a chat message (Anthropic format)
// Content can be either a string (for backward compatibility) or []ContentBlock (for tool use)
type Message struct {
	Role    string      `json:"role"`    // "user", "assistant", "system"
	Content interface{} `json:"content"` // string or []message.ContentBlock
}

// Tool represents a tool definition for function calling
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// ToolCall represents a tool use request from the LLM
type ToolCall struct {
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name"`
	Input     string         `json:"input,omitempty"`
	Arguments string         `json:"arguments,omitempty"`
	InputMap  map[string]any `json:"-"` // Parsed input as map
}

// StreamChunk represents a single chunk from the streaming response
type StreamChunk struct {
	Token           string
	Done            bool
	ToolCall        *ToolCall
	ToolUseComplete bool   // True when tool_use block is complete
	StopReason      string // Reason for stopping: "end_turn", "tool_use", "max_tokens", etc.
	Error           error
	Usage           *UsageInfo // Token usage from message_delta
}

// UsageInfo represents token usage from API responses
type UsageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ProxyClient implements GLMClient using the api.z.ai Anthropic-compatible proxy
type ProxyClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
	version    string // API version
}

// NewProxyClient creates a new client for the api.z.ai proxy
// It reads credentials from ~/.claude/settings.json if not provided
func NewProxyClient(apiKey, baseURL string) *ProxyClient {
	// Default to Anthropic-compatible proxy for GLM Coding Plan
	if baseURL == "" {
		baseURL = "https://api.z.ai/api/anthropic"
	}

	// Read from Claude settings if no key provided
	if apiKey == "" {
		homeDir, _ := os.UserHomeDir()
		settingsPath := homeDir + "/.claude/settings.json"
		data, err := os.ReadFile(settingsPath)
		if err == nil {
			var settings struct {
				Env struct {
					AnthropicAuthToken string `json:"ANTHROPIC_AUTH_TOKEN"`
				} `json:"env"`
			}
			if json.Unmarshal(data, &settings) == nil {
				apiKey = settings.Env.AnthropicAuthToken
			}
		}
	}

	// Fallback to environment variable
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_AUTH_TOKEN")
	}
	if apiKey == "" {
		apiKey = os.Getenv("GLM_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("ZHIPU_API_KEY")
	}

	return &ProxyClient{
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      "claude-opus-4", // Will map to GLM-4.6 via proxy
		httpClient: &http.Client{
			// Timeout: 0, // No timeout for streaming (relies on context)
		},
		version: "2023-06-01",
	}
}

// SetModel sets the model to use
func (c *ProxyClient) SetModel(model string) {
	c.model = model
}

// SetVersion sets the API version
func (c *ProxyClient) SetVersion(version string) {
	c.version = version
}

// SetBeta sets beta feature flags on a ChatRequest
// Supports multiple beta flags like "computer-use-20250124", "prompt-caching-2024-01-01"
func (r *ChatRequest) SetBeta(features ...string) {
	r.Beta = append(r.Beta, features...)
}

// HasBeta checks if a specific beta feature is enabled
func (r *ChatRequest) HasBeta(feature string) bool {
	for _, f := range r.Beta {
		if f == feature {
			return true
		}
	}
	return false
}

// buildBetaHeader constructs the anthropic-beta header value
func (r *ChatRequest) buildBetaHeader() string {
	if len(r.Beta) == 0 {
		return ""
	}
	return strings.Join(r.Beta, ", ")
}

// StreamChat sends a chat request and returns a channel of stream chunks
func (c *ProxyClient) StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	// Set defaults
	if req.Model == "" {
		req.Model = c.model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 4096
	}
	req.Stream = true

	// Build request body
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers (Anthropic API format)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", c.version)

	// Add beta features header if specified
	if betaHeader := req.buildBetaHeader(); betaHeader != "" {
		httpReq.Header.Set("anthropic-beta", betaHeader)
	}

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Create output channel
	ch := make(chan StreamChunk, 100)

	// Start goroutine to read stream
	go c.readAnthropicStream(ctx, resp.Body, ch)

	return ch, nil
}

// readAnthropicStream reads Anthropic's SSE stream format
// Enhanced to handle tool_use blocks and content_block_delta for tool input
func (c *ProxyClient) readAnthropicStream(ctx context.Context, body io.ReadCloser, ch chan<- StreamChunk) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)

	// Track active tool calls being assembled
	type activeTool struct {
		ID         string
		Name       string
		InputJSON  strings.Builder
		IsComplete bool
	}
	activeTools := make(map[int]*activeTool) // keyed by content block index

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			ch <- StreamChunk{Done: true, Error: ctx.Err()}
			return
		default:
		}

		line := scanner.Text()

		// SSE format: lines starting with "data: "
		if len(line) < 6 || line[:6] != "data: " {
			continue
		}

		data := line[6:]

		// Check for end of stream
		if data == "[DONE]" {
			ch <- StreamChunk{Done: true}
			return
		}

		// Parse Anthropic streaming event - enhanced for tool use
		var event struct {
			Type  string `json:"type"`
			Index int    `json:"index,omitempty"`

			// content_block_start fields
			ContentBlock struct {
				Type string `json:"type"`
				ID   string `json:"id,omitempty"`
				Name string `json:"name,omitempty"`
			} `json:"content_block,omitempty"`

			// content_block_delta fields
			Delta struct {
				Type        string `json:"type,omitempty"`
				Text        string `json:"text,omitempty"`
				PartialJSON string `json:"partial_json,omitempty"` // For tool input streaming
				StopReason  string `json:"stop_reason,omitempty"`
			} `json:"delta,omitempty"`

			// message_start fields
			Message struct {
				ID      string `json:"id"`
				Type    string `json:"type"`
				Role    string `json:"role"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text,omitempty"`
				} `json:"content"`
				StopReason string `json:"stop_reason,omitempty"`
			} `json:"message,omitempty"`

			// message_delta fields
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage,omitempty"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			// Skip invalid chunks
			continue
		}

		switch event.Type {
		case "message_start":
			// Beginning of stream - initialize
			continue

		case "content_block_start":
			// A new content block is starting
			// Check if it's a tool_use block
			if event.ContentBlock.Type == "tool_use" {
				activeTools[event.Index] = &activeTool{
					ID:         event.ContentBlock.ID,
					Name:       event.ContentBlock.Name,
					IsComplete: false,
				}
			}

		case "content_block_delta":
			// Content is being streamed into the current block
			switch event.Delta.Type {
			case "text_delta":
				// Regular text content
				if event.Delta.Text != "" {
					ch <- StreamChunk{Token: event.Delta.Text}
				}
			case "input_json_delta":
				// Tool input JSON being streamed
				if tool, ok := activeTools[event.Index]; ok {
					tool.InputJSON.WriteString(event.Delta.PartialJSON)
				}
			}

		case "content_block_stop":
			// Content block is complete - finalize tool call if applicable
			if tool, ok := activeTools[event.Index]; ok && tool.Name != "" {
				// Parse the accumulated JSON
				var inputMap map[string]any
				if tool.InputJSON.Len() > 0 {
					json.Unmarshal([]byte(tool.InputJSON.String()), &inputMap)
				}

				ch <- StreamChunk{
					ToolCall: &ToolCall{
						ID:       tool.ID,
						Name:     tool.Name,
						Input:    tool.InputJSON.String(),
						InputMap: inputMap,
					},
					ToolUseComplete: true,
				}

				// Mark as complete and remove from active tracking
				tool.IsComplete = true
				delete(activeTools, event.Index)
			}

		case "message_delta":
			// Message is ending - check stop reason and usage
			if event.Delta.StopReason != "" {
				usage := &UsageInfo{}
				if event.Usage.InputTokens > 0 || event.Usage.OutputTokens > 0 {
					usage.InputTokens = event.Usage.InputTokens
					usage.OutputTokens = event.Usage.OutputTokens
				}

				ch <- StreamChunk{
					Done:       true,
					StopReason: event.Delta.StopReason,
					Usage:      usage,
				}
				return
			}

		case "ping":
			// Keepalive - ignore
			continue
		}

		// Legacy format handling (some proxies use simpler format)
		if event.Type == "" && len(event.Message.Content) > 0 {
			for _, content := range event.Message.Content {
				if content.Type == "text" && content.Text != "" {
					ch <- StreamChunk{Token: content.Text}
				}
			}
			if event.Message.StopReason != "" {
				ch <- StreamChunk{Done: true, StopReason: event.Message.StopReason}
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		ch <- StreamChunk{Done: true, Error: err}
	}

	ch <- StreamChunk{Done: true}
}

// Cancel cancels an active stream (by context cancellation)
func (c *ProxyClient) Cancel(streamID string) {
	// Context cancellation is handled by the caller
}

// Chat performs a non-streaming chat request
func (c *ProxyClient) Chat(ctx context.Context, req ChatRequest) (string, error) {
	req.Stream = false
	if req.Model == "" {
		req.Model = c.model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 4096
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", c.version)

	// Add beta features header if specified
	if betaHeader := req.buildBetaHeader(); betaHeader != "" {
		httpReq.Header.Set("anthropic-beta", betaHeader)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	// Extract text from content blocks
	var output strings.Builder
	for _, content := range result.Content {
		if content.Type == "text" {
			output.WriteString(content.Text)
		}
	}

	return output.String(), nil
}

// Helper to encode API key for Basic auth if needed
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
