package floyd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Nomadcxx/sysc-Go/agent"
	"github.com/Nomadcxx/sysc-Go/agent/prompt"
	"github.com/Nomadcxx/sysc-Go/agent/tools"
)

const (
	// ModeDefault is the default FLOYD mode
	ModeDefault = ""

	// ModePlanner is for planning tasks
	ModePlanner = "planner"

	// ModeCoder is for writing code
	ModeCoder = "coder"

	// ModeTester is for testing
	ModeTester = "tester"

	// ModeSearch is for codebase exploration
	ModeSearch = "search"
)

// AgentType represents the type of specialist agent to spawn
type AgentType string

const (
	AgentTypePlanner  AgentType = "planner"
	AgentTypeCoder    AgentType = "coder"
	AgentTypeTester   AgentType = "tester"
	AgentTypeSearcher AgentType = "search"
	AgentTypeReviewer AgentType = "reviewer"
)

// ProtocolManager manages the FLOYD protocol integration
type ProtocolManager struct {
	floydDir string
	mu       sync.RWMutex

	// Cached protocol data
	masterPlan  string
	progressLog string
	stack       string
	scratchpad  string
}

// NewProtocolManager creates a new protocol manager
func NewProtocolManager(floydDir string) *ProtocolManager {
	if floydDir == "" {
		floydDir = prompt.FloydDir
	}

	return &ProtocolManager{
		floydDir: floydDir,
	}
}

// InjectProtocol injects the FLOYD protocol into messages
// It prepends the system prompt with protocol and context
func (pm *ProtocolManager) InjectProtocol(messages []agent.Message, mode string) []agent.Message {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Build the full system message with protocol and context
	sysMsg := prompt.BuildSystemMessage(mode)

	// Convert to agent.Message format
	protoMsg := agent.Message{
		Role:    "system",
		Content: sysMsg.Content,
	}

	// Prepend to messages
	result := make([]agent.Message, 0, len(messages)+1)
	result = append(result, protoMsg)
	result = append(result, messages...)

	return result
}

// LoadContextFromFilesystem loads all relevant .floyd/ files
func (pm *ProtocolManager) LoadContextFromFilesystem() (masterPlan, progress, stack string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.masterPlan = prompt.LoadMasterPlan()
	pm.progressLog = prompt.LoadProgressLog(10) // Last 10 entries
	pm.stack = prompt.LoadStack()
	pm.scratchpad = prompt.LoadScratchpad()

	return pm.masterPlan, pm.progressLog, pm.stack
}

// GetContextForPrompt returns formatted context for injection into prompts
func (pm *ProtocolManager) GetContextForPrompt() string {
	masterPlan, progress, stack := pm.LoadContextFromFilesystem()

	var parts []string

	if masterPlan != "" && !strings.Contains(masterPlan, "No master plan found") {
		parts = append(parts, "## CURRENT MASTER PLAN")
		parts = append(parts, truncateLines(masterPlan, 30))
	}

	if stack != "" && !strings.Contains(stack, "Stack not defined") {
		parts = append(parts, "## TECHNOLOGY STACK")
		parts = append(parts, truncateLines(stack, 20))
	}

	if progress != "" && !strings.Contains(progress, "No progress log") {
		parts = append(parts, "## RECENT PROGRESS")
		parts = append(parts, progress)
	}

	return strings.Join(parts, "\n\n")
}

// CheckSafetyRules validates an action against FLOYD safety rules
func (pm *ProtocolManager) CheckSafetyRules(action string) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	actionLower := strings.ToLower(strings.TrimSpace(action))

	// Rule 1: NEVER push to main/master
	if isMainPush(actionLower) {
		return &SafetyViolationError{
			Rule:    "NO_MAIN_PUSH",
			Action:  action,
			Message: "Direct push to main/master branch is blocked by FLOYD protocol. Use a feature branch instead.",
		}
	}

	// Rule 2: Check for direct commits to main
	if isMainCommit(actionLower) {
		return &SafetyViolationError{
			Rule:    "NO_MAIN_COMMIT",
			Action:  action,
			Message: "Direct commit to main/master is blocked. Work on a feature branch.",
		}
	}

	return nil
}

