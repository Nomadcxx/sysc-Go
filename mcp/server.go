// Package mcp provides an in-process MCP (Model Context Protocol) server
// that wraps the existing floydtools for Agent Engine consumption.
//
// This is a simplified in-process interface, not a full JSON-RPC server.
// The Agent Engine can call it directly as a Go package.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Nomadcxx/sysc-Go/tui/floydtools"
)

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ToolCall represents a tool execution request.
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// Server wraps floydtools for MCP access.
type Server struct {
	mu    sync.RWMutex
	tools map[string]ToolHandler
}

// ToolHandler is the interface for tool execution.
type ToolHandler interface {
	// Execute runs the tool with the given arguments.
	Execute(ctx context.Context, args map[string]interface{}) (string, error)

	// Schema returns the tool's definition.
	Schema() Tool
}

// adapter wraps a floydtools.Tool to implement ToolHandler.
type adapter struct {
	tool floydtools.Tool
}

// Execute adapts the floydtools streaming interface to a simple request/response.
func (a *adapter) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Convert arguments to the string input format expected by floydtools
	input, err := argsToInput(a.tool.Name(), args)
	if err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Validate input
	if err := a.tool.Validate(input); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// Run the tool and collect streaming output
	ch := make(chan floydtools.StreamMsg, 100)
	doneCh := make(chan struct{})

	// Start the tool in a goroutine
	go a.tool.Run(input)(ch)
	var output strings.Builder
	var toolErr error

	go func() {
		defer close(doneCh)
		for msg := range ch {
			if msg.Chunk != "" {
				output.WriteString(msg.Chunk)
			}
			if msg.Status != "" && !strings.HasPrefix(msg.Status, "completed") &&
				!strings.HasPrefix(msg.Status, "Successfully") &&
				!strings.HasPrefix(msg.Status, "Listed") &&
				!strings.HasPrefix(msg.Status, "Found") &&
				!strings.HasPrefix(msg.Status, "Wrote") &&
				!strings.HasPrefix(msg.Status, "Edited") &&
				!strings.HasPrefix(msg.Status, "Read") &&
				!strings.HasPrefix(msg.Status, "warning") &&
				msg.Status != "success" {
				// Capture error status messages
				if strings.Contains(strings.ToLower(msg.Status), "error") ||
					strings.Contains(strings.ToLower(msg.Status), "failed") {
					toolErr = fmt.Errorf("tool error: %s", msg.Status)
				}
			}
			if msg.Done {
				break
			}
		}
	}()

	// Wait for completion or context cancellation
	select {
	case <-doneCh:
		// Tool finished
	case <-ctx.Done():
		return "", ctx.Err()
	}

	result := output.String()
	if result == "" && toolErr != nil {
		return "", toolErr
	}

	return result, toolErr
}

// Schema returns the tool's MCP schema definition.
func (a *adapter) Schema() Tool {
	return Tool{
		Name:        a.tool.Name(),
		Description: a.tool.Description(),
		InputSchema: schemaForTool(a.tool.Name()),
	}
}

