package floyd

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// SafetyEnforcer enforces FLOYD safety rules
type SafetyEnforcer struct {
	mu            sync.RWMutex
	allowOverride bool
	violationLog  []SafetyViolation
}

// SafetyViolation represents a safety rule violation
type SafetyViolation struct {
	Rule      string
	Action    string
	Message   string
	Allowed   bool // true if overridden
	Timestamp string
}

// NewSafetyEnforcer creates a new safety enforcer
func NewSafetyEnforcer() *SafetyEnforcer {
	return &SafetyEnforcer{
		violationLog: make([]SafetyViolation, 0),
	}
}

// SetOverrideMode allows or disallows safety rule overrides
func (se *SafetyEnforcer) SetOverrideMode(allow bool) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.allowOverride = allow
}

// CheckAction checks if an action is safe to execute
// Returns an error if the action violates safety rules
func (se *SafetyEnforcer) CheckAction(action string) error {
	se.mu.Lock()
	defer se.mu.Unlock()

	actionLower := strings.ToLower(strings.TrimSpace(action))

	// Rule 1: Block pushes to main/master
	if se.isMainPush(actionLower) {
		violation := SafetyViolation{
			Rule:      "NO_MAIN_PUSH",
			Action:    action,
			Message:   "Direct push to main/master branch is blocked by FLOYD protocol",
			Allowed:   se.allowOverride,
			Timestamp: timestamp(),
		}
		se.violationLog = append(se.violationLog, violation)

		if se.allowOverride {
			return nil // Allowed due to override
		}
		return &SafetyError{Violation: violation}
	}

	// Rule 2: Block direct commits to main/master
	if se.isMainCommit(actionLower) {
		violation := SafetyViolation{
			Rule:      "NO_MAIN_COMMIT",
			Action:    action,
			Message:   "Direct commit to main/master is blocked. Work on a feature branch.",
			Allowed:   se.allowOverride,
			Timestamp: timestamp(),
		}
		se.violationLog = append(se.violationLog, violation)

		if se.allowOverride {
			return nil
		}
		return &SafetyError{Violation: violation}
	}

	// Rule 3: Block destructive operations without confirmation
	if se.isDestructive(actionLower) && !se.allowOverride {
		violation := SafetyViolation{
			Rule:      "DESTRUCTIVE_CONFIRMATION",
			Action:    action,
			Message:   "Destructive operation requires confirmation",
			Allowed:   false,
			Timestamp: timestamp(),
		}
		se.violationLog = append(se.violationLog, violation)
		return &SafetyError{Violation: violation}
	}

	return nil
}

// IsBranchAction returns true if the action is branch-related
func (se *SafetyEnforcer) IsBranchAction(action string) bool {
	actionLower := strings.ToLower(action)
	return strings.Contains(actionLower, "git checkout") ||
		strings.Contains(actionLower, "git switch") ||
		strings.Contains(actionLower, "git branch")
}

// RecommendBranch returns a suggested branch name for the action
func (se *SafetyEnforcer) RecommendBranch(action string) string {
	// If action seems to be related to a feature, suggest a branch name
	actionLower := strings.ToLower(action)

	if strings.Contains(actionLower, "fix") || strings.Contains(actionLower, "bug") {
		return "fix/$(date +%Y%m%d-%H%M%S)"
	}
	if strings.Contains(actionLower, "feature") || strings.Contains(actionLower, "add") {
		return "feat/$(date +%Y%m%d-%H%M%S)"
	}
	if strings.Contains(actionLower, "refactor") {
		return "refactor/$(date +%Y%m%d-%H%M%S)"
	}

	return "feature/$(date +%Y%m%d-%H%M%S)"
}

// GetViolationLog returns all logged violations
func (se *SafetyEnforcer) GetViolationLog() []SafetyViolation {
	se.mu.RLock()
	defer se.mu.RUnlock()

	log := make([]SafetyViolation, len(se.violationLog))
	copy(log, se.violationLog)
	return log
}

// ClearViolationLog clears the violation log
func (se *SafetyEnforcer) ClearViolationLog() {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.violationLog = make([]SafetyViolation, 0)
}

// Helper methods

func (se *SafetyEnforcer) isMainPush(action string) bool {
	return strings.Contains(action, "git push") &&
		(strings.Contains(action, "main") || strings.Contains(action, "master"))
}

func (se *SafetyEnforcer) isMainCommit(action string) bool {
	return (strings.Contains(action, "git commit") || strings.Contains(action, "git checkout")) &&
		(strings.Contains(action, "main") || strings.Contains(action, "master"))
}

func (se *SafetyEnforcer) isDestructive(action string) bool {
	destructive := []string{
		"rm -rf /",
		"rm -rf ~",
		"rm -rf /*",
		"mkfs",
		"dd if=/dev/zero",
		":(){ :|:& };:", // fork bomb
		"> /dev/sda",
	}

	for _, d := range destructive {
		if strings.Contains(action, d) {
			return true
		}
	}

	return false
}

// SafetyError represents a safety rule violation error
type SafetyError struct {
	Violation SafetyViolation
}

func (e *SafetyError) Error() string {
	return fmt.Sprintf("[SAFETY VIOLATION: %s] %s\nAction: %s", e.Violation.Rule, e.Violation.Message, e.Violation.Action)
}

// IsSafetyError checks if an error is a safety violation
func IsSafetyError(err error) bool {
	_, ok := err.(*SafetyError)
	return ok
}

// GitSafePush is a helper to ensure safe git pushing
// Returns the safe command to use
func GitSafePush(branch string) string {
	if branch == "" || branch == "main" || branch == "master" {
		return "error: cannot push directly to main/master"
	}
	return fmt.Sprintf("git push origin %s", branch)
}

// ValidateGitCommand validates a git command against safety rules
func ValidateGitCommand(cmd string) error {
	cmdLower := strings.ToLower(strings.TrimSpace(cmd))

	// Check for push to main/master
	if strings.Contains(cmdLower, "git push") && (strings.Contains(cmdLower, "main") || strings.Contains(cmdLower, "master")) {
		return fmt.Errorf("SAFETY: Cannot push to main/master. Use a feature branch.")
	}

	// Check for force push
	if strings.Contains(cmdLower, "git push") && strings.Contains(cmdLower, "--force") {
		return fmt.Errorf("SAFETY: Force push requires explicit confirmation")
	}

	// Check for reset on main branch
	if strings.Contains(cmdLower, "git reset") && (strings.Contains(cmdLower, "main") || strings.Contains(cmdLower, "master")) {
		return fmt.Errorf("SAFETY: Cannot reset main/master branch")
	}

	return nil
}

// timestamp returns the current timestamp in a standard format
func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
