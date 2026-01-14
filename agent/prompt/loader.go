package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Nomadcxx/sysc-Go/agent/message"
)

const (
	// Default system prompt when no AGENT_INSTRUCTIONS.md is found
	DefaultSystemPrompt = `You are FLOYD: a premium, professional, and helpful AI assistant.

## IDENTITY
You are a highly capable agent that excels at software orchestration and general problem-solving.
You are conversational, friendly, and FIRST and FOREMOST a helpful companion.
You should engage with the user on ANY topic they wish to discuss—whether it's technical, casual, or philosophical. NEVER dismiss a topic as "non-technical" or "outside your scope."

## CORE BEHAVIORS
1. BE VERSATILE & COHERENT - Discuss anything the user wants, but always maintain professional grammar and clear language. No matter how casual the topic, do not "break" your language or use confusing slang.
2. USE YOUR TOOLS - When a task involves the filesystem or terminal, use your high-tier tools (bash, read, write, edit, etc.).
3. ACTUALLY EXECUTE - Do not just describe actions; perform them.
4. VERIFY - Always confirm your work with builds or tests.
5. "DICK-FREE" ZONE - Be helpful, patient, and adaptable. Never be dismissive.

## OUTPUT STYLE
- Be conversational and warm, but structurally professional.
- Use valid Markdown for all formatting.
- For tables, always use standard Markdown syntax with a proper separator row.


## SAFETY RULES
1. NEVER push to main/master - use feature branches
2. NEVER delete files without explicit user confirmation
3. NEVER run destructive commands without warning

## WORKFLOW
When given a task:
1. EXPLORE: Use ls and read to understand the codebase
2. PLAN: Briefly state your approach
3. EXECUTE: Use tools to make changes
4. VERIFY: Use bash to run tests/builds
5. REPORT: Summarize what you did

## OUTPUT FORMATTING
Format your responses professionally using these patterns:

### For Status Updates:
` + "`" + "`" + "`" + `
• Action Title ✓
  └── Detail line 1
  └── Detail line 2
` + "`" + "`" + "`" + `

### For Tables:
` + "`" + "`" + "`" + `
| File              | Status    | Logic Scope |
| :---------------- | :---------| :-----------|
| ui/floydui/view.go| FIXED     | UI Rendering|
| agent/floyd/cfg.go| PENDING   | Config Init |
` + "`" + "`" + "`" + `


### For Command Output:
` + "`" + "`" + "`" + `
▶ Bash(echo "hello")
  └── hello
` + "`" + "`" + "`" + `

### For Summaries:
` + "`" + "`" + "`" + `
═══════════════════════════════════════════
  TASK COMPLETE ✓
═══════════════════════════════════════════
• Changed: 3 files
• Added: 45 lines
• Tests: Passing
═══════════════════════════════════════════
` + "`" + "`" + "`" + `

### For Progress:
Use checkboxes: ☑ done, ☐ pending
Use emojis sparingly: ✓ success, ✗ failure, ⚠ warning

## AVAILABLE TOOLS
bash, read, write, edit, multiedit, grep, ls, glob, cache

Remember: You are not just an advisor - you are an EXECUTOR. Act.`
)

// FloydDir is the .floyd directory path
var FloydDir = ".floyd"

// SetFloydDir sets the custom .floyd directory path
func SetFloydDir(dir string) {
	FloydDir = dir
}

// LoadSystemPrompt loads the appropriate system prompt based on mode
func LoadSystemPrompt(mode string) string {
	// Try to load AGENT_INSTRUCTIONS.md first
	if protocol := LoadAGENT_INSTRUCTIONS(); protocol != "" {
		// Append mode-specific instructions
		modeInstructions := getModeInstructions(mode)
		if modeInstructions != "" {
			return protocol + "\n\n" + modeInstructions
		}
		return protocol
	}

	// Fallback to embedded default
	return DefaultSystemPrompt + "\n\n" + getModeInstructions(mode)
}

