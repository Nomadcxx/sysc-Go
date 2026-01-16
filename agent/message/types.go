package message

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// ContentBlock represents a flexible content block that can be text, image, tool use, or tool result
// This supports the full Anthropic Messages API format
type ContentBlock struct {
	Type string `json:"type"` // "text", "image", "tool_use", "tool_result"

	// Text content (for type="text")
	Text string `json:"text,omitempty"`

	// Cache control for prompt caching
	CacheControl *CacheControl `json:"cache_control,omitempty"`

	// Image source (for type="image")
	Source *ImageSource `json:"source,omitempty"`

	// Tool use content (for type="tool_use")
	ID    string         `json:"id,omitempty"`
	Name  string         `json:"name,omitempty"`
	Input map[string]any `json:"input,omitempty"`

	// Tool result content (for type="tool_result")
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`

	// Thinking content (GLM-4.7 extended thinking)
	Thinking *ThinkingBlock `json:"_thinking,omitempty"`
}

// CacheControl enables prompt caching for content blocks
type CacheControl struct {
	Type string `json:"type"` // "ephemeral"
	TTL  string `json:"ttl,omitempty"` // "5m", "1h", etc.
}

// ImageSource represents an image input
type ImageSource struct {
	Type      string `json:"type"` // "base64" or "url"
	MediaType string `json:"media_type,omitempty"` // "image/png", "image/jpeg", etc.
	Data      string `json:"data,omitempty"`   // base64 data for type="base64"
	URL       string `json:"url,omitempty"`    // url for type="url"
}

// ThinkingBlock represents extended thinking from the model
type ThinkingBlock struct {
	Content    string    `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
	TokenCount int       `json:"token_count"`
}

// Message represents a chat message with enhanced content block support
type Message struct {
	Role      string         `json:"role"`    // "user", "assistant", "system"
	Content   []ContentBlock `json:"content"` // Array of content blocks
	Usage     *UsageInfo     `json:"usage,omitempty"`
	Model     string         `json:"model,omitempty"`
	StopReason string         `json:"stop_reason,omitempty"`
	ID        string         `json:"id,omitempty"`
	Type      string         `json:"type,omitempty"`
}

