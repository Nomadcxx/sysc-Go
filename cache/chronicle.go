package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ChronicleEntry represents a single entry in the project chronicle
type ChronicleEntry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	SessionID string                 `json:"session_id"`
	Type      string                 `json:"type"` // "surgery", "feature", "bugfix", "refactor", "test"
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Metadata  map[string]any         `json:"metadata"`
	Tags      []string               `json:"tags"`
}

// ChronicleManager manages the project chronicle (persistent log of all work)
type ChronicleManager struct {
	rootDir    string
	summaryDir string
	contextDir string
	mu         sync.RWMutex
	sessionID  string
}

// NewChronicleManager creates a new chronicle manager
func NewChronicleManager(rootDir string) *ChronicleManager {
	summaryDir := filepath.Join(rootDir, "project", "phase_summaries")
	contextDir := filepath.Join(rootDir, "project", "context")

	os.MkdirAll(summaryDir, 0755)
	os.MkdirAll(contextDir, 0755)

	return &ChronicleManager{
		rootDir:    rootDir,
		summaryDir: summaryDir,
		contextDir: contextDir,
		sessionID:  generateSessionID(),
	}
}

// RecordEntry adds a new entry to the chronicle
func (cm *ChronicleManager) RecordEntry(entryType, title, content string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	entry := ChronicleEntry{
		ID:        generateEntryID(),
		Timestamp: time.Now(),
		SessionID: cm.sessionID,
		Type:      entryType,
		Title:     title,
		Content:   content,
		Metadata:  make(map[string]any),
		Tags:      make([]string, 0),
	}

	// Store in daily log file
	dailyFile := cm.dailyLogFilename()
	return cm.appendToFile(dailyFile, entry)
}

// RecordSurgery records a surgical operation
func (cm *ChronicleManager) RecordSurgery(title, description string, filesModified []string, impact string) error {
	metadata := map[string]any{
		"files_modified": filesModified,
		"impact":         impact,
	}

	entry := ChronicleEntry{
		ID:        generateEntryID(),
		Timestamp: time.Now(),
		SessionID: cm.sessionID,
		Type:      "surgery",
		Title:     title,
		Content:   description,
		Metadata:  metadata,
		Tags:      []string{"surgery", "refactor"},
	}

	dailyFile := cm.dailyLogFilename()
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.appendToFile(dailyFile, entry)
}

// RecordPhase records a phase completion
func (cm *ChronicleManager) RecordPhase(phaseNum int, title, description string, deliverables []string) error {
	phaseFile := filepath.Join(cm.summaryDir, fmt.Sprintf("phase_%d.json", phaseNum))

	summary := map[string]any{
		"phase":       phaseNum,
		"title":       title,
		"description": description,
		"deliverables": deliverables,
		"completed_at": time.Now().Format(time.RFC3339),
		"session_id":  cm.sessionID,
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(phaseFile, data, 0644)
}

// StoreContext stores context information for the current session
func (cm *ChronicleManager) StoreContext(key string, value any) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	contextFile := filepath.Join(cm.contextDir, cm.sessionID+".json")

	// Load existing context
	var context map[string]any
	if data, err := os.ReadFile(contextFile); err == nil {
		json.Unmarshal(data, &context)
	} else {
		context = make(map[string]any)
	}

	// Update
	context[key] = value
	context["updated_at"] = time.Now().Format(time.RFC3339)

	// Save
	data, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(contextFile, data, 0644)
}

// GetContext retrieves context information
func (cm *ChronicleManager) GetContext(key string) (any, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	contextFile := filepath.Join(cm.contextDir, cm.sessionID+".json")
	data, err := os.ReadFile(contextFile)
	if err != nil {
		return nil, false
	}

	var context map[string]any
	if err := json.Unmarshal(data, &context); err != nil {
		return nil, false
	}

	val, ok := context[key]
	return val, ok
}

// GetRecentEntries returns recent chronicle entries
func (cm *ChronicleManager) GetRecentEntries(n int) ([]ChronicleEntry, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	dailyFile := cm.dailyLogFilename()
	entries, err := cm.loadEntriesFromFile(dailyFile)
	if err != nil {
		return nil, err
	}

	// Return last n entries
	if len(entries) <= n {
		return entries, nil
	}

	return entries[len(entries)-n:], nil
}

// dailyLogFilename returns the filename for today's log
func (cm *ChronicleManager) dailyLogFilename() string {
	date := time.Now().Format("2006-01-02")
	return filepath.Join(cm.rootDir, "project", fmt.Sprintf("chronicle_%s.log", date))
}

// appendToFile appends an entry to a log file
func (cm *ChronicleManager) appendToFile(filename string, entry ChronicleEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Append to file
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(append(data, '\n'))
	return err
}

// loadEntriesFromFile loads entries from a log file
func (cm *ChronicleManager) loadEntriesFromFile(filename string) ([]ChronicleEntry, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []ChronicleEntry{}, nil
		}
		return nil, err
	}

	// File is newline-delimited JSON
	lines := splitLines(data)
	entries := make([]ChronicleEntry, 0)

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var entry ChronicleEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// splitLines splits data into lines
func splitLines(data []byte) [][]byte {
	lines := make([][]byte, 0)
	start := 0

	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}

	if start < len(data) {
		lines = append(lines, data[start:])
	}

	return lines
}

// generateEntryID creates a unique entry ID
func generateEntryID() string {
	return fmt.Sprintf("entry_%d", time.Now().UnixNano())
}