// ShouldSpawnSpecialist determines if a specialist agent should be spawned
func (pm *ProtocolManager) ShouldSpawnSpecialist(task string) (bool, AgentType) {
	taskLower := strings.ToLower(task)

	// Planning tasks
	if containsAny(taskLower, []string{"plan", "design", "architecture", "break down"}) {
		return true, AgentTypePlanner
	}

	// Testing tasks
	if containsAny(taskLower, []string{"test", "verify", "check bug", "find error"}) {
		return true, AgentTypeTester
	}

	// Search/exploration tasks
	if containsAny(taskLower, []string{"find", "search", "locate", "where is", "list files"}) {
		return true, AgentTypeSearcher
	}

	// Review tasks
	if containsAny(taskLower, []string{"review", "critique", "analyze code"}) {
		return true, AgentTypeReviewer
	}

	// Complex coding tasks may spawn a specialist
	if containsAny(taskLower, []string{"implement", "build", "create feature"}) {
		// Check task complexity - if very long or multi-step, spawn coder agent
		if len(task) > 200 || strings.Count(task, ".") > 3 {
			return true, AgentTypeCoder
		}
	}

	return false, ""
}

// FormatOutputAsChecklist converts prose output to actionable checklist format
func (pm *ProtocolManager) FormatOutputAsChecklist(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inList := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Convert numbered or bulleted lists to checkboxes
		if matchesListFormat(trimmed) {
			checkboxed := convertToCheckbox(trimmed)
			result = append(result, checkboxed)
			inList = true
		} else if inList && trimmed == "" {
			result = append(result, "")
		} else if !inList {
			result = append(result, line)
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// LogAction logs an action to the progress file
func (pm *ProtocolManager) LogAction(action, result, nextStep string) error {
	return prompt.UpdateProgress(action, result, nextStep)
}

// LogError logs an error to the scratchpad
func (pm *ProtocolManager) LogError(entry string) error {
	return prompt.LogScratchpad(entry)
}

// GetMasterPlan returns the current master plan
func (pm *ProtocolManager) GetMasterPlan() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return prompt.LoadMasterPlan()
}

// GetStack returns the current tech stack
func (pm *ProtocolManager) GetStack() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return prompt.LoadStack()
}

// SafetyViolationError represents a safety rule violation
type SafetyViolationError struct {
	Rule    string
	Action  string
	Message string
}

func (e *SafetyViolationError) Error() string {
	return fmt.Sprintf("[%s] %s\nAction: %s", e.Rule, e.Message, e.Action)
}

// Helper functions

func isMainPush(action string) bool {
	return strings.Contains(action, "git push") &&
		(strings.Contains(action, "main") || strings.Contains(action, "master"))
}

func isMainCommit(action string) bool {
	return strings.Contains(action, "git checkout") &&
		(strings.Contains(action, "main") || strings.Contains(action, "master")) &&
		strings.Contains(action, "commit")
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func matchesListFormat(line string) bool {
	return strings.HasPrefix(line, "- ") ||
		strings.HasPrefix(line, "* ") ||
		(len(line) > 2 && line[0] >= '0' && line[0] <= '9' && (line[1] == '.' || line[1] == ')'))
}

func convertToCheckbox(line string) string {
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		return "- [ ] " + trimmed[2:]
	}

	// Numbered list: "1." or "1)"
	for i := 0; i < 10; i++ {
		prefix := fmt.Sprintf("%d.", i)
		if strings.HasPrefix(trimmed, prefix) {
			return "- [ ] " + trimmed[len(prefix):]
		}
		prefix = fmt.Sprintf("%d)", i)
		if strings.HasPrefix(trimmed, prefix) {
			return "- [ ] " + trimmed[len(prefix):]
		}
	}

	return line
}

func truncateLines(content string, maxLines int) string {
	lines := strings.Split(content, "\n")
	if len(lines) > maxLines {
		return strings.Join(lines[:maxLines], "\n") + "\n... (truncated)"
	}
	return content
}

// InitializeWorkspace ensures the .floyd directory exists
func (pm *ProtocolManager) InitializeWorkspace() error {
	return prompt.InitializeFloydDir()
}

