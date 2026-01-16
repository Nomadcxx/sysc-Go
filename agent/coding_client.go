package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// CodingClient implements GLMClient for the GLM Coding Plan endpoint
type CodingClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewCodingClient creates a client for the GLM Coding Plan endpoint
func NewCodingClient(apiKey, baseURL string) *CodingClient {
	if baseURL == "" {
		baseURL = "https://api.z.ai/api/coding/paas/v4"
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

	return &CodingClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   "glm-4.7",
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// SetModel sets the model to use
func (c *CodingClient) SetModel(model string) {
	c.model = model
}

// Cancel cancels an active stream
func (c *CodingClient) Cancel(streamID string) {
	// Context cancellation is handled by the caller
}

// CodingChatRequest is the request format for the coding endpoint
type CodingChatRequest struct {
	Model       string     `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int        `json:"max_tokens"`
	Temperature float64    `json:"temperature,omitempty"`
	Stream      bool       `json:"stream"`
}

// CodingChatResponse is the response format
type CodingChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index         int `json:"index"`
		Message       struct {
			Role            string `json:"role"`
			Content         string `json:"content"`
			ReasoningContent string `json:"reasoning_content,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// CodingStreamChunk represents a streaming chunk from the coding endpoint
type CodingStreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role            string `json:"role,omitempty"`
			Content         string `json:"content,omitempty"`
			ReasoningContent string `json:"reasoning_content,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

// StreamChat sends a chat request and returns a channel of stream chunks
func (c *CodingClient) StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	// Convert to coding format
	codingReq := CodingChatRequest{
		Model:       c.model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
	}

	// Convert messages
	for _, msg := range req.Messages {
		codingReq.Messages = append(codingReq.Messages, Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	body, err := json.Marshal(codingReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	ch := make(chan StreamChunk, 100)
	go c.readCodingStream(ctx, resp.Body, ch)

	return ch, nil
}

// readCodingStream reads the SSE stream from the coding endpoint
func (c *CodingClient) readCodingStream(ctx context.Context, body io.ReadCloser, ch chan<- StreamChunk) {
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

		// Parse the chunk
		var chunk CodingStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		// Extract content from choices[0].delta
		if len(chunk.Choices) > 0 {
			choice := chunk.Choices[0]
			delta := choice.Delta

			// Prefer reasoning_content if available, otherwise content
			token := delta.ReasoningContent
			if token == "" {
				token = delta.Content
			}

			if token != "" {
				ch <- StreamChunk{Token: token}
			}

			if choice.FinishReason != "" {
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
