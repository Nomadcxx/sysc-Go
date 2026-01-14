package cache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// CacheManager provides caching operations for the agent
type CacheManager struct {
	rootDir    string
	frames     *FrameManager
	chronicle  *ChronicleManager
	vault      *VaultManager
	mu         sync.RWMutex
}

// NewCacheManager creates a new cache manager
func NewCacheManager(rootDir string) *CacheManager {
	return &CacheManager{
		rootDir:   rootDir,
		frames:    NewFrameManager(rootDir),
		chronicle: NewChronicleManager(rootDir),
		vault:     NewVaultManager(rootDir),
	}
}

// GetFrameManager returns the frame manager
func (cm *CacheManager) FrameManager() *FrameManager {
	return cm.frames
}

// GetChronicle returns the chronicle manager
func (cm *CacheManager) Chronicle() *ChronicleManager {
	return cm.chronicle
}

// GetVault returns the vault manager
func (cm *CacheManager) Vault() *VaultManager {
	return cm.vault
}

// CacheStoreOp stores a value in the cache
type CacheStoreOp struct {
	Tier  string `json:"tier"`  // "reasoning", "project", "vault"
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CacheRetrieveOp retrieves a value from the cache
type CacheRetrieveOp struct {
	Tier string `json:"tier"`
	Key  string `json:"key"`
}

// Store stores data in the specified cache tier
func (cm *CacheManager) Store(ctx context.Context, tier, key, value string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	switch tier {
	case "reasoning":
		// Store as a reasoning frame artifact
		dir := filepath.Join(cm.rootDir, "reasoning", "active")
		return cm.storeInDir(dir, key, value)

	case "project":
		// Store in project context
		dir := filepath.Join(cm.rootDir, "project", "context")
		return cm.storeInDir(dir, key, value)

	case "vault":
		// Store in solution vault
		dir := filepath.Join(cm.rootDir, "vault", "patterns")
		return cm.storeInDir(dir, key, value)

	default:
		return fmt.Errorf("unknown cache tier: %s", tier)
	}
}

// storeInDir stores a key-value pair in a directory
func (cm *CacheManager) storeInDir(dir, key, value string) error {
	os.MkdirAll(dir, 0755)
	filename := filepath.Join(dir, key+".json")
	return os.WriteFile(filename, []byte(value), 0644)
}

// Retrieve retrieves data from the specified cache tier
func (cm *CacheManager) Retrieve(ctx context.Context, tier, key string) (string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var dir string
	switch tier {
	case "reasoning":
		dir = filepath.Join(cm.rootDir, "reasoning", "active")
	case "project":
		dir = filepath.Join(cm.rootDir, "project", "context")
	case "vault":
		dir = filepath.Join(cm.rootDir, "vault", "patterns")
	default:
		return "", fmt.Errorf("unknown cache tier: %s", tier)
	}

	filename := filepath.Join(dir, key+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("key not found: %s", key)
	}

	return string(data), nil
}

// List lists all keys in a cache tier
func (cm *CacheManager) List(ctx context.Context, tier string) ([]string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var dir string
	switch tier {
	case "reasoning":
		dir = filepath.Join(cm.rootDir, "reasoning", "active")
	case "project":
		dir = filepath.Join(cm.rootDir, "project", "context")
	case "vault":
		dir = filepath.Join(cm.rootDir, "vault", "patterns")
	default:
		return nil, fmt.Errorf("unknown cache tier: %s", tier)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// Remove .json extension
		name := entry.Name()
		if len(name) > 5 && name[len(name)-5:] == ".json" {
			keys = append(keys, name[:len(name)-5])
		}
	}

	return keys, nil
}

// Clear removes all entries from a cache tier
func (cm *CacheManager) Clear(ctx context.Context, tier string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var dir string
	switch tier {
	case "reasoning":
		dir = filepath.Join(cm.rootDir, "reasoning", "active")
	case "project":
		dir = filepath.Join(cm.rootDir, "project", "context")
	case "vault":
		dir = filepath.Join(cm.rootDir, "vault", "patterns")
	default:
		return fmt.Errorf("unknown cache tier: %s", tier)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			os.Remove(filepath.Join(dir, entry.Name()))
		}
	}

	return nil
}

// GetStats returns statistics about the cache
func (cm *CacheManager) GetStats() map[string]any {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stats := make(map[string]any)

	// Count files in each tier
	tiers := []string{"reasoning", "project", "vault"}
	for _, tier := range tiers {
		var dir string
		switch tier {
		case "reasoning":
			dir = filepath.Join(cm.rootDir, "reasoning", "active")
		case "project":
			dir = filepath.Join(cm.rootDir, "project", "context")
		case "vault":
			dir = filepath.Join(cm.rootDir, "vault", "patterns")
		}

		entries, _ := os.ReadDir(dir)
		count := 0
		for _, e := range entries {
			if !e.IsDir() {
				count++
			}
		}
		stats[tier] = count
	}

	return stats
}
