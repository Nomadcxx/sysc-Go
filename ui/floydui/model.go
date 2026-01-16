package floydui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Nomadcxx/sysc-Go/agent"
	"github.com/Nomadcxx/sysc-Go/agent/floyd"
	"github.com/Nomadcxx/sysc-Go/agent/loop"
	agenttools "github.com/Nomadcxx/sysc-Go/agent/tools"
	"github.com/Nomadcxx/sysc-Go/cache"
	"github.com/Nomadcxx/sysc-Go/ui/components/footer"
	"github.com/Nomadcxx/sysc-Go/ui/components/header"
	"github.com/Nomadcxx/sysc-Go/ui/components/palette"
)

// Session data for persistence
type Session struct {
	Messages []string `json:"messages"`
	History  []string `json:"history"`
}

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Role        string
	Content     string
	ToolUse     *ToolUseInfo
	ToolResult  *ToolResultInfo
	IsStreaming bool
	// RenderedContent caches the expensive glamour output
	RenderedContent string
}

// ToolUseInfo contains information about a tool use
type ToolUseInfo struct {
	Name string
	ID   string
}

// ToolResultInfo contains information about a tool result
type ToolResultInfo struct {
	Name    string
	Content string
	IsError bool
}

type AppMode int

const (
	ModeChat AppMode = iota
	ModeCommand
	ModeModelSelect
	ModeExitSummary
)

// Model is the shared FLOYD UI model
type Model struct {
	// UI Components
	Viewport viewport.Model
	Input    textarea.Model

	Progress progress.Model

	Header  header.Model
	Footer  footer.Model
	Palette palette.Model

	// State
	Mode         AppMode
	Messages     []ChatMessage
	History      []string
	HistoryIndex int
	Ready        bool
	Quitting     bool
	Width        int
	Height       int

	// Theme
	CurrentTheme Theme

	// Thinking state
	IsThinking         bool
	CurrentThinking    int
	AnimationFrame     int
	LastTick           time.Time
	LastPacketReceived time.Time

	// Error state
	HasError  bool
	LastError string

	// Help
	ShowHelp bool

	// Progress
	ProgressActive  bool
	ProgressPercent float64

	// Session
	SessionPath string

	// FLOYD Brain (shared across all interfaces)
	ProxyClient  *agent.ProxyClient
	LoopAgent    *loop.LoopAgent
	ToolExecutor *agenttools.Executor
	ProtocolMgr  *floyd.ProtocolManager
	Safety       *floyd.SafetyEnforcer
	CacheMgr     *cache.CacheManager

	// Conversation state
	Conversation []agent.Message

	// Streaming state
	// ⚠️  CRITICAL: CurrentResponse MUST be a pointer because strings.Builder
	// cannot be copied by value once used, and Bubble Tea copies Model on Update.
	CurrentResponse *strings.Builder
	PendingTokens   []string
	// TokenStreamChan is a PERSISTENT channel - created once in NewModel.
	// ⚠️  CRITICAL: NEVER close this channel! Closing causes panics on reuse.
	// Use sentinel value "\x00DONE" to signal stream completion instead.
	// See Claude.md "TUI Channel Pattern" for details.
	TokenStreamChan chan string
	StreamCancel    context.CancelFunc
	StreamWaitGroup *sync.WaitGroup
	StreamDone      bool

	// Tool execution state
	ExecutingTools  bool
	CurrentToolName string
	ToolIteration   int

	// Command and Skills registry
	CommandRegistry *CommandRegistry
}

// NewModel creates a new FLOYD UI model with default session path
func NewModel() Model {
	sessionPath, _ := os.UserHomeDir()
	sessionPath = filepath.Join(sessionPath, ".floyd_session.json")
	return NewModelWithSession(sessionPath)
}

