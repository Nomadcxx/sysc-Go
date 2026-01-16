package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ReasoningFrame represents a single reasoning step in the agent's thought process
// Designed for GLM-optimized state preservation with "Preserved Thinking" mode
type ReasoningFrame struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	SessionID string                 `json:"session_id"`
	StepNumber int                   `json:"step_number"`

	// Core reasoning content
	Prompt    string                 `json:"prompt"`
	Response  string                 `json:"response"`
	TokensUsed int                   `json:"tokens_used"`

	// GLM-specific metadata
	GLMContext map[string]any         `json:"glm_context,omitempty"`

	// Cognition steps (chain of thought)
	CogSteps  []CogStep              `json:"cog_steps"`

	// Tool invocations
	ToolCalls  []ToolCall            `json:"tool_calls,omitempty"`

	// Associated context files
	ContextFiles []string            `json:"context_files,omitempty"`

	// Metadata for indexing
	Metadata  map[string]any         `json:"metadata,omitempty"`
}

// CogStep represents a single cognitive step in the reasoning chain
type CogStep struct {
	Number    int       `json:"number"`
	Type      string    `json:"type"` // "analysis", "planning", "execution", "validation"
	Thought   string    `json:"thought"`
	Data      any       `json:"data,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ToolCall represents a tool invocation during reasoning
type ToolCall struct {
	ToolName   string                 `json:"tool_name"`
	Arguments  map[string]any         `json:"arguments"`
	Result     string                 `json:"result,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Duration   time.Duration          `json:"duration"`
	Timestamp  time.Time              `json:"timestamp"`
}

// FrameManager manages reasoning frames with persistence
type FrameManager struct {
	rootDir    string
	activeDir  string
	archiveDir string
	sessionID  string
	mu         sync.RWMutex
	currentFrame *ReasoningFrame
}

// NewFrameManager creates a new frame manager
func NewFrameManager(rootDir string) *FrameManager {
	activeDir := filepath.Join(rootDir, "reasoning", "active")
	archiveDir := filepath.Join(rootDir, "reasoning", "archive")

	// Ensure directories exist
	os.MkdirAll(activeDir, 0755)
	os.MkdirAll(archiveDir, 0755)

	return &FrameManager{
		rootDir:    rootDir,
		activeDir:  activeDir,
		archiveDir: archiveDir,
		sessionID:  generateSessionID(),
	}
}

// StartFrame begins a new reasoning frame
func (fm *FrameManager) StartFrame(prompt string) *ReasoningFrame {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Archive current frame if exists
	if fm.currentFrame != nil {
		fm.archiveFrame(fm.currentFrame)
	}

	frame := &ReasoningFrame{
		ID:        generateFrameID(),
		Timestamp: time.Now(),
		SessionID: fm.sessionID,
		StepNumber: 0,
		Prompt:    prompt,
		GLMContext: make(map[string]any),
		CogSteps:  make([]CogStep, 0),
		ToolCalls: make([]ToolCall, 0),
		Metadata:  make(map[string]any),
	}

	fm.currentFrame = frame
	return frame
}

// AddCogStep adds a cognitive step to the current frame
func (fm *FrameManager) AddCogStep(stepType, thought string, data any) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.currentFrame == nil {
		return
	}

	step := CogStep{
		Number:    len(fm.currentFrame.CogSteps) + 1,
		Type:      stepType,
		Thought:   thought,
		Data:      data,
		Timestamp: time.Now(),
	}

	fm.currentFrame.CogSteps = append(fm.currentFrame.CogSteps, step)
	fm.currentFrame.StepNumber = step.Number
}

// AddToolCall records a tool invocation
func (fm *FrameManager) AddToolCall(toolName string, args map[string]any) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.currentFrame == nil {
		return
	}

	call := ToolCall{
		ToolName:  toolName,
		Arguments: args,
		Timestamp: time.Now(),
	}

	fm.currentFrame.ToolCalls = append(fm.currentFrame.ToolCalls, call)
}

