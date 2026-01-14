package floyd

import (
	"fmt"
	"regexp"
	"strings"
)

// OutputFormatter handles FLOYD output formatting
type OutputFormatter struct {
	preferChecklists bool
	compactMode      bool
}

// NewOutputFormatter creates a new output formatter
func NewOutputFormatter() *OutputFormatter {
	return &OutputFormatter{
		preferChecklists: true,
		compactMode:      false,
	}
}

// SetChecklistMode enables or disables checklist formatting
func (of *OutputFormatter) SetChecklistMode(enabled bool) {
	of.preferChecklists = enabled
}

// SetCompactMode enables or disables compact output
func (of *OutputFormatter) SetCompactMode(enabled bool) {
	of.compactMode = enabled
}

// ConvertToChecklist converts prose to checklist format
func (of *OutputFormatter) ConvertToChecklist(prose string) string {
	if !of.preferChecklists {
		return prose
	}

	lines := strings.Split(prose, "\n")
	var result []string
	inCodeBlock := false
	codeBlockFence := ""

	for _, line := range lines {
		// Track code blocks
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				codeBlockFence = line
			} else if strings.HasPrefix(line, codeBlockFence) {
				inCodeBlock = false
				codeBlockFence = ""
			}
			result = append(result, line)
			continue
		}

		// Don't modify code block contents
		if inCodeBlock {
			result = append(result, line)
			continue
		}

		// Convert list items to checkboxes
		converted := of.convertLine(line)
		result = append(result, converted)
	}

	return strings.TrimSuffix(strings.Join(result, "\n"), "\n")
}

// ExtractCommands extracts code blocks from text
func (of *OutputFormatter) ExtractCommands(text string) []string {
	var commands []string
	lines := strings.Split(text, "\n")
	inCodeBlock := false
	var currentBlock strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "```") && !inCodeBlock {
			inCodeBlock = true
			currentBlock.Reset()
			continue
		}

		if strings.HasPrefix(line, "```") && inCodeBlock {
			inCodeBlock = false
			block := strings.TrimSpace(currentBlock.String())
			if block != "" {
				commands = append(commands, block)
			}
			continue
		}

		if inCodeBlock {
			currentBlock.WriteString(line + "\n")
		}
	}

	return commands
}

// FormatActionable formats output to prioritize actionable content
func (of *OutputFormatter) FormatActionable(message string) string {
	// If compact mode, just return the message
	if of.compactMode {
		return message
	}

	// Extract key sections
	var parts []string

	// Add checklist for tasks
	checklist := of.ConvertToChecklist(message)
	if strings.Contains(checklist, "- [") {
		parts = append(parts, checklist)
	}

	// Add commands section
	commands := of.ExtractCommands(message)
	if len(commands) > 0 {
		parts = append(parts, "## Commands")
		for i, cmd := range commands {
			parts = append(parts, fmtCommand(i+1, cmd))
		}
	}

	if len(parts) == 0 {
		return message
	}

	return strings.Join(parts, "\n\n")
}

// FormatStatus formats a status update
func (of *OutputFormatter) FormatStatus(action, result string) string {
	return fmt.Sprintf("✓ %s\n\n%s", action, result)
}

// FormatError formats an error message
func (of *OutputFormatter) FormatError(err error) string {
	return fmt.Sprintf("❌ Error: %v", err)
}

// Helper methods

func (of *OutputFormatter) convertLine(line string) string {
	trimmed := strings.TrimSpace(line)

	// Empty lines
	if trimmed == "" {
		return line
	}

	// Already a checkbox
	if strings.HasPrefix(trimmed, "- [") {
		return line
	}

	// Convert bullet lists
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		return "- [ ] " + strings.TrimLeft(trimmed, "-* ")
	}

	// Convert numbered lists
	if matched, _ := regexp.MatchString(`^\d+[\.\)]\s`, trimmed); matched {
		re := regexp.MustCompile(`^(\d+[\.\)]\s+)`)
		return "- [ ] " + re.ReplaceAllString(trimmed, "")
	}

	// Convert task keywords
	taskPrefixes := []string{"TODO:", "TASK:", "To Do:", "• Do:"}
	for _, prefix := range taskPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return "- [ ] " + strings.TrimPrefix(trimmed, prefix)
		}
	}

	return line
}

func fmtCommand(num int, cmd string) string {
	return fmt.Sprintf("%d. ```bash\n%s\n```", num, strings.TrimSpace(cmd))
}

// DetectTaskType detects the type of task from a message
func DetectTaskType(message string) string {
	msgLower := strings.ToLower(message)

	if strings.Contains(msgLower, "implement") || strings.Contains(msgLower, "build") {
		return "implementation"
	}
	if strings.Contains(msgLower, "test") || strings.Contains(msgLower, "verify") {
		return "testing"
	}
	if strings.Contains(msgLower, "fix") || strings.Contains(msgLower, "bug") {
		return "bugfix"
	}
	if strings.Contains(msgLower, "refactor") || strings.Contains(msgLower, "clean up") {
		return "refactoring"
	}
	if strings.Contains(msgLower, "plan") || strings.Contains(msgLower, "design") {
		return "planning"
	}
	if strings.Contains(msgLower, "review") || strings.Contains(msgLower, "audit") {
		return "review"
	}

	return "general"
}

// EstimateComplexity estimates the complexity of a task
func EstimateComplexity(description string) string {
	// Count indicators
	wordCount := len(strings.Fields(description))
	lineCount := strings.Count(description, "\n")
	hasMultiple := strings.Contains(description, "and") || strings.Contains(description, ",")

	// Simple heuristic
	if wordCount < 20 && lineCount < 2 {
		return "simple"
	}
	if wordCount < 50 && lineCount < 5 && !hasMultiple {
		return "moderate"
	}
	if wordCount < 100 {
		return "complex"
	}
	return "very_complex"
}

// SuggestBreakdown suggests if a task should be broken down
func SuggestBreakdown(description string) bool {
	complexity := EstimateComplexity(description)
	return complexity == "complex" || complexity == "very_complex"
}

// CreateTaskSummary creates a formatted task summary
func CreateTaskSummary(title, description string) string {
	complexity := EstimateComplexity(description)
	taskType := DetectTaskType(description)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))
	sb.WriteString(fmt.Sprintf("**Type:** %s | **Complexity:** %s\n\n", taskType, complexity))
	sb.WriteString(description)
	sb.WriteString("\n\n---\n\n")
	sb.WriteString("**Status:** [ ] Not Started\n")
	sb.WriteString("**Estimated:** TBD\n")

	return sb.String()
}