// argsToInput converts MCP arguments to the string format expected by each tool.
func argsToInput(toolName string, args map[string]interface{}) (string, error) {
	switch toolName {
	case "bash":
		cmd, ok := args["command"].(string)
		if !ok {
			return "", fmt.Errorf("command argument required (string)")
		}
		return cmd, nil

	case "read":
		path, ok := args["file_path"].(string)
		if !ok {
			return "", fmt.Errorf("file_path argument required (string)")
		}
		return path, nil

	case "write":
		path, ok := args["file_path"].(string)
		if !ok {
			return "", fmt.Errorf("file_path argument required (string)")
		}
		content, _ := args["content"].(string)
		// Format: path:content
		return fmt.Sprintf("%s:%s", path, content), nil

	case "edit":
		path, ok := args["file_path"].(string)
		if !ok {
			return "", fmt.Errorf("file_path argument required (string)")
		}
		oldStr, ok := args["old_string"].(string)
		if !ok {
			return "", fmt.Errorf("old_string argument required (string)")
		}
		newStr, ok := args["new_string"].(string)
		if !ok {
			return "", fmt.Errorf("new_string argument required (string)")
		}
		// Use :: separator for multiline support
		return fmt.Sprintf("%s::%s::%s", path, oldStr, newStr), nil

	case "multiedit":
		path, ok := args["file_path"].(string)
		if !ok {
			return "", fmt.Errorf("file_path argument required (string)")
		}
		edits, ok := args["edits"].([]interface{})
		if !ok {
			return "", fmt.Errorf("edits argument required (array of {old, new} objects)")
		}
		parts := []string{path}
		for _, edit := range edits {
			editMap, ok := edit.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("each edit must be an object with old and new strings")
			}
			oldStr, ok1 := editMap["old"].(string)
			newStr, ok2 := editMap["new"].(string)
			if !ok1 || !ok2 {
				return "", fmt.Errorf("each edit must have old and new strings")
			}
			parts = append(parts, oldStr, newStr)
		}
		// Use :: separator for multiline support
		return strings.Join(parts, "::"), nil

	case "grep":
		pattern, ok := args["pattern"].(string)
		if !ok {
			return "", fmt.Errorf("pattern argument required (string)")
		}
		path, _ := args["path"].(string)
		if path == "" {
			path = "."
		}
		glob, _ := args["glob"].(string)
		if glob != "" {
			return fmt.Sprintf("%s %s %s", pattern, path, glob), nil
		}
		return fmt.Sprintf("%s %s", pattern, path), nil

	case "ls":
		path, ok := args["path"].(string)
		if !ok || path == "" {
			return ".", nil
		}
		return path, nil

	case "glob":
		pattern, ok := args["pattern"].(string)
		if !ok {
			return "", fmt.Errorf("pattern argument required (string)")
		}
		path, _ := args["path"].(string)
		if path == "" {
			path = "."
		}
		return fmt.Sprintf("%s %s", pattern, path), nil

	case "cache":
		// Cache expects JSON input
		data, err := json.Marshal(args)
		if err != nil {
			return "", fmt.Errorf("failed to encode cache arguments: %w", err)
		}
		return string(data), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

// schemaForTool returns the JSON schema for each tool's input.
func schemaForTool(toolName string) map[string]interface{} {
	switch toolName {
	case "bash":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "Shell command to execute",
				},
			},
			"required": []string{"command"},
		}

	case "read":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the file to read",
				},
			},
			"required": []string{"file_path"},
		}

	case "write":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the file to write",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Content to write to the file",
				},
			},
			"required": []string{"file_path", "content"},
		}

	case "edit":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the file to edit",
				},
				"old_string": map[string]interface{}{
					"type":        "string",
					"description": "Exact string to replace (must be unique in file)",
				},
				"new_string": map[string]interface{}{
					"type":        "string",
					"description": "Replacement string",
				},
			},
			"required": []string{"file_path", "old_string", "new_string"},
		}

	case "multiedit":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the file to edit",
				},
				"edits": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"old": map[string]interface{}{
								"type": "string",
							},
							"new": map[string]interface{}{
								"type": "string",
							},
						},
						"required": []string{"old", "new"},
					},
					"description": "Array of {old, new} edit pairs",
				},
			},
			"required": []string{"file_path", "edits"},
		}

	case "grep":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Regex pattern to search for",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Directory to search in (default: .)",
				},
				"glob": map[string]interface{}{
					"type":        "string",
					"description": "Optional glob pattern to filter files",
				},
			},
			"required": []string{"pattern"},
		}

	case "ls":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Directory path to list (default: .)",
				},
			},
		}

	case "glob":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern (e.g., **/*.go)",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Base path for pattern (default: .)",
				},
			},
			"required": []string{"pattern"},
		}

	case "cache":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"store", "retrieve", "list", "clear", "stats"},
					"description": "Cache operation to perform",
				},
				"tier": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"reasoning", "project", "vault"},
					"description": "Cache tier to operate on",
				},
				"key": map[string]interface{}{
					"type":        "string",
					"description": "Cache key (for store/retrieve)",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "Value to store (for store operation)",
				},
			},
			"required": []string{"operation"},
		}

	default:
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}
}

// NewServer creates a new MCP server with all floydtools registered.
func NewServer() *Server {
	s := &Server{
		tools: make(map[string]ToolHandler),
	}

	// Import floydtools to trigger their init() functions
	// The tools register themselves with floydtools global registry

	// Register all available tools from the floydtools registry
	for _, name := range floydtools.List() {
		if tool, err := floydtools.Get(name); err == nil {
			s.tools[name] = &adapter{tool: tool}
		}
	}

	return s
}

// ListTools returns all available tools.
func (s *Server) ListTools() []Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]Tool, 0, len(s.tools))
	for _, handler := range s.tools {
		tools = append(tools, handler.Schema())
	}
	return tools
}

// GetTool returns a specific tool's schema by name.
func (s *Server) GetTool(name string) (Tool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	handler, ok := s.tools[name]
	if !ok {
		return Tool{}, false
	}
	return handler.Schema(), true
}

// CallTool executes a tool by name.
func (s *Server) CallTool(ctx context.Context, call ToolCall) ToolResult {
	s.mu.RLock()
	handler, ok := s.tools[call.Name]
	s.mu.RUnlock()

	if !ok {
		return ToolResult{
			Error: fmt.Sprintf("tool not found: %s", call.Name),
		}
	}

	// Set a default timeout if no context deadline
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	content, err := handler.Execute(ctx, call.Arguments)
	if err != nil {
		return ToolResult{
			Content: content,
			Error:   err.Error(),
		}
	}

	return ToolResult{
		Content: content,
	}
}

// CallToolSync is a convenience wrapper that creates a context.
func (s *Server) CallToolSync(call ToolCall) ToolResult {
	return s.CallTool(context.Background(), call)
}

// ToolCount returns the number of registered tools.
func (s *Server) ToolCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tools)
}

// HasTool checks if a tool is registered.
func (s *Server) HasTool(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.tools[name]
	return ok
}