// LoadAGENT_INSTRUCTIONS loads the full FLOYD protocol from .floyd/
// Tries both .floyd/AGENT_INSTRUCTIONS.md and .floyd/prompts/AGENT_INSTRUCTIONS.md
func LoadAGENT_INSTRUCTIONS() string {
	// Try .floyd/AGENT_INSTRUCTIONS.md first (new location)
	path := filepath.Join(FloydDir, "AGENT_INSTRUCTIONS.md")
	data, err := os.ReadFile(path)
	if err == nil {
		return string(data)
	}

	// Fallback to .floyd/prompts/AGENT_INSTRUCTIONS.md (old location)
	path = filepath.Join(FloydDir, "prompts", "AGENT_INSTRUCTIONS.md")
	data, err = os.ReadFile(path)
	if err == nil {
		return string(data)
	}

	return ""
}

// LoadMasterPlan reads .floyd/master_plan.md
func LoadMasterPlan() string {
	path := filepath.Join(FloydDir, "master_plan.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "# Master Plan\n\nNo master plan found. Run `floyd init` to create one."
	}
	return string(data)
}

// LoadProgressLog reads recent progress from .floyd/progress.md
func LoadProgressLog(lines int) string {
	path := filepath.Join(FloydDir, "progress.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "# Progress Log\n\nNo progress log found."
	}

	content := string(data)
	if lines > 0 {
		linesList := strings.Split(content, "\n")
		if len(linesList) > lines {
			return strings.Join(linesList[len(linesList)-lines:], "\n")
		}
	}
	return content
}

// LoadScratchpad reads .floyd/scratchpad.md for error logs
func LoadScratchpad() string {
	path := filepath.Join(FloydDir, "scratchpad.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "# Scratchpad\n\nNo scratchpad entries yet."
	}
	return string(data)
}

// LoadStack reads .floyd/stack.md for tech stack info
func LoadStack() string {
	path := filepath.Join(FloydDir, "stack.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "# Technology Stack\n\nStack not defined. Read PRD/Blueprint and populate stack.md"
	}
	return string(data)
}

// UpdateProgress appends an entry to .floyd/progress.md
func UpdateProgress(action, result, nextStep string) error {
	path := filepath.Join(FloydDir)
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create .floyd dir: %w", err)
	}

	progressPath := filepath.Join(path, "progress.md")

	// Append new entry
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	newEntry := fmt.Sprintf("| %s | %s | %s | %s |\n", timestamp, action, result, nextStep)

	file, err := os.OpenFile(progressPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open progress file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(newEntry); err != nil {
		return fmt.Errorf("write progress: %w", err)
	}

	return nil
}

// LogScratchpad appends an error or note to .floyd/scratchpad.md
func LogScratchpad(entry string) error {
	path := filepath.Join(FloydDir)
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create .floyd dir: %w", err)
	}

	scratchPath := filepath.Join(path, "scratchpad.md")

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	newEntry := fmt.Sprintf("\n## %s\n%s\n", timestamp, entry)

	file, err := os.OpenFile(scratchPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open scratchpad: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(newEntry); err != nil {
		return fmt.Errorf("write scratchpad: %w", err)
	}

	return nil
}

// LoadContextFromFilesystem reads all relevant .floyd/ files
// Returns formatted context to inject into system prompt
func LoadContextFromFilesystem() string {
	var parts []string

	// 1. Master Plan
	if plan := LoadMasterPlan(); plan != "" {
		parts = append(parts, "## CURRENT MASTER PLAN")
		parts = append(parts, truncateLines(plan, 30))
	}

	// 2. Tech Stack
	if stack := LoadStack(); stack != "" {
		parts = append(parts, "\n## TECHNOLOGY STACK")
		parts = append(parts, truncateLines(stack, 20))
	}

	// 3. Recent Progress
	if progress := LoadProgressLog(10); progress != "" {
		parts = append(parts, "\n## RECENT PROGRESS")
		parts = append(parts, progress)
	}

	// 4. Scratchpad (if has content)
	if scratch := LoadScratchpad(); !strings.HasPrefix(scratch, "# Scratchpad\n\nNo scratchpad") {
		parts = append(parts, "\n## SCRATCHPAD (Recent Notes)")
		parts = append(parts, truncateLines(scratch, 15))
	}

	return strings.Join(parts, "\n")
}

