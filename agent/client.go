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
	"time"
)

// GLMClient is the interface for GLM API client (via Anthropic-compatible proxy)
type GLMClient interface {
	StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
	Cancel(streamID string)
}

// ChatRequest represents a chat request using Anthropic API format
// This works with api.z.ai proxy which translates to GLM
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	TopK        int       `json:"top_k,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	System      string    `json:"system,omitempty"`
	Tools       []Tool    `json:"tools,omitempty"`
	ToolChoice  any       `json:"tool_choice,omitempty"` // "auto", "any", or specific tool
}

// Message represents a chat message (Anthropic format)
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"`
}

// Tool represents a tool definition for function calling
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]any         `json:"input_schema"`
}

// ToolCall represents a tool use request from the LLM
type ToolCall struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Input    string `json:"input,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// StreamChunk represents a single chunk from the streaming response
type StreamChunk struct {
	Token    string
	Done     bool
	ToolCall *ToolCall
	Error    error
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

	return &ProxyClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   "claude-opus-4", // Will map to GLM-4.6 via proxy
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
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
func (c *ProxyClient) readAnthropicStream(ctx context.Context, body io.ReadCloser, ch chan<- StreamChunk) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)

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

		// Parse Anthropic streaming event
		var event struct {
			Type  string `json:"type"`
			Index int    `json:"index,omitempty"`
			Delta struct {
				Type  string `json:"type,omitempty"`
				Text  string `json:"text,omitempty"`
				StopReason string `json:"stop_reason,omitempty"`
			} `json:"delta,omitempty"`
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
			// Beginning of stream - no content yet
			continue

		case "content_block_start":
			// Content block starting
			continue

		case "content_block_delta":
			// Text content incoming - check for text_delta type
			if event.Delta.Type == "text_delta" && event.Delta.Text != "" {
				ch <- StreamChunk{Token: event.Delta.Text}
			}

		case "content_block_stop":
			// Content block complete
			continue

		case "message_delta":
			// Message ending
			if event.Delta.StopReason != "" {
				ch <- StreamChunk{Done: true}
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
				ch <- StreamChunk{Done: true}
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
		ID           string `json:"id"`
		Type         string `json:"type"`
		Role         string `json:"role"`
		Content      []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason   string `json:"stop_reason"`
		Usage        struct {
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