// NewModelWithSession creates a new FLOYD UI model with a specific session path
func NewModelWithSession(sessionPath string) Model {
	// Initialize styles
	InitStyles()

	// Load session
	session, _ := loadSession(sessionPath)

	// Convert loaded messages to ChatMessage format
	messages := make([]ChatMessage, 0, len(session.Messages))
	for _, msg := range session.Messages {
		if strings.HasPrefix(msg, "USER:") {
			messages = append(messages, ChatMessage{
				Role:    "user",
				Content: strings.TrimPrefix(msg, "USER: "),
			})
		} else if strings.HasPrefix(msg, "FLOYD:") {
			messages = append(messages, ChatMessage{
				Role:    "assistant",
				Content: strings.TrimPrefix(msg, "FLOYD: "),
			})
		}
	}

	ti := textarea.New()
	ti.Placeholder = "Ask FLOYD... (Enter to send, Shift+Enter for newline)"
	ti.Focus()
	ti.CharLimit = 1000
	ti.SetWidth(60)
	ti.SetHeight(1) // Start compact
	ti.Prompt = "◈ "
	ti.ShowLineNumbers = false

	vp := viewport.New(0, 0)

	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	// Load config and initialize FLOYD brain
	cfg := loadConfig()

	// Validate API key
	if cfg.apiKey == "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: "ERROR: No API key configured. Set ANTHROPIC_AUTH_TOKEN or GLM_API_KEY environment variable.",
		})
		return Model{
			Viewport:        vp,
			Input:           ti,
			Progress:        prog,
			Messages:        messages,
			History:         session.History,
			HistoryIndex:    len(session.History),
			Ready:           true,
			SessionPath:     sessionPath,
			HasError:        true,
			LastError:       "No API key configured",
			CurrentResponse: &strings.Builder{},
			TokenStreamChan: make(chan string, 100),
			StreamWaitGroup: &sync.WaitGroup{},
		}
	}

	proxyClient := agent.NewProxyClient(cfg.apiKey, cfg.baseURL)
	if cfg.model != "" {
		proxyClient.SetModel(cfg.model)
	}

	toolExecutor := agenttools.NewExecutor(5 * time.Minute)
	protocolMgr := floyd.NewProtocolManager(".floyd")
	safety := floyd.NewSafetyEnforcer()
	loopAgent := loop.NewLoopAgent(proxyClient, toolExecutor)

	// Add welcome message if no messages
	if len(messages) == 0 {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: "FLOYD: PRISM side initialized. Full protocol loaded. Tools armed. Safety enabled.",
		})
	}

	m := Model{
		Viewport:        vp,
		Input:           ti,
		Progress:        prog,
		Messages:        messages,
		History:         session.History,
		HistoryIndex:    len(session.History),
		Ready:           false,
		CurrentTheme:    DefaultTheme,
		IsThinking:      false,
		CurrentThinking: 0,
		AnimationFrame:  0,
		LastTick:        time.Now(),
		HasError:        false,
		ShowHelp:        false,
		ProgressActive:  false,
		ProgressPercent: 0.0,
		SessionPath:     sessionPath,
		ProxyClient:     proxyClient,
		LoopAgent:       loopAgent,
		ToolExecutor:    toolExecutor,
		ProtocolMgr:     protocolMgr,
		Safety:          safety,
		CacheMgr:        cache.NewCacheManager(".floyd/.cache"),
		CurrentResponse: &strings.Builder{},
		PendingTokens:   make([]string, 0, 50),
		TokenStreamChan: make(chan string, 100),
		StreamWaitGroup: &sync.WaitGroup{},
		CommandRegistry: NewCommandRegistry(),

		// Components
		Header:  header.New(),
		Footer:  footer.New(),
		Palette: palette.New(),
	}

	// Initialize conversation from loaded messages
	for _, mMsg := range messages {
		m.Conversation = append(m.Conversation, agent.Message{
			Role:    mMsg.Role,
			Content: mMsg.Content,
		})
	}

	// Load spinners from JSON
	LoadSpinners()

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		Tick(time.Millisecond*250),
		m.waitForToken(),
	)
}

// config holds configuration
type config struct {
	apiKey  string
	baseURL string
	model   string
}