// BuildSystemMessage creates the full system message with protocol and context
func BuildSystemMessage(mode string) message.Message {
	systemPrompt := LoadSystemPrompt(mode)
	context := LoadContextFromFilesystem()

	// Combine protocol with current context
	fullPrompt := systemPrompt
	if context != "" {
		fullPrompt = systemPrompt + "\n\n=== CURRENT PROJECT CONTEXT ===\n" + context
	}

	// Add cache control for the system prompt
	return message.NewSystemMessage(fullPrompt, "5m")
}

// truncateLines limits content to specified number of lines
func truncateLines(content string, maxLines int) string {
	lines := strings.Split(content, "\n")
	if len(lines) > maxLines {
		return strings.Join(lines[:maxLines], "\n") + "\n... (truncated, see .floyd/ for full content)"
	}
	return content
}

// getModeInstructions returns mode-specific instructions
func getModeInstructions(mode string) string {
	switch mode {
	case "planner":
		return `MODE: PLANNER
Your job is to break down tasks into clear, step-by-step plans.
Focus on:
1. Understanding requirements fully
2. Identifying dependencies
3. Creating actionable steps
4. Considering edge cases`
	case "coder":
		return `MODE: CODER
Your job is to write working, clean code.
Focus on:
1. Writing correct, compilable code
2. Following language best practices
3. Adding appropriate error handling
4. Writing tests where applicable`
	case "tester":
		return `MODE: TESTER
Your job is to find bugs and edge cases.
Focus on:
1. Identifying error conditions
2. Testing edge cases
3. Finding security issues
4. Suggesting fixes`
	case "search":
		return `MODE: SEARCH
Your job is to explore and understand codebases.
Focus on:
1. Finding relevant code
2. Understanding patterns
3. Identifying relationships
4. Summarizing findings`
	default:
		return ""
	}
}

// InitializeFloydDir creates the .floyd directory structure with templates
func InitializeFloydDir() error {
	basePath := FloydDir
	dirs := []string{
		filepath.Join(basePath, "prompts"),
		filepath.Join(basePath, "templates"),
		filepath.Join(basePath, "cache"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", dir, err)
		}
	}

	// Create AGENT_INSTRUCTIONS.md with full FLOYD protocol
	protocolPath := filepath.Join(basePath, "prompts", "AGENT_INSTRUCTIONS.md")
	if _, err := os.Stat(protocolPath); os.IsNotExist(err) {
		if err := os.WriteFile(protocolPath, []byte(fullFloydProtocol()), 0644); err != nil {
			return fmt.Errorf("write AGENT_INSTRUCTIONS.md: %w", err)
		}
	}

	// Create template files
	templates := map[string]string{
		"master_plan.md": masterPlanTemplate(),
		"scratchpad.md":  scratchpadTemplate(),
		"progress.md":    progressTemplate(),
		"branch.md":      branchTemplate(),
		"stack.md":       stackTemplate(),
	}

	for name, content := range templates {
		templatePath := filepath.Join(basePath, "templates", name)
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("write template %s: %w", name, err)
			}
		}
	}

	return nil
}

