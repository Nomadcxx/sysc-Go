package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Pattern represents a reusable code pattern stored in the solution vault
type Pattern struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Category    string                 `json:"category"` // "registry", "animation", "tool", "ui"
	Description string                 `json:"description"`
	Language    string                 `json:"language"`
	Code        string                 `json:"code"`
	Template    string                 `json:"template,omitempty"`
	Usage       string                 `json:"usage"`
	Files       []string               `json:"files,omitempty"` // Example files
	Metadata    map[string]any         `json:"metadata"`
	Tags        []string               `json:"tags"`
	CreatedAt   time.Time              `json:"created_at"`
	UsedCount   int                    `json:"used_count"`
	LastUsed    time.Time              `json:"last_used"`
}

// VaultIndex represents the searchable index of the solution vault
type VaultIndex struct {
	Patterns    map[string]*Pattern    `json:"patterns"`
	Categories  map[string][]string    `json:"categories"`    // category -> pattern IDs
	Tags        map[string][]string    `json:"tags"`          // tag -> pattern IDs
	LastUpdated time.Time              `json:"last_updated"`
}

// VaultManager manages the solution vault (reusable patterns and solutions)
type VaultManager struct {
	rootDir    string
	patternDir string
	indexDir   string
	index      *VaultIndex
	mu         sync.RWMutex
}

// NewVaultManager creates a new vault manager
func NewVaultManager(rootDir string) *VaultManager {
	patternDir := filepath.Join(rootDir, "vault", "patterns")
	indexDir := filepath.Join(rootDir, "vault", "index")

	os.MkdirAll(patternDir, 0755)
	os.MkdirAll(indexDir, 0755)

	vm := &VaultManager{
		rootDir:    rootDir,
		patternDir: patternDir,
		indexDir:   indexDir,
		index: &VaultIndex{
			Patterns:   make(map[string]*Pattern),
			Categories: make(map[string][]string),
			Tags:       make(map[string][]string),
		},
	}

	// Load existing index
	vm.loadIndex()

	return vm
}

// StorePattern stores a new pattern in the vault
func (vm *VaultManager) StorePattern(pattern *Pattern) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// Set metadata
	if pattern.ID == "" {
		pattern.ID = generatePatternID()
	}
	pattern.CreatedAt = time.Now()
	pattern.LastUsed = time.Time{}
	pattern.UsedCount = 0

	// Store pattern file
	patternFile := filepath.Join(vm.patternDir, pattern.ID+".json")
	data, err := json.MarshalIndent(pattern, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal pattern: %w", err)
	}

	if err := os.WriteFile(patternFile, data, 0644); err != nil {
		return err
	}

	// Update index
	vm.index.Patterns[pattern.ID] = pattern
	vm.index.Categories[pattern.Category] = append(vm.index.Categories[pattern.Category], pattern.ID)
	for _, tag := range pattern.Tags {
		vm.index.Tags[tag] = append(vm.index.Tags[tag], pattern.ID)
	}
	vm.index.LastUpdated = time.Now()

	// Save index
	return vm.saveIndex()
}

// GetPattern retrieves a pattern by ID
func (vm *VaultManager) GetPattern(id string) (*Pattern, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	pattern, ok := vm.index.Patterns[id]
	if !ok {
		return nil, fmt.Errorf("pattern not found: %s", id)
	}

	// Increment usage
	go vm.recordUsage(id)

	return pattern, nil
}

// FindPatterns searches for patterns by category or tag
func (vm *VaultManager) FindPatterns(category string, tags []string) ([]*Pattern, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	results := make([]*Pattern, 0)
	seen := make(map[string]bool)

	// Search by category
	if category != "" {
		for _, id := range vm.index.Categories[category] {
			if !seen[id] {
				if pattern, ok := vm.index.Patterns[id]; ok {
					results = append(results, pattern)
					seen[id] = true
				}
			}
		}
	}

	// Search by tags
	for _, tag := range tags {
		for _, id := range vm.index.Tags[tag] {
			if !seen[id] {
				if pattern, ok := vm.index.Patterns[id]; ok {
					results = append(results, pattern)
					seen[id] = true
				}
			}
		}
	}

	return results, nil
}

