package cache

import (
	"strings"

	"github.com/Nomadcxx/sysc-Go/agent/message"
)

const (
	// DefaultCacheTTL is the default cache TTL for system prompts
	DefaultCacheTTL = "5m"

	// LongCacheTTL is for long-lived content like master plans
	LongCacheTTL = "1h"

	// PromptCachingBetaHeader is the beta header for prompt caching
	PromptCachingBetaHeader = "prompt-caching-2024-01-01"

	// MaxCacheableContentSize is the minimum size for content to be cacheable
	// Content blocks smaller than this won't benefit from caching
	MinCacheableSize = 1024 // 1KB
)

// CacheControl manages cache control headers for Anthropic API
type CacheControl struct {
	enabled bool
	ttl     string
}

// NewCacheControl creates a new cache controller
func NewCacheControl(ttl string) *CacheControl {
	if ttl == "" {
		ttl = DefaultCacheTTL
	}
	return &CacheControl{
		enabled: true,
		ttl:     ttl,
	}
}

// Disabled returns a disabled cache controller
func Disabled() *CacheControl {
	return &CacheControl{
		enabled: false,
	}
}

// AddCacheControl adds cache control to content blocks
// Only blocks larger than MinCacheableSize get cache control
func AddCacheControl(blocks []message.ContentBlock, ttl string) []message.ContentBlock {
	result := make([]message.ContentBlock, len(blocks))

	for i, block := range blocks {
		result[i] = block

		// Add cache control to text blocks over minimum size
		if block.Type == "text" && len(block.Text) >= MinCacheableSize {
			result[i].CacheControl = &message.CacheControl{
				Type: "ephemeral",
				TTL:  ttl,
			}
		}
	}

	return result
}

// MarkSystemPromptForCache marks a system prompt for caching
func MarkSystemPromptForCache(system []message.ContentBlock) []message.ContentBlock {
	return AddCacheControl(system, DefaultCacheTTL)
}

// MarkLongContentForCache marks long-lived content for caching
func MarkLongContentForCache(content []message.ContentBlock) []message.ContentBlock {
	return AddCacheControl(content, LongCacheTTL)
}

// BuildCacheHeader constructs the anthropic-beta header for caching
func BuildCacheHeader() string {
	return PromptCachingBetaHeader
}

// IsCacheable returns true if content should be cached
func IsCacheable(content string) bool {
	return len(content) >= MinCacheableSize
}

// CacheStats represents cache read/write statistics from API responses
type CacheStats struct {
	CacheReadInputTokens   int
	CacheWriteInputTokens int
}

// GetSavings returns the number of tokens saved by caching
func (s CacheStats) GetSavings() int {
	return s.CacheReadInputTokens
}

// GetCacheHitRate returns the cache hit rate as a percentage
func (s CacheStats) GetCacheHitRate() float64 {
	total := s.CacheReadInputTokens + s.CacheWriteInputTokens
	if total == 0 {
		return 0
	}
	return float64(s.CacheReadInputTokens) / float64(total) * 100
}

// EstimateCacheCost estimates the cost savings from caching
// Cache reads are 90% cheaper than cache writes
func (s CacheStats) EstimateCacheCost(writeCostPerToken float64) float64 {
	readCost := float64(s.CacheReadInputTokens) * writeCostPerToken * 0.1
	writeCost := float64(s.CacheWriteInputTokens) * writeCostPerToken
	return readCost + writeCost
}

// CreateCacheControlBlock creates a cache control block
func CreateCacheControlBlock(ttl string) message.CacheControl {
	if ttl == "" {
		ttl = DefaultCacheTTL
	}
	return message.CacheControl{
		Type: "ephemeral",
		TTL:  ttl,
	}
}

// CreateTextBlockWithCache creates a text content block with cache control
func CreateTextBlockWithCache(text, ttl string) message.ContentBlock {
	block := message.ContentBlock{
		Type: "text",
		Text: text,
	}

	if IsCacheable(text) {
		block.CacheControl = &message.CacheControl{
			Type: "ephemeral",
			TTL:  ttl,
		}
	}

	return block
}

// PrepareSystemForCache prepares system content blocks for optimal caching
// Breaks large system prompts into cacheable chunks
func PrepareSystemForCache(systemPrompt string) []message.ContentBlock {
	// If the prompt is small, return as single block without cache
	if len(systemPrompt) < MinCacheableSize {
		return []message.ContentBlock{
			{Type: "text", Text: systemPrompt},
		}
	}

	// For large prompts, split into sections and mark cacheable
	sections := splitIntoCacheableSections(systemPrompt)
	blocks := make([]message.ContentBlock, 0, len(sections))

	for _, section := range sections {
		if IsCacheable(section) {
			cc := CreateCacheControlBlock(DefaultCacheTTL)
			blocks = append(blocks, message.ContentBlock{
				Type:         "text",
				Text:         section,
				CacheControl: &cc,
			})
		} else {
			blocks = append(blocks, message.ContentBlock{
				Type: "text",
				Text: section,
			})
		}
	}

	return blocks
}

// splitIntoCacheableSections splits a large prompt into cacheable sections
func splitIntoCacheableSections(prompt string) []string {
	// Split by double newlines (paragraphs)
	paragraphs := strings.Split(prompt, "\n\n")

	var sections []string
	var current strings.Builder

	for _, para := range paragraphs {
		testLength := current.Len() + len(para) + 2 // +2 for "\n\n"

		if testLength > 4096 && current.Len() > 0 {
			// Section is getting too large, finalize current
			sections = append(sections, strings.TrimSpace(current.String()))
			current.Reset()
		}

		current.WriteString(para)
		current.WriteString("\n\n")
	}

	if current.Len() > 0 {
		sections = append(sections, strings.TrimSpace(current.String()))
	}

	return sections
}