// WorkspaceStatus returns the current workspace status
func (pm *ProtocolManager) WorkspaceStatus() map[string]string {
	status := make(map[string]string)

	// Check master plan
	if _, err := os.Stat(filepath.Join(pm.floydDir, "master_plan.md")); err == nil {
		status["master_plan"] = "exists"
	} else {
		status["master_plan"] = "missing"
	}

	// Check progress
	if info, err := os.Stat(filepath.Join(pm.floydDir, "progress.md")); err == nil {
		status["progress"] = info.ModTime().Format("2006-01-02 15:04")
	} else {
		status["progress"] = "missing"
	}

	// Check stack
	if _, err := os.Stat(filepath.Join(pm.floydDir, "stack.md")); err == nil {
		status["stack"] = "exists"
	} else {
		status["stack"] = "missing"
	}

	// Check protocol
	if _, err := os.Stat(filepath.Join(pm.floydDir, "prompts", "AGENT_INSTRUCTIONS.md")); err == nil {
		status["protocol"] = "loaded"
	} else {
		status["protocol"] = "missing"
	}

	return status
}

// EnhancedChatRequest wraps a ChatRequest with FLOYD protocol
func (pm *ProtocolManager) EnhancedChatRequest(req agent.ChatRequest, mode string) agent.ChatRequest {
	// Build the system prompt as a string (NOT as a message - Anthropic API rejects "system" role in messages)
	systemPrompt := prompt.LoadSystemPrompt(mode)

	// Add runtime environment context
	envContext := pm.buildEnvironmentContext()
	systemPrompt += "\n\n" + envContext

	// Add explicit tool usage instructions
	toolInstructions := pm.buildToolInstructions()
	systemPrompt += "\n\n" + toolInstructions

	// Add context to system prompt if available
	context := pm.GetContextForPrompt()
	if context != "" {
		systemPrompt += "\n\n=== CURRENT PROJECT CONTEXT ===\n" + context
	}

	// Set the System field (this is how Anthropic API expects system prompts)
	req.System = systemPrompt

	// Filter out any "system" role messages from the conversation (they're invalid in Anthropic API)
	filteredMessages := make([]agent.Message, 0, len(req.Messages))
	for _, msg := range req.Messages {
		if msg.Role != "system" {
			filteredMessages = append(filteredMessages, msg)
		}
	}
	req.Messages = filteredMessages

	// Add tools to request
	req.Tools = tools.BuildToolRegistry()

	return req
}

// buildEnvironmentContext creates runtime context about the current environment
func (pm *ProtocolManager) buildEnvironmentContext() string {
	var parts []string

	parts = append(parts, "=== RUNTIME ENVIRONMENT ===")

	// Current time (critical for the agent to know)
	now := time.Now()
	parts = append(parts, fmt.Sprintf("CURRENT_TIME: %s", now.Format("2006-01-02 15:04:05 MST")))
	parts = append(parts, fmt.Sprintf("TIMEZONE: %s", now.Location().String()))

	// OS info
	parts = append(parts, fmt.Sprintf("OS: %s", getOSInfo()))

	// Current working directory
	if cwd, err := os.Getwd(); err == nil {
		parts = append(parts, fmt.Sprintf("WORKING_DIRECTORY: %s", cwd))

		// Try to identify repo name
		repoName := filepath.Base(cwd)
		parts = append(parts, fmt.Sprintf("REPO_NAME: %s", repoName))

		// Check if .git exists
		if _, err := os.Stat(filepath.Join(cwd, ".git")); err == nil {
			parts = append(parts, "GIT_REPO: true")
			// Try to get current branch
			if branch := getCurrentGitBranch(cwd); branch != "" {
				parts = append(parts, fmt.Sprintf("GIT_BRANCH: %s", branch))
			}
		}

		// Check if .floyd exists
		if _, err := os.Stat(filepath.Join(cwd, ".floyd")); err == nil {
			parts = append(parts, "FLOYD_INITIALIZED: true")
		} else {
			parts = append(parts, "FLOYD_INITIALIZED: false (run /init to initialize)")
		}

		// Check for common project files
		projectType := detectProjectType(cwd)
		if projectType != "" {
			parts = append(parts, fmt.Sprintf("PROJECT_TYPE: %s", projectType))
		}
	}

	// User info
	if user := os.Getenv("USER"); user != "" {
		parts = append(parts, fmt.Sprintf("USER: %s", user))
	}

	// Home directory
	if home, err := os.UserHomeDir(); err == nil {
		parts = append(parts, fmt.Sprintf("HOME: %s", home))
	}

	// Shell
	if shell := os.Getenv("SHELL"); shell != "" {
		parts = append(parts, fmt.Sprintf("SHELL: %s", filepath.Base(shell)))
	}

	return strings.Join(parts, "\n")
}