// UsageInfo represents token usage from API responses
type UsageInfo struct {
	InputTokens              int `json:"input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	OutputTokens             int `json:"output_tokens"`
}

// ToolUseBlock represents a complete tool use call extracted from a response
type ToolUseBlock struct {
	ID    string
	Name  string
	Input map[string]any
}

// ToolResultBlock represents a tool execution result to send back
type ToolResultBlock struct {
	ToolUseID string
	Content   string
	IsError   bool
}

// NewTextMessage creates a simple text message (backward compatible)
func NewTextMessage(role, text string) Message {
	return Message{
		Role:    role,
		Content: []ContentBlock{{Type: "text", Text: text}},
	}
}

// NewSystemMessage creates a system message with optional cache control
func NewSystemMessage(text string, cacheTTL string) Message {
	msg := Message{
		Role:    "system",
		Content: []ContentBlock{{Type: "text", Text: text}},
	}
	if cacheTTL != "" {
		msg.Content[0].CacheControl = &CacheControl{Type: "ephemeral", TTL: cacheTTL}
	}
	return msg
}

// NewToolUseMessage creates a message containing a tool use call
func NewToolUseMessage(toolID, name string, input map[string]any) Message {
	return Message{
		Role: "assistant",
		Content: []ContentBlock{{
			Type:  "tool_use",
			ID:    toolID,
			Name:  name,
			Input: input,
		}},
	}
}

// NewToolResultMessage creates a user message containing tool results
func NewToolResultMessage(toolUseID, content string, isError bool) Message {
	return Message{
		Role: "user",
		Content: []ContentBlock{{
			Type:      "tool_result",
			ToolUseID: toolUseID,
			Content:   content,
			IsError:   isError,
		}},
	}
}

// NewImageMessage creates a message with an image attachment
func NewImageMessage(role string, imagePath string) (Message, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return Message{}, fmt.Errorf("read image: %w", err)
	}

	// Detect media type
	ext := strings.ToLower(imagePath[strings.LastIndex(imagePath, "."):])
	mediaType := "image/png"
	if ext == ".jpg" || ext == ".jpeg" {
		mediaType = "image/jpeg"
	} else if ext == ".gif" {
		mediaType = "image/gif"
	} else if ext == ".webp" {
		mediaType = "image/webp"
	}

	base64Data := base64.StdEncoding.EncodeToString(data)

	return Message{
		Role: role,
		Content: []ContentBlock{
			{Type: "text", Text: "[Image attached]"},
			{
				Type: "image",
				Source: &ImageSource{
					Type:      "base64",
					MediaType: mediaType,
					Data:      base64Data,
				},
			},
		},
	}, nil
}

// NewImageMessageFromURL creates a message with an image from URL
func NewImageMessageFromURL(role, imageURL string) Message {
	return Message{
		Role: role,
		Content: []ContentBlock{
			{Type: "text", Text: "[Image attached]"},
			{
				Type: "image",
				Source: &ImageSource{
					Type: "url",
					URL:  imageURL,
				},
			},
		},
	}
}

// ExtractTextContent extracts all text from content blocks
func (m Message) ExtractTextContent() string {
	var sb strings.Builder
	for _, block := range m.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	return sb.String()
}

// HasToolUse checks if message contains tool_use blocks
func (m Message) HasToolUse() bool {
	for _, block := range m.Content {
		if block.Type == "tool_use" {
			return true
		}
	}
	return false
}

// GetToolUseBlocks extracts all tool_use blocks from a message
func (m Message) GetToolUseBlocks() []ToolUseBlock {
	var blocks []ToolUseBlock
	for _, block := range m.Content {
		if block.Type == "tool_use" {
			blocks = append(blocks, ToolUseBlock{
				ID:    block.ID,
				Name:  block.Name,
				Input: block.Input,
			})
		}
	}
	return blocks
}

// HasToolResult checks if message contains tool_result blocks
func (m Message) HasToolResult() bool {
	for _, block := range m.Content {
		if block.Type == "tool_result" {
			return true
		}
	}
	return false
}

// GetToolResults extracts all tool_result blocks from a message
func (m Message) GetToolResults() []ToolResultBlock {
	var blocks []ToolResultBlock
	for _, block := range m.Content {
		if block.Type == "tool_result" {
			blocks = append(blocks, ToolResultBlock{
				ToolUseID: block.ToolUseID,
				Content:   block.Content,
				IsError:   block.IsError,
			})
		}
	}
	return blocks
}

// AddCacheControl adds cache control to all cacheable text blocks
func (m *Message) AddCacheControl(ttl string) {
	for i := range m.Content {
		if m.Content[i].Type == "text" && len(m.Content[i].Text) > 1024 {
			m.Content[i].CacheControl = &CacheControl{
				Type: "ephemeral",
				TTL:  ttl,
			}
		}
	}
}

// MarshalJSON implements custom JSON marshaling for backward compatibility
func (m Message) MarshalJSON() ([]byte, error) {
	// For simple text-only messages, use string format for compatibility
	if len(m.Content) == 1 && m.Content[0].Type == "text" && m.Content[0].CacheControl == nil {
		type Alias Message
		return json.Marshal(&struct {
			Content string `json:"content"`
			*Alias
		}{
			Content: m.Content[0].Text,
			Alias:   (*Alias)(&m),
		})
	}
	return json.Marshal(struct {
		Role    string         `json:"role"`
		Content []ContentBlock `json:"content"`
	}{
		Role:    m.Role,
		Content: m.Content,
	})
}

// ToSimpleMessage converts to old-style string content message (backward compat)
func (m Message) ToSimpleMessage() Message {
	if len(m.Content) == 0 {
		return m
	}
	if m.Content[0].Type == "text" {
		return NewTextMessage(m.Role, m.Content[0].Text)
	}
	return m
}

// MessagesToSimple converts a slice of messages to simple format
func MessagesToSimple(messages []Message) []Message {
	result := make([]Message, len(messages))
	for i, m := range messages {
		result[i] = m.ToSimpleMessage()
	}
	return result
}
