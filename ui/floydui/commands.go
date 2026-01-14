package floydui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// SlashCommand represents a registered slash command
type SlashCommand struct {
	Name        string                                        `json:"name"`
	Aliases     []string                                      `json:"aliases,omitempty"`
	Description string                                        `json:"description"`
	Usage       string                                        `json:"usage,omitempty"`
	Handler     func(m Model, args []string) (Model, tea.Cmd) `json:"-"`
	// For custom commands loaded from files
	Prompt    string `json:"prompt,omitempty"`
	IsBuiltIn bool   `json:"is_built_in"`
}

// Skill represents a FLOYD skill (similar to Claude's skills)
type Skill struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Triggers    []string `json:"triggers,omitempty"` // Keywords that activate the skill
	Prompt      string   `json:"prompt"`             // Injected into system prompt when active
	Commands    []string `json:"commands,omitempty"` // Slash commands this skill adds
	Enabled     bool     `json:"enabled"`
}

// CommandRegistry manages all slash commands
type CommandRegistry struct {
	Commands map[string]*SlashCommand
	Skills   map[string]*Skill
}

// NewCommandRegistry creates a new registry with built-in commands
func NewCommandRegistry() *CommandRegistry {
	r := &CommandRegistry{
		Commands: make(map[string]*SlashCommand),
		Skills:   make(map[string]*Skill),
	}
	r.registerBuiltInCommands()
	r.loadGlobalCommands()
	r.loadSkills()
	return r
}

// registerBuiltInCommands adds all built-in slash commands
func (r *CommandRegistry) registerBuiltInCommands() {
	builtins := []*SlashCommand{
		{
			Name:        "help",
			Aliases:     []string{"h", "?"},
			Description: "Show help and available commands",
			IsBuiltIn:   true,
		},
		{
			Name:        "exit",
			Aliases:     []string{"quit", "q"},
			Description: "Exit FLOYD",
			IsBuiltIn:   true,
		},
		{
			Name:        "clear",
			Aliases:     []string{"cls"},
			Description: "Clear chat history",
			IsBuiltIn:   true,
		},
		{
			Name:        "status",
			Description: "Show workspace status",
			IsBuiltIn:   true,
		},
		{
			Name:        "tools",
			Description: "List available tools",
			IsBuiltIn:   true,
		},
		{
			Name:        "theme",
			Aliases:     []string{"t"},
			Description: "Change theme",
			Usage:       "/theme <name>",
			IsBuiltIn:   true,
		},
		{
			Name:        "protocol",
			Description: "Show FLOYD protocol status",
			IsBuiltIn:   true,
		},
		{
			Name:        "init",
			Description: "Initialize .floyd/ workspace",
			IsBuiltIn:   true,
		},
		{
			Name:        "compact",
			Description: "Summarize and compact the conversation",
			IsBuiltIn:   true,
		},
		{
			Name:        "commands",
			Aliases:     []string{"cmds"},
			Description: "List all available commands",
			IsBuiltIn:   true,
		},
		{
			Name:        "skills",
			Description: "List and manage skills",
			Usage:       "/skills [enable|disable <name>]",
			IsBuiltIn:   true,
		},
		{
			Name:        "memory",
			Aliases:     []string{"cache"},
			Description: "Show SUPERCACHE status",
			IsBuiltIn:   true,
		},
	}

	for _, cmd := range builtins {
		r.Register(cmd)
	}
}

// Register adds a command to the registry
func (r *CommandRegistry) Register(cmd *SlashCommand) {
	r.Commands[cmd.Name] = cmd
	for _, alias := range cmd.Aliases {
		r.Commands[alias] = cmd
	}
}

// Get retrieves a command by name or alias
func (r *CommandRegistry) Get(name string) *SlashCommand {
	return r.Commands[strings.ToLower(name)]
}

// loadGlobalCommands loads custom commands from ~/.floyd/commands/
func (r *CommandRegistry) loadGlobalCommands() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// Check ~/.floyd/commands/
	commandsDir := filepath.Join(home, ".floyd", "commands")
	r.loadCommandsFromDir(commandsDir)

	// Also check project-local .floyd/commands/
	r.loadCommandsFromDir(filepath.Join(".floyd", "commands"))
}

// loadCommandsFromDir loads command definitions from a directory
func (r *CommandRegistry) loadCommandsFromDir(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return // Directory doesn't exist, that's fine
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var cmd SlashCommand
		if err := json.Unmarshal(data, &cmd); err != nil {
			continue
		}

		cmd.IsBuiltIn = false
		r.Register(&cmd)
	}
}

// loadSkills loads skills from ~/.floyd/skills/ and .floyd/skills/
func (r *CommandRegistry) loadSkills() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// Load global skills
	r.loadSkillsFromDir(filepath.Join(home, ".floyd", "skills"))

	// Load project-local skills
	r.loadSkillsFromDir(filepath.Join(".floyd", "skills"))
}

