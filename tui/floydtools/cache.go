package floydtools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Nomadcxx/sysc-Go/cache"
)

var (
	// Global cache manager instance
	cacheManager *cache.CacheManager
)

func init() {
	// Initialize cache manager with .floyd/.cache as root
	cacheManager = cache.NewCacheManager(".floyd/.cache")

	// Register the cache tool
	Register("cache", func() Tool {
		return &cacheToolAdapter{manager: cacheManager}
	})
}

// GetCacheManager returns the global cache manager
func GetCacheManager() *cache.CacheManager {
	return cacheManager
}

// cacheToolAdapter implements the Tool interface for cache operations
type cacheToolAdapter struct {
	manager *cache.CacheManager
}

func (t *cacheToolAdapter) Name() string {
	return "cache"
}

func (t *cacheToolAdapter) Description() string {
	return "Manage FLOYD's SUPERCACHE tiers (reasoning, project, vault). Operations: store, retrieve, list, clear, stats."
}

func (t *cacheToolAdapter) Validate(input string) error {
	// Basic validation - should be JSON with operation field
	// Try parsing as JSON first
	var op struct {
		Operation string `json:"operation"`
	}
	if err := json.Unmarshal([]byte(input), &op); err != nil {
		// Not JSON? Treat raw input as operation
		op.Operation = strings.TrimSpace(input)
	}

	validOps := map[string]bool{
		"store": true, "retrieve": true, "list": true,
		"clear": true, "stats": true,
	}

	if !validOps[op.Operation] {
		return fmt.Errorf("unknown operation: %s", op.Operation)
	}

	return nil
}

func (t *cacheToolAdapter) Run(input string) func(chan<- StreamMsg) {
	return func(out chan<- StreamMsg) {
		defer close(out)

		var op struct {
			Operation string `json:"operation"`
			Tier      string `json:"tier,omitempty"`
			Key       string `json:"key,omitempty"`
			Value     string `json:"value,omitempty"`
		}

		if err := json.Unmarshal([]byte(input), &op); err != nil {
			// Fallback: If not JSON, treat raw input as operation
			op.Operation = strings.TrimSpace(input)
		}

		ctx := context.Background()

		switch op.Operation {
		case "store":
			if op.Tier == "" || op.Key == "" || op.Value == "" {
				out <- StreamMsg{Status: "error", Chunk: "store requires tier, key, and value", Done: true}
				return
			}
			if err := t.manager.Store(ctx, op.Tier, op.Key, op.Value); err != nil {
				out <- StreamMsg{Status: "error", Chunk: err.Error(), Done: true}
				return
			}
			out <- StreamMsg{Status: "success", Chunk: fmt.Sprintf("Stored %s in %s tier", op.Key, op.Tier), Done: true}

		case "retrieve":
			if op.Tier == "" || op.Key == "" {
				out <- StreamMsg{Status: "error", Chunk: "retrieve requires tier and key", Done: true}
				return
			}
			value, err := t.manager.Retrieve(ctx, op.Tier, op.Key)
			if err != nil {
				out <- StreamMsg{Status: "error", Chunk: err.Error(), Done: true}
				return
			}
			out <- StreamMsg{Status: "success", Chunk: value, Done: true}

		case "list":
			if op.Tier == "" {
				out <- StreamMsg{Status: "error", Chunk: "list requires tier", Done: true}
				return
			}
			keys, err := t.manager.List(ctx, op.Tier)
			if err != nil {
				out <- StreamMsg{Status: "error", Chunk: err.Error(), Done: true}
				return
			}
			data, _ := json.Marshal(keys)
			out <- StreamMsg{Status: "success", Chunk: string(data), Done: true}

		case "clear":
			if op.Tier == "" {
				out <- StreamMsg{Status: "error", Chunk: "clear requires tier", Done: true}
				return
			}
			if err := t.manager.Clear(ctx, op.Tier); err != nil {
				out <- StreamMsg{Status: "error", Chunk: err.Error(), Done: true}
				return
			}
			out <- StreamMsg{Status: "success", Chunk: fmt.Sprintf("Cleared %s tier", op.Tier), Done: true}

		case "stats":
			stats := t.manager.GetStats()
			data, _ := json.MarshalIndent(stats, "", "  ")
			out <- StreamMsg{Status: "success", Chunk: string(data), Done: true}
		}
	}
}

func (t *cacheToolAdapter) FrameDelay() time.Duration {
	return 10 * time.Millisecond
}