// loadConfig loads configuration from environment
func loadConfig() config {
	apiKey := os.Getenv("ANTHROPIC_AUTH_TOKEN")
	if apiKey == "" {
		apiKey = os.Getenv("GLM_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("ZHIPU_API_KEY")
	}
	if apiKey == "" {
		homeDir, _ := os.UserHomeDir()
		data, _ := os.ReadFile(homeDir + "/.claude/settings.json")
		var settings struct {
			Env struct {
				AnthropicAuthToken string `json:"ANTHROPIC_AUTH_TOKEN"`
			} `json:"env"`
		}
		if json.Unmarshal(data, &settings) == nil {
			apiKey = settings.Env.AnthropicAuthToken
		}
	}

	return config{
		apiKey:  apiKey,
		baseURL: "https://api.z.ai/api/anthropic",
		model:   "claude-opus-4",
	}
}

var sessionMutex sync.Mutex

// loadSession loads session from disk with corrupt file recovery
func loadSession(path string) (Session, error) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	var s Session
	file, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(file, &s)
	if err != nil {
		// Move corrupt file for recovery
		corruptPath := path + ".corrupt"
		_ = os.Rename(path, corruptPath)
		return s, fmt.Errorf("session file corrupt, moved to %s: %w", corruptPath, err)
	}
	return s, nil
}

// saveSession saves session to disk atomically
func saveSession(s Session, path string) error {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: temp file + rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// SaveCurrentSession saves the current model state
func (m Model) SaveCurrentSession() error {
	// Convert ChatMessages to strings for storage
	messages := make([]string, 0, len(m.Messages))
	for _, msg := range m.Messages {
		switch msg.Role {
		case "user":
			messages = append(messages, "USER: "+msg.Content)
		case "assistant":
			messages = append(messages, "FLOYD: "+msg.Content)
		case "system":
			if msg.ToolUse != nil {
				messages = append(messages, "SYSTEM: [Tool: "+msg.ToolUse.Name+"]")
			} else if msg.ToolResult != nil {
				messages = append(messages, "SYSTEM: [Tool Result: "+msg.ToolResult.Name+"]")
			} else {
				messages = append(messages, "SYSTEM: "+msg.Content)
			}
		}
	}

	s := Session{
		Messages: messages,
		History:  m.History,
	}
	return saveSession(s, m.SessionPath)
}

// SetTheme changes the current theme
func (m *Model) SetTheme(name string) tea.Cmd {
	if newTheme, ok := Themes[name]; ok {
		m.CurrentTheme = newTheme
		SetTheme(newTheme)
		m.Input.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(m.CurrentTheme.White)
		return func() tea.Msg {
			return themeChangedMsg{themeName: name}
		}
	}
	return nil
}

// AddUserMessage adds a user message to the conversation
func (m *Model) AddUserMessage(content string) {
	m.Messages = append(m.Messages, ChatMessage{
		Role:    "user",
		Content: content,
	})
	m.Conversation = append(m.Conversation, agent.Message{
		Role:    "user",
		Content: content,
	})
	m.History = append(m.History, content)
	m.HistoryIndex = len(m.History)
}

// AddAssistantMessage adds an assistant message to the conversation
func (m *Model) AddAssistantMessage(content string) {
	m.Messages = append(m.Messages, ChatMessage{
		Role:    "assistant",
		Content: content,
	})
	m.Conversation = append(m.Conversation, agent.Message{
		Role:    "assistant",
		Content: content,
	})
}

// AddSystemMessage adds a system message
func (m *Model) AddSystemMessage(content string) {
	m.Messages = append(m.Messages, ChatMessage{
		Role:    "system",
		Content: content,
	})
}

// AddToolUse adds a tool use notification
func (m *Model) AddToolUse(toolName, toolID string) {
	m.Messages = append(m.Messages, ChatMessage{
		Role:    "system",
		ToolUse: &ToolUseInfo{Name: toolName, ID: toolID},
	})
}

// AddToolResult adds a tool result message
func (m *Model) AddToolResult(toolName, content string, isError bool) {
	m.Messages = append(m.Messages, ChatMessage{
		Role:       "system",
		ToolResult: &ToolResultInfo{Name: toolName, Content: content, IsError: isError},
	})
}

// ClearMessages clears all messages
func (m *Model) ClearMessages() {
	m.Messages = make([]ChatMessage, 0)
	m.Conversation = make([]agent.Message, 0)
}

// GetLastUserMessage returns the last user message
func (m *Model) GetLastUserMessage() string {
	for i := len(m.Messages) - 1; i >= 0; i-- {
		if m.Messages[i].Role == "user" {
			return m.Messages[i].Content
		}
	}
	return ""
}
