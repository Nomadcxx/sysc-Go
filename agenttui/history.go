package agenttui

import (
	"strings"
	"sync"
	"time"

	"github.com/Nomadcxx/sysc-Go/agent"
)

// Role represents the message sender role
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

// Message represents a single message in the conversation history
type Message struct {
	ID        string
	Role      Role
	Content   string
	Timestamp time.Time
	TokenCount int

	// Tool call fields (if Role == RoleTool)
	ToolName   string
	ToolInput  string
	ToolOutput string

	// Metadata
	Metadata map[string]any
}

// Formatted returns the message as a formatted string for display
func (m Message) Formatted() string {
	var sb strings.Builder

	timestamp := m.Timestamp.Format("15:04:05")

	switch m.Role {
	case RoleUser:
		sb.WriteString("│ [")
		sb.WriteString(timestamp)
		sb.WriteString("] You:\n")
		sb.WriteString("│ ")
		sb.WriteString(m.Content)
		sb.WriteString("\n")

	case RoleAssistant:
		sb.WriteString("│ [")
		sb.WriteString(timestamp)
		sb.WriteString("] FLOYD:\n")
		sb.WriteString("│ ")
		sb.WriteString(m.Content)
		sb.WriteString("\n")

	case RoleSystem:
		sb.WriteString("│ [")
		sb.WriteString(timestamp)
		sb.WriteString("] System: ")
		sb.WriteString(m.Content)
		sb.WriteString("\n")

	case RoleTool:
		sb.WriteString("│ [")
		sb.WriteString(timestamp)
		sb.WriteString("] Tool: ")
		sb.WriteString(m.ToolName)
		sb.WriteString("\n")
		if m.ToolInput != "" {
			sb.WriteString("│ Input: ")
			sb.WriteString(m.ToolInput)
			sb.WriteString("\n")
		}
		if m.ToolOutput != "" {
			sb.WriteString("│ Output:\n")
			for _, line := range strings.Split(m.ToolOutput, "\n") {
				sb.WriteString("│ │ ")
				sb.WriteString(line)
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

// HistoryStore manages conversation history with efficient iteration
type HistoryStore struct {
	mu       sync.RWMutex
	messages []Message
	maxSize  int
	cursor   int // Current position for navigation
}

// NewHistoryStore creates a new history store with the given maximum size
func NewHistoryStore(maxSize int) *HistoryStore {
	if maxSize <= 0 {
		maxSize = 10000 // Default max messages
	}

	return &HistoryStore{
		messages: make([]Message, 0, maxSize),
		maxSize:  maxSize,
		cursor:   -1,
	}
}

// Append adds a new message to history
func (h *HistoryStore) Append(msg Message) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if msg.ID == "" {
		msg.ID = generateID()
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	h.messages = append(h.messages, msg)

	// Evict oldest if at capacity
	if len(h.messages) > h.maxSize {
		h.messages = h.messages[1:]
	}

	// Move cursor to the newest message
	h.cursor = len(h.messages) - 1
}

// Iter returns a Go 1.23+ iterator for memory-efficient traversal
// Iterates from newest to oldest (reverse chronological order)
func (h *HistoryStore) Iter() func(yield func(Message) bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return func(yield func(Message) bool) {
		for i := len(h.messages) - 1; i >= 0; i-- {
			if !yield(h.messages[i]) {
				return
			}
		}
	}
}

// IterRange returns an iterator for a range of messages
// start is the index to start from (inclusive)
// count is the maximum number of messages to yield
func (h *HistoryStore) IterRange(start, count int) func(yield func(Message) bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Ensure start is within bounds
	if start < 0 {
		start = 0
	}
	if start >= len(h.messages) {
		start = len(h.messages) - 1
	}

	// Determine end index
	end := start - count
	if end < 0 {
		end = -1
	}

	return func(yield func(Message) bool) {
		for i := start; i > end; i-- {
			if !yield(h.messages[i]) {
				return
			}
		}
	}
}

// GetLatest returns the most recent N messages
func (h *HistoryStore) GetLatest(n int) []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n <= 0 || len(h.messages) == 0 {
		return []Message{}
	}

	start := len(h.messages) - n
	if start < 0 {
		start = 0
	}

	result := make([]Message, len(h.messages)-start)
	copy(result, h.messages[start:])
	return result
}

// GetMessagesForAPI returns messages formatted for LLM API calls
// Excludes system messages if includeSystem is false
func (h *HistoryStore) GetMessagesForAPI(includeSystem bool) []agent.Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]agent.Message, 0, len(h.messages))

	for _, msg := range h.messages {
		if !includeSystem && msg.Role == RoleSystem {
			continue
		}
		if msg.Role == RoleTool {
			// Skip tool result messages in API call
			// Tool calls are embedded in assistant messages
			continue
		}
		result = append(result, agent.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	return result
}

// Len returns the current number of messages
func (h *HistoryStore) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.messages)
}

// GetTokenCount returns the total token count (approximate)
func (h *HistoryStore) GetTokenCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total := 0
	for _, msg := range h.messages {
		total += msg.TokenCount
		if msg.TokenCount == 0 {
			// Rough estimate: ~4 chars per token
			total += len(msg.Content) / 4
		}
	}
	return total
}

// TruncateToTokens removes oldest messages until token count is below limit
func (h *HistoryStore) TruncateToTokens(maxTokens int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for h.GetTokenCount() > maxTokens && len(h.messages) > 1 {
		// Remove oldest message, but keep system messages
		found := false
		for i, msg := range h.messages {
			if msg.Role != RoleSystem {
				// Remove this message
				h.messages = append(h.messages[:i], h.messages[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			// Only system messages left, stop truncating
			break
		}
	}
}

// Clear removes all messages except system messages
func (h *HistoryStore) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	system := make([]Message, 0)
	for _, msg := range h.messages {
		if msg.Role == RoleSystem {
			system = append(system, msg)
		}
	}

	h.messages = system
	h.cursor = len(h.messages) - 1
}

// FindByID finds a message by its ID
func (h *HistoryStore) FindByID(id string) (Message, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, msg := range h.messages {
		if msg.ID == id {
			return msg, true
		}
	}
	return Message{}, false
}

// String returns the full conversation as a string
func (h *HistoryStore) String() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var sb strings.Builder
	for _, msg := range h.messages {
		sb.WriteString(msg.Formatted())
		sb.WriteString("\n")
	}
	return sb.String()
}

// generateID creates a unique ID for a message
func generateID() string {
	return time.Now().Format("20060102-150405") + "-" + randomString(4)
}

// randomString generates a random alphanumeric string
func randomString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		// Simple pseudo-random based on time
		b[i] = chars[int(time.Now().UnixNano()+int64(i))%len(chars)]
	}
	return string(b)
}