// getOSInfo returns basic OS information
func getOSInfo() string {
	// Simple runtime detection
	switch {
	case os.Getenv("HOME") != "" && strings.HasPrefix(os.Getenv("HOME"), "/Users"):
		return "macOS"
	case os.Getenv("HOME") != "" && strings.HasPrefix(os.Getenv("HOME"), "/home"):
		return "Linux"
	default:
		return "Unix-like"
	}
}

// getCurrentGitBranch returns the current git branch or empty string
func getCurrentGitBranch(dir string) string {
	headPath := filepath.Join(dir, ".git", "HEAD")
	data, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}
	content := strings.TrimSpace(string(data))
	if strings.HasPrefix(content, "ref: refs/heads/") {
		return strings.TrimPrefix(content, "ref: refs/heads/")
	}
	return ""
}

// detectProjectType identifies the project type based on files present
func detectProjectType(dir string) string {
	checks := []struct {
		file     string
		projType string
	}{
		{"go.mod", "Go"},
		{"package.json", "Node.js"},
		{"Cargo.toml", "Rust"},
		{"requirements.txt", "Python"},
		{"pyproject.toml", "Python"},
		{"pom.xml", "Java/Maven"},
		{"build.gradle", "Java/Gradle"},
		{"Gemfile", "Ruby"},
		{"composer.json", "PHP"},
	}

	for _, check := range checks {
		if _, err := os.Stat(filepath.Join(dir, check.file)); err == nil {
			return check.projType
		}
	}
	return ""
}

// buildToolInstructions creates explicit instructions for tool usage
func (pm *ProtocolManager) buildToolInstructions() string {
	return `=== TOOL USAGE INSTRUCTIONS ===

You have access to the following tools. USE THEM to accomplish tasks - do not just talk about what you would do.

FILESYSTEM TOOLS:
- bash: Execute shell commands. Example: {"command": "ls -la"}
  Use for: running builds, git commands, npm/go commands, file operations
  ALWAYS check command output for errors before proceeding.

- read: Read file contents. Example: {"file_path": "/path/to/file.go"}
  Use for: understanding code, checking configurations, reviewing files
  Use "query" param to search within large files.

- write: Write/create files. Example: {"file_path": "/path/to/file.go", "content": "..."}
  Use for: creating new files, overwriting existing files completely

- edit: Edit files with find/replace. Example: {"file_path": "...", "old_string": "...", "new_string": "..."}
  Use for: surgical edits to existing files
  The old_string must match EXACTLY including whitespace.

- multiedit: Multiple edits in one file. Example: {"file_path": "...", "edits": [{"old_string": "...", "new_string": "..."}]}
  Use for: multiple non-contiguous changes in one file

- grep: Search files with regex. Example: {"pattern": "func.*Error", "path": ".", "glob": "*.go"}
  Use for: finding code patterns, locating functions, searching across files

- ls: List directory contents. Example: {"path": ".", "ignore": [".git", "node_modules"]}
  Use for: exploring project structure

- glob: Find files by pattern. Example: {"pattern": "**/*.go"}
  Use for: finding all files matching a pattern

WORKFLOW:
1. FIRST: Use ls and read to understand the codebase
2. PLAN: Think through the changes needed
3. EXECUTE: Use edit/write tools to make changes
4. VERIFY: Use bash to run tests/builds to verify your changes work
5. REPORT: Summarize what you did

COMMON COMMANDS BY PROJECT TYPE:
- Go: go build, go test ./..., go run main.go, go mod tidy
- Node.js: npm install, npm run dev, npm test, npm run build
- Python: pip install -r requirements.txt, python main.py, pytest
- Rust: cargo build, cargo test, cargo run
- Git: git status, git diff, git add ., git commit -m "msg", git checkout -b branch

USEFUL SHELL COMMANDS:
- Get current time: date
- Find files: find . -name "*.go" -type f
- Check disk space: df -h
- See running processes: ps aux
- Environment variables: env | grep KEY

BEST PRACTICES:
1. Read README.md first if it exists
2. Check for existing tests before making changes
3. Run the build/test after EVERY change
4. Make small, incremental changes
5. If something fails, read the error message carefully
6. Use grep to find where things are defined before editing
7. Never assume - always verify with tools

NEVER say "I would do X" - actually DO it using the tools.
ALWAYS verify your changes compile/work before reporting success.`
}
