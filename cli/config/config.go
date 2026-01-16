package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Nomadcxx/sysc-Go/cache"
)

// Config holds the FLOYD CLI configuration
type Config struct {
	// API settings
	APIKey  string
	BaseURL string
	Model   string

	// Agent settings
	SystemPrompt string
	MaxTokens    int
	Temperature  float64

	// Cache
	CacheManager *cache.CacheManager
}

// Load loads configuration from file and environment
func Load(configPath string) *Config {
	cfg := &Config{
		BaseURL:      "https://api.z.ai/api/anthropic",
		Model:        "claude-opus-4",
		SystemPrompt: "You are FLOYD, a professional coding assistant. You are concise, helpful, and accurate. You write working code, avoid word salad, and get straight to the point.",
		MaxTokens:    8192,
		Temperature:  0.7,
	}

	// Try to read API key from Claude settings
	if cfg.APIKey == "" {
		homeDir, _ := os.UserHomeDir()
		settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
		data, err := os.ReadFile(settingsPath)
		if err == nil {
			var settings struct {
				Env struct {
					AnthropicAuthToken string `json:"ANTHROPIC_AUTH_TOKEN"`
				} `json:"env"`
			}
			if err := json.Unmarshal(data, &settings); err == nil {
				cfg.APIKey = settings.Env.AnthropicAuthToken
			}
		}
	}

	// Fallback to environment variable
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("ANTHROPIC_AUTH_TOKEN")
	}
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("GLM_API_KEY")
	}
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("ZHIPU_API_KEY")
	}

	// Initialize cache manager
	cfg.CacheManager = cache.NewCacheManager(".floyd/.cache")

	return cfg
}