// ListPatterns returns all patterns in a category
func (vm *VaultManager) ListPatterns(category string) ([]*Pattern, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if category == "" {
		// Return all patterns
		results := make([]*Pattern, 0, len(vm.index.Patterns))
		for _, pattern := range vm.index.Patterns {
			results = append(results, pattern)
		}
		return results, nil
	}

	results := make([]*Pattern, 0)
	for _, id := range vm.index.Categories[category] {
		if pattern, ok := vm.index.Patterns[id]; ok {
			results = append(results, pattern)
		}
	}

	return results, nil
}

// ListCategories returns all categories
func (vm *VaultManager) ListCategories() []string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	categories := make([]string, 0, len(vm.index.Categories))
	for cat := range vm.index.Categories {
		categories = append(categories, cat)
	}
	return categories
}

// DeletePattern removes a pattern from the vault
func (vm *VaultManager) DeletePattern(id string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	_, ok := vm.index.Patterns[id]
	if !ok {
		return fmt.Errorf("pattern not found: %s", id)
	}

	// Delete file
	patternFile := filepath.Join(vm.patternDir, id+".json")
	os.Remove(patternFile)

	// Update index
	delete(vm.index.Patterns, id)

	// Remove from categories
	for cat, ids := range vm.index.Categories {
		filtered := make([]string, 0)
		for _, pid := range ids {
			if pid != id {
				filtered = append(filtered, pid)
			}
		}
		vm.index.Categories[cat] = filtered
	}

	// Remove from tags
	for tag, ids := range vm.index.Tags {
		filtered := make([]string, 0)
		for _, pid := range ids {
			if pid != id {
				filtered = append(filtered, pid)
			}
		}
		vm.index.Tags[tag] = filtered
	}

	vm.index.LastUpdated = time.Now()
	return vm.saveIndex()
}

// recordUsage increments the usage count for a pattern
func (vm *VaultManager) recordUsage(id string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if pattern, ok := vm.index.Patterns[id]; ok {
		pattern.UsedCount++
		pattern.LastUsed = time.Now()
		vm.saveIndex()
	}
}

// saveIndex persists the index to disk
func (vm *VaultManager) saveIndex() error {
	indexFile := filepath.Join(vm.indexDir, "vault_index.json")
	data, err := json.MarshalIndent(vm.index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexFile, data, 0644)
}

// loadIndex loads the index from disk
func (vm *VaultManager) loadIndex() error {
	indexFile := filepath.Join(vm.indexDir, "vault_index.json")
	data, err := os.ReadFile(indexFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No index yet, that's ok
		}
		return err
	}

	return json.Unmarshal(data, vm.index)
}

// RebuildIndex rebuilds the index from pattern files
func (vm *VaultManager) RebuildIndex() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// Clear current index
	vm.index = &VaultIndex{
		Patterns:   make(map[string]*Pattern),
		Categories: make(map[string][]string),
		Tags:       make(map[string][]string),
	}

	// Scan pattern directory
	entries, err := os.ReadDir(vm.patternDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		// Load pattern
		patternFile := filepath.Join(vm.patternDir, entry.Name())
		data, err := os.ReadFile(patternFile)
		if err != nil {
			continue
		}

		var pattern Pattern
		if err := json.Unmarshal(data, &pattern); err != nil {
			continue
		}

		// Add to index
		vm.index.Patterns[pattern.ID] = &pattern
		vm.index.Categories[pattern.Category] = append(vm.index.Categories[pattern.Category], pattern.ID)
		for _, tag := range pattern.Tags {
			vm.index.Tags[tag] = append(vm.index.Tags[tag], pattern.ID)
		}
	}

	vm.index.LastUpdated = time.Now()
	return vm.saveIndex()
}

// GetStats returns vault statistics
func (vm *VaultManager) GetStats() map[string]any {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	stats := map[string]any{
		"total_patterns":    len(vm.index.Patterns),
		"total_categories":  len(vm.index.Categories),
		"total_tags":        len(vm.index.Tags),
		"last_updated":      vm.index.LastUpdated,
	}

	// Count by category
	categoryCounts := make(map[string]int)
	for cat, ids := range vm.index.Categories {
		categoryCounts[cat] = len(ids)
	}
	stats["by_category"] = categoryCounts

	return stats
}

// generatePatternID creates a unique pattern ID
func generatePatternID() string {
	return fmt.Sprintf("pattern_%d", time.Now().UnixNano())
}