// CompleteToolCall records the result of a tool invocation
func (fm *FrameManager) CompleteToolCall(toolName string, result string, err error, duration time.Duration) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.currentFrame == nil {
		return
	}

	// Find the last tool call with this name and update it
	for i := len(fm.currentFrame.ToolCalls) - 1; i >= 0; i-- {
		if fm.currentFrame.ToolCalls[i].ToolName == toolName && fm.currentFrame.ToolCalls[i].Result == "" {
			fm.currentFrame.ToolCalls[i].Result = result
			fm.currentFrame.ToolCalls[i].Duration = duration
			if err != nil {
				fm.currentFrame.ToolCalls[i].Error = err.Error()
			}
			return
		}
	}
}

// SetResponse sets the agent's response for the current frame
func (fm *FrameManager) SetResponse(response string, tokensUsed int) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.currentFrame == nil {
		return
	}

	fm.currentFrame.Response = response
	fm.currentFrame.TokensUsed = tokensUsed
}

// SaveFrame persists the current frame to disk
func (fm *FrameManager) SaveFrame() error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.currentFrame == nil {
		return fmt.Errorf("no active frame to save")
	}

	return fm.saveFrame(fm.currentFrame)
}

// saveFrame writes a frame to the active directory
func (fm *FrameManager) saveFrame(frame *ReasoningFrame) error {
	data, err := json.MarshalIndent(frame, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal frame: %w", err)
	}

	filename := filepath.Join(fm.activeDir, frame.ID+".json")
	return os.WriteFile(filename, data, 0644)
}

// archiveFrame moves a frame to the archive directory
func (fm *FrameManager) archiveFrame(frame *ReasoningFrame) error {
	srcPath := filepath.Join(fm.activeDir, frame.ID+".json")
	dstPath := filepath.Join(fm.archiveDir, frame.ID+".json")

	// Read the frame
	data, err := os.ReadFile(srcPath)
	if err != nil {
		// File doesn't exist, try to save first
		if err := fm.saveFrame(frame); err != nil {
			return err
		}
		data, err = os.ReadFile(srcPath)
		if err != nil {
			return err
		}
	}

	// Write to archive
	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		return err
	}

	// Remove from active
	os.Remove(srcPath)
	return nil
}

// GetActiveFrame returns the current active frame
func (fm *FrameManager) GetActiveFrame() *ReasoningFrame {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.currentFrame
}

// LoadFrame loads a frame from disk by ID
func (fm *FrameManager) LoadFrame(frameID string) (*ReasoningFrame, error) {
	// Try active first
	activePath := filepath.Join(fm.activeDir, frameID+".json")
	if data, err := os.ReadFile(activePath); err == nil {
		var frame ReasoningFrame
		if err := json.Unmarshal(data, &frame); err != nil {
			return nil, err
		}
		return &frame, nil
	}

	// Try archive
	archivePath := filepath.Join(fm.archiveDir, frameID+".json")
	data, err := os.ReadFile(archivePath)
	if err != nil {
		return nil, fmt.Errorf("frame not found: %s", frameID)
	}

	var frame ReasoningFrame
	if err := json.Unmarshal(data, &frame); err != nil {
		return nil, err
	}
	return &frame, nil
}

// ListFrames returns all frames in a directory
func (fm *FrameManager) ListFrames(archive bool) ([]*ReasoningFrame, error) {
	dir := fm.activeDir
	if archive {
		dir = fm.archiveDir
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	frames := make([]*ReasoningFrame, 0)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}

		var frame ReasoningFrame
		if err := json.Unmarshal(data, &frame); err != nil {
			continue
		}

		frames = append(frames, &frame)
	}

	return frames, nil
}

// CleanupOldFrames removes archived frames older than the specified duration
func (fm *FrameManager) CleanupOldFrames(olderThan time.Duration) (int, error) {
	frames, err := fm.ListFrames(true)
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-olderThan)
	removed := 0

	for _, frame := range frames {
		if frame.Timestamp.Before(cutoff) {
			path := filepath.Join(fm.archiveDir, frame.ID+".json")
			if err := os.Remove(path); err == nil {
				removed++
			}
		}
	}

	return removed, nil
}

// GetSessionID returns the current session ID
func (fm *FrameManager) GetSessionID() string {
	return fm.sessionID
}

// generateFrameID creates a unique frame ID
func generateFrameID() string {
	return fmt.Sprintf("frame_%d", time.Now().UnixNano())
}

// generateSessionID creates a session ID
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().Unix()/3600) // Changes hourly
}
