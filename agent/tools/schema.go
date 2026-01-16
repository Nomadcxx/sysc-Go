package tools

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Nomadcxx/sysc-Go/agent"
	"github.com/Nomadcxx/sysc-Go/tui/floydtools"
)

var (
	// Cached tool definitions
	cachedToolDefs []agent.Tool
	cachedOnce     sync.Once
	cacheMutex    sync.RWMutex
)

// Tool represents the Anthropic API tool definition format
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// BuildToolRegistry converts all floydtools to Anthropic API format
func BuildToolRegistry() []agent.Tool {
	cacheMutex.RLock()
	if cachedToolDefs != nil {
		defer cacheMutex.RUnlock()
		return cachedToolDefs
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if cachedToolDefs != nil {
		return cachedToolDefs
	}

	allTools := floydtools.AllTools()
	toolDefs := make([]agent.Tool, 0, len(allTools))

	for _, floydTool := range allTools {
		anthropicTool := convertFloydToolToAnthropic(floydTool)
		if anthropicTool != nil {
			toolDefs = append(toolDefs, *anthropicTool)
		}
	}

	cachedToolDefs = toolDefs
	return toolDefs
}

// GetToolDefinitions returns a map of tool name -> tool definition
func GetToolDefinitions() map[string]agent.Tool {
	tools := BuildToolRegistry()
	result := make(map[string]agent.Tool)
	for _, tool := range tools {
		result[tool.Name] = tool
	}
	return result
}

// GetToolByName returns a specific tool definition by name
func GetToolByName(name string) (agent.Tool, bool) {
	tools := BuildToolRegistry()
	for _, tool := range tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return agent.Tool{}, false
}

// convertFloydToolToAnthropic converts a floydtools.Tool to Anthropic format
func convertFloydToolToAnthropic(floydTool floydtools.Tool) *agent.Tool {
	// Build input schema from tool's Validate method
	inputSchema := buildInputSchema(floydTool)

	return &agent.Tool{
		Name:        floydTool.Name(),
		Description: floydTool.Description(),
		InputSchema: inputSchema,
	}
}

// buildInputSchema creates an input schema from a tool's validation
// This generates proper JSON Schema for Anthropic's API
func buildInputSchema(tool floydtools.Tool) map[string]interface{} {
	// Get example input to infer schema
	exampleInput := getExampleInput(tool)

	schema := map[string]interface{}{
		"type":       "object",
		"properties":  make(map[string]interface{}),
		"required":   getRequiredFields(tool),
	}

	// Add properties from example input
	if props, ok := exampleInput.(map[string]interface{}); ok {
		for key, value := range props {
			schema["properties"].(map[string]interface{})[key] = inferPropertyType(value)
		}
	}

	return schema
}

// inferPropertyType infers JSON schema type from a value
func inferPropertyType(value interface{}) map[string]interface{} {
	switch v := value.(type) {
	case string:
		return map[string]interface{}{
			"type":        "string",
			"description": value,
		}
	case bool:
		return map[string]interface{}{
			"type":        "boolean",
			"description": "True or false",
		}
	case float64, int, int64:
		return map[string]interface{}{
			"type":        "number",
			"description": "Numeric value",
		}
	case []interface{}:
		return map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"description": "List of items",
		}
	case map[string]interface{}:
		return map[string]interface{}{
			"type":       "object",
			"properties": v,
			"description": "Object with nested properties",
		}
	default:
		return map[string]interface{}{
			"type":        "string",
			"description": "Any value",
		}
	}
}

// getRequiredFields extracts required fields from tool metadata
func getRequiredFields(tool floydtools.Tool) []string {
	// Common required fields for each tool type
	switch tool.Name() {
	case "bash":
		return []string{"command"}
	case "read":
		return []string{"file_path"}
	case "write":
		return []string{"file_path", "content"}
	case "edit":
		return []string{"file_path", "old_string", "new_string"}
	case "multiedit":
		return []string{"file_path", "edits"}
	case "grep":
		return []string{"pattern"}
	case "ls":
		return []string{"path"}
	case "glob":
		return []string{"pattern"}
	case "cache":
		return []string{"action"}
	default:
		return []string{}
	}
}

// getExampleInput returns example input for schema inference
func getExampleInput(tool floydtools.Tool) interface{} {
	switch tool.Name() {
	case "bash":
		return map[string]interface{}{
			"command":             "ls -la",
			"run_in_background": false,
			"timeout":             30000,
		}
	case "read":
		return map[string]interface{}{
			"file_path": "/path/to/file.txt",
			"query":     "optional search query",
			"offset":    0,
			"limit":     100,
		}
	case "write":
		return map[string]interface{}{
			"file_path": "/path/to/file.txt",
			"content":   "file contents",
		}
	case "edit":
		return map[string]interface{}{
			"file_path":   "/path/to/file.txt",
			"old_string":  "old text",
			"new_string":  "new text",
			"replace_all": false,
		}
	case "multiedit":
		return map[string]interface{}{
			"file_path": "/path/to/file.txt",
			"edits": []map[string]interface{}{
				{"old_string": "old1", "new_string": "new1"},
				{"old_string": "old2", "new_string": "new2"},
			},
		}
	case "grep":
		return map[string]interface{}{
			"pattern":      "search pattern",
			"path":         ".",
			"glob":         "*.go",
			"output_mode":  "content",
		}
	case "ls":
		return map[string]interface{}{
			"path":    ".",
			"ignore":  []string{".git", "node_modules"},
		}
	case "glob":
		return map[string]interface{}{
			"pattern": "**/*.go",
		}
	case "cache":
		return map[string]interface{}{
			"action":   "stats",
			"key":      "optional key",
			"value":    "optional value",
			"tier":     "reasoning",
		}
	default:
		return map[string]interface{}{}
	}
}

// CreateToolDefinition creates an Anthropic tool definition from components
func CreateToolDefinition(name, description string, properties map[string]interface{}) agent.Tool {
	// Extract required fields from properties if available
	var required []string
	if req, ok := properties["required"].([]string); ok {
		required = req
		delete(properties, "required")
	}

	return agent.Tool{
		Name:        name,
		Description: description,
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties":  properties,
			"required":   required,
		},
	}
}

// BuildToolSchemasJSON returns all tool definitions as JSON
func BuildToolSchemasJSON() (string, error) {
	tools := BuildToolRegistry()
	data, err := json.MarshalIndent(tools, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal tools: %w", err)
	}
	return string(data), nil
}

// AddComputerUseTool adds the computer use tool definition
func AddComputerUseTool(tools []agent.Tool) []agent.Tool {
	computerTool := agent.Tool{
		Name:        "computer",
		Description: "Interact with the computer desktop environment. Can take screenshots, type text, click buttons, and more.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action": map[string]interface{}{
					"type": "string",
					"enum": []string{"screenshot", "left_click", "right_click", "type", "key", "mouse_move"},
					"description": "The action to perform",
				},
				"coordinate": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{"type": "number"},
					"description": "[x, y] coordinates for click actions",
				},
				"text": map[string]interface{}{
					"type":        "string",
					"description": "Text to type or key to press",
				},
			},
			"required": []string{"action"},
		},
	}
	return append(tools, computerTool)
}

// InvalidateCache clears the cached tool definitions
// Call this after registering new tools
func InvalidateCache() {
	cacheMutex.Lock()
	cachedToolDefs = nil
	cacheMutex.Unlock()
}