// loadSkillsFromDir loads skill definitions from a directory
func (r *CommandRegistry) loadSkillsFromDir(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") && !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, file.Name())

		if strings.HasSuffix(file.Name(), ".json") {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			var skill Skill
			if err := json.Unmarshal(data, &skill); err != nil {
				continue
			}
			r.Skills[skill.Name] = &skill
		} else if strings.HasSuffix(file.Name(), ".md") {
			// Load markdown skill (prompt is the file content)
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			name := strings.TrimSuffix(file.Name(), ".md")
			r.Skills[name] = &Skill{
				Name:        name,
				Description: fmt.Sprintf("Skill loaded from %s", file.Name()),
				Prompt:      string(data),
				Enabled:     true,
			}
		}
	}
}

// ListCommands returns all unique commands sorted by name
func (r *CommandRegistry) ListCommands() []*SlashCommand {
	seen := make(map[string]bool)
	var result []*SlashCommand

	for _, cmd := range r.Commands {
		if !seen[cmd.Name] {
			seen[cmd.Name] = true
			result = append(result, cmd)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// ListSkills returns all skills
func (r *CommandRegistry) ListSkills() []*Skill {
	var result []*Skill
	for _, skill := range r.Skills {
		result = append(result, skill)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// GetActiveSkillPrompts returns prompts for all enabled skills
func (r *CommandRegistry) GetActiveSkillPrompts() string {
	var parts []string
	for _, skill := range r.Skills {
		if skill.Enabled && skill.Prompt != "" {
			parts = append(parts, fmt.Sprintf("=== SKILL: %s ===\n%s", skill.Name, skill.Prompt))
		}
	}
	return strings.Join(parts, "\n\n")
}

// EnableSkill enables a skill by name
func (r *CommandRegistry) EnableSkill(name string) bool {
	if skill, ok := r.Skills[name]; ok {
		skill.Enabled = true
		return true
	}
	return false
}

// DisableSkill disables a skill by name
func (r *CommandRegistry) DisableSkill(name string) bool {
	if skill, ok := r.Skills[name]; ok {
		skill.Enabled = false
		return true
	}
	return false
}

// CompactConversation generates a summary of the current conversation
func (m Model) CompactConversation() string {
	var sb strings.Builder

	sb.WriteString("# Conversation Compact\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Count messages by role
	userCount := 0
	assistantCount := 0
	systemCount := 0
	for _, msg := range m.Messages {
		switch msg.Role {
		case "user":
			userCount++
		case "assistant", "floyd":
			assistantCount++
		case "system":
			systemCount++
		}
	}

	sb.WriteString("## Session Stats\n")
	sb.WriteString(fmt.Sprintf("- User messages: %d\n", userCount))
	sb.WriteString(fmt.Sprintf("- Assistant responses: %d\n", assistantCount))
	sb.WriteString(fmt.Sprintf("- System messages: %d\n", systemCount))
	sb.WriteString("\n")

	// Extract key topics from user messages
	sb.WriteString("## Topics Discussed\n")
	topics := extractTopics(m.Messages)
	for _, topic := range topics {
		sb.WriteString(fmt.Sprintf("- %s\n", topic))
	}
	sb.WriteString("\n")

	// Recent context (last 5 exchanges)
	sb.WriteString("## Recent Context\n")
	recentMessages := m.Messages
	if len(recentMessages) > 10 {
		recentMessages = recentMessages[len(recentMessages)-10:]
	}
	for _, msg := range recentMessages {
		role := msg.Role
		content := msg.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		sb.WriteString(fmt.Sprintf("**%s**: %s\n\n", role, content))
	}

	return sb.String()
}

// extractTopics extracts key topics from messages (simple implementation)
func extractTopics(messages []ChatMessage) []string {
	topicCounts := make(map[string]int)

	// Simple keyword extraction
	keywords := []string{
		"TUI", "agent", "tool", "bug", "fix", "error", "test", "build",
		"cache", "protocol", "API", "streaming", "channel", "panic",
		"command", "theme", "session", "file", "code", "function",
	}

	for _, msg := range messages {
		if msg.Role != "user" {
			continue
		}
		content := strings.ToLower(msg.Content)
		for _, kw := range keywords {
			if strings.Contains(content, strings.ToLower(kw)) {
				topicCounts[kw]++
			}
		}
	}

	// Sort by count and return top topics
	type topicCount struct {
		topic string
		count int
	}
	var sorted []topicCount
	for topic, count := range topicCounts {
		sorted = append(sorted, topicCount{topic, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	var result []string
	for i, tc := range sorted {
		if i >= 8 { // Top 8 topics
			break
		}
		result = append(result, tc.topic)
	}

	if len(result) == 0 {
		result = append(result, "General conversation")
	}

	return result
}
