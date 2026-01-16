package agent

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ToolKind represents the type of tool
type ToolKind int

const (
	ToolBash ToolKind = iota
	ToolRead
	ToolWrite
	ToolLS
	ToolGlob
)

// ToolDefinition defines a tool that can be called
type ToolDefinition struct {
	Name        string
	Description string
	Kind        ToolKind
	Parameters  map[string]string
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolID string
	Output string
	Error  error
}

// AsyncToolExecutor executes tools asynchronously with context cancellation
type AsyncToolExecutor struct {
	mu       sync.Mutex
	running  map[string]context.CancelFunc
	timeout  time.Duration
}

// NewAsyncToolExecutor creates a new async tool executor
func NewAsyncToolExecutor() *AsyncToolExecutor {
	return &AsyncToolExecutor{
		running: make(map[string]context.CancelFunc),
		timeout: 2 * time.Minute,
	}
}

// SetTimeout sets the default timeout for tool execution
func (e *AsyncToolExecutor) SetTimeout(d time.Duration) {
	e.timeout = d
}

// ExecBash executes a bash command asynchronously
func (e *AsyncToolExecutor) ExecBash(ctx context.Context, cmd string, outChan chan<- ToolResult) string {
	toolID := "bash_" + generateID()
	subCtx, cancel := context.WithTimeout(ctx, e.timeout)

	e.mu.Lock()
	e.running[toolID] = cancel
	e.mu.Unlock()

	go func() {
		defer func() {
			e.mu.Lock()
			delete(e.running, toolID)
			e.mu.Unlock()
		}()

		result := ToolResult{ToolID: toolID}

		c := exec.CommandContext(subCtx, "bash", "-lc", cmd)
		var output bytes.Buffer
		c.Stdout = &output
		c.Stderr = &output

		err := c.Run()
		result.Output = output.String()

		if subCtx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Errorf("command timed out after %v", e.timeout)
		} else if err != nil {
			result.Error = fmt.Errorf("command failed: %w", err)
		}

		select {
		case outChan <- result:
		case <-ctx.Done():
		}
	}()

	return toolID
}

// ExecRead reads a file asynchronously
func (e *AsyncToolExecutor) ExecRead(ctx context.Context, path string, outChan chan<- ToolResult) string {
	toolID := "read_" + generateID()
	_, cancel := context.WithCancel(ctx)

	e.mu.Lock()
	e.running[toolID] = cancel
	e.mu.Unlock()

	go func() {
		defer func() {
			e.mu.Lock()
			delete(e.running, toolID)
			e.mu.Unlock()
		}()

		result := ToolResult{ToolID: toolID}

		data, err := os.ReadFile(path)
		if err != nil {
			result.Error = err
		} else {
			result.Output = string(data)
		}

		select {
		case outChan <- result:
		case <-ctx.Done():
		}
	}()

	return toolID
}

// ExecWrite writes a file asynchronously
func (e *AsyncToolExecutor) ExecWrite(ctx context.Context, path, content string, outChan chan<- ToolResult) string {
	toolID := "write_" + generateID()
_, cancel := context.WithCancel(ctx)

	e.mu.Lock()
	e.running[toolID] = cancel
	e.mu.Unlock()

	go func() {
		defer func() {
			e.mu.Lock()
			delete(e.running, toolID)
			e.mu.Unlock()
		}()

		result := ToolResult{ToolID: toolID}

		// Ensure parent directory exists
		dir := filepath.Dir(path)
		if dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				result.Error = err
			} else {
				select {
				case outChan <- result:
				case <-ctx.Done():
				}
				return
			}
		}

		err := os.WriteFile(path, []byte(content), fs.FileMode(0o644))
		if err != nil {
			result.Error = err
		} else {
			result.Output = fmt.Sprintf("Wrote %d bytes to %s", len(content), path)
		}

		select {
		case outChan <- result:
		case <-ctx.Done():
		}
	}()

	return toolID
}

// ExecLS lists a directory asynchronously
func (e *AsyncToolExecutor) ExecLS(ctx context.Context, path string, outChan chan<- ToolResult) string {
	toolID := "ls_" + generateID()
_, cancel := context.WithCancel(ctx)

	e.mu.Lock()
	e.running[toolID] = cancel
	e.mu.Unlock()

	go func() {
		defer func() {
			e.mu.Lock()
			delete(e.running, toolID)
			e.mu.Unlock()
		}()

		result := ToolResult{ToolID: toolID}

		if strings.TrimSpace(path) == "" {
			path = "."
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			result.Error = err
		} else {
			var lines []string
			for _, entry := range entries {
				name := entry.Name()
				if entry.IsDir() {
					name += "/"
				}
				lines = append(lines, name)
			}
			result.Output = strings.Join(lines, "\n")
		}

		select {
		case outChan <- result:
		case <-ctx.Done():
		}
	}()

	return toolID
}

// ExecGlob performs glob matching asynchronously
func (e *AsyncToolExecutor) ExecGlob(ctx context.Context, pattern string, outChan chan<- ToolResult) string {
	toolID := "glob_" + generateID()
_, cancel := context.WithCancel(ctx)

	e.mu.Lock()
	e.running[toolID] = cancel
	e.mu.Unlock()

	go func() {
		defer func() {
			e.mu.Lock()
			delete(e.running, toolID)
			e.mu.Unlock()
		}()

		result := ToolResult{ToolID: toolID}

		matches, err := filepath.Glob(pattern)
		if err != nil {
			result.Error = err
		} else {
			result.Output = strings.Join(matches, "\n")
		}

		select {
		case outChan <- result:
		case <-ctx.Done():
		}
	}()

	return toolID
}

// Cancel cancels a running tool by ID
func (e *AsyncToolExecutor) Cancel(toolID string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if cancel, ok := e.running[toolID]; ok {
		cancel()
		delete(e.running, toolID)
		return true
	}
	return false
}

// CancelAll cancels all running tools
func (e *AsyncToolExecutor) CancelAll() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for toolID, cancel := range e.running {
		cancel()
		delete(e.running, toolID)
		_ = toolID
	}
}

// RunningCount returns the number of running tools
func (e *AsyncToolExecutor) RunningCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.running)
}

// GetDefaultToolDefinitions returns the default tool definitions
func GetDefaultToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "Bash",
			Description: "Run shell commands (bash -lc) and capture stdout/stderr.",
			Kind:        ToolBash,
			Parameters: map[string]string{
				"command": "The shell command to execute",
			},
		},
		{
			Name:        "Read",
			Description: "Read a file from disk and display its contents.",
			Kind:        ToolRead,
			Parameters: map[string]string{
				"path": "Absolute or relative path to the file",
			},
		},
		{
			Name:        "Write",
			Description: "Create or overwrite a file with content.",
			Kind:        ToolWrite,
			Parameters: map[string]string{
				"path":    "Path to the file",
				"content": "Content to write",
			},
		},
		{
			Name:        "LS",
			Description: "List directory entries.",
			Kind:        ToolLS,
			Parameters: map[string]string{
				"path": "Directory path (default: current directory)",
			},
		},
		{
			Name:        "Glob",
			Description: "Match files with Go-style patterns.",
			Kind:        ToolGlob,
			Parameters: map[string]string{
				"pattern": "Glob pattern (e.g., **/*.go)",
			},
		},
	}
}

// generateID creates a unique short ID
func generateID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}