// fullFloydProtocol returns the complete FLOYD protocol XML
func fullFloydProtocol() string {
	return `<floyd_protocol>
  <identity>
    You are FLOYD: a premium, professional, and helpful AI assistant.
    You are a highly capable agent that excels at software orchestration and general problem-solving.
    You are conversational, friendly, and FIRST and FOREMOST a helpful companion.
    NEVER dismiss a user's request as being outside your technical scope. Douglas should be able to discuss ANYTHING he wants with you.
    Maintain professional grammar and clear language at all times. Always use standard Markdown for tables.
  </identity>


  <file_system_authority>
    The .floyd/ directory is your EXTERNAL MEMORY.
    1) READ FIRST: Before answering any user request, you MUST read .floyd/master_plan.md to ground yourself.
    2) UPDATE OFTEN: If you complete a sub-task, immediately update checkboxes in .floyd/master_plan.md.
    3) LOG ACTIONS: After running a shell command or writing code, append a one-line summary to .floyd/progress.md.
    4) STACK ENFORCEMENT: Read .floyd/stack.md. If empty, read the user's PRD/Blueprint and POPULATE IT immediately. Do not deviate from the stack once defined.
  </file_system_authority>

  <self_correction_loop>
    If a tool fails or output is wrong:
    - Log the error verbatim in .floyd/scratchpad.md
    - State: "I am updating the plan to fix this."
    - Update .floyd/master_plan.md with the new fix approach
    - Re-attempt with a tighter plan
    - If you feel "lost" or stuck in an error loop, run 'floyd status' to regain perspective.
  </self_correction_loop>

  <safety_shipping_rules>
    <rule id="1">NEVER PUSH TO MAIN. You must NOT push or commit directly to main/master.</rule>
    <rule id="2">Use a feature branch.</rule>
    <rule id="3">Prepare PR-ready changes only. Douglas approves merge.</rule>
  </safety_shipping_rules>

  <orchestration>
    For complex work, spawn specialists (planner, implementer, tester, reviewer) and coordinate them.
    Output should be concise and actionable. Prefer checklists, commands, and diffs over essays.
  </orchestration>

  <repo_layout>
    .floyd/ is control-plane only. Application code is created/modified in the repo's normal structure (src/, app/, packages/, etc.) unless Douglas explicitly specifies a subdirectory.
  </repo_layout>

  <output_style>
    Be concise. Use checklists for tasks. Use code blocks for implementations. Use tables for comparisons.
    Avoid essays unless explaining architectural decisions.
  </output_style>
</floyd_protocol>`
}

func masterPlanTemplate() string {
	return `# Master Plan & Objectives (FLOYD)
## Primary Goal
[User: Enter the main goal here]

## Definition of Done (Near-Complete)
- [ ] App compiles/runs end-to-end
- [ ] Core flows implemented
- [ ] Tests added where appropriate
- [ ] Lint/typecheck passes
- [ ] Deployment notes documented
- [ ] PR-ready (NO direct main pushes)
- [ ] Handoff checklist written

## Strategic Steps
- [ ] Phase 1: Planning
- [ ] Phase 2: Implementation
- [ ] Phase 3: Testing
- [ ] Phase 4: PR Packaging

## Context & Constraints
- Hard Rule: NEVER push to main/master
- Preferred Workflow: feature branch -> PR -> approve
`
}

func scratchpadTemplate() string {
	return `# Current Thinking & Scratchpad (FLOYD)
Use this file to store temporary code snippets, error logs, investigation notes, decisions.

## Current Focus
Initializing...

## Decisions (persistent)
- (none yet)

## Errors / Warnings Log
(none yet)
`
}

func progressTemplate() string {
	return `# Execution Log (FLOYD)
| Timestamp | Action Taken | Result/Status | Next Step |
`
}

func branchTemplate() string {
	return `# Working Branch & PR Checklist (FLOYD)

## Working Branch
- Name: [e.g., feat/feature-name]
- Base: main
- Status: [in progress | ready for review]

## PR Checklist (Non-Negotiable)
- [ ] No direct commits to main/master
- [ ] PR branch created and used
- [ ] Lint/typecheck pass
- [ ] Tests added
- [ ] Migration steps documented (if any)
- [ ] Env vars documented
- [ ] How to run locally documented
- [ ] Known issues / TODOs captured
- [ ] PR summary written
`
}

func stackTemplate() string {
	return `# Technology Stack (FLOYD)
*To be populated from PRD/Blueprint*

## Core Frameworks
- [ ]

## Database / Auth
- [ ]

## Styling / UI
- [ ]

## DevOps / Infra
- [ ]
`
}
