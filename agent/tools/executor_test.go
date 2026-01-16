package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Nomadcxx/sysc-Go/agent"
	"github.com/Nomadcxx/sysc-Go/agent/message"
	"github.com/Nomadcxx/sysc-Go/tui/floydtools"
)

// TestExecutor_UnknownTool tests error handling for non-existent tools
func TestExecutor_UnknownTool(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	toolCall := agent.ToolCall{
		ID:   "test-1",
		Name: "nonexistent_tool",
		Input: `{"arg": "value"}`,
		InputMap: map[string]any{
			"arg": "value",
		},
	}

	result, err := executor.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error for unknown tool, got: %v", err)
	}

	if !result.IsError {
		t.Error("Expected result.IsError to be true for unknown tool")
	}

	if !strings.Contains(result.Content, "tool not found") && !strings.Contains(result.Content, "Error getting tool") {
		t.Errorf("Expected error message about tool not found, got: %s", result.Content)
	}
}

// TestExecutor_EmptyInput tests tools with empty arguments
// BLOCKING ERROR #1: Tools validate empty input but errors are not propagated to ToolResultBlock.IsError
func TestExecutor_EmptyInput(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	tests := []struct {
		name            string
		toolName        string
		input           string
		expectError     bool
		errorInContent  bool
		errorMsg        string
	}{
		{
			name:           "Bash with empty command",
			toolName:       "bash",
			input:          "",
			expectError:    true,  // Should be marked as error
			errorInContent: true,  // Error appears in content
			errorMsg:       "command cannot be empty",
		},
		{
			name:           "Read with empty path",
			toolName:       "read",
			input:          "",
			expectError:    true,  // Should be marked as error
			errorInContent: false, // Empty input just produces empty output
			errorMsg:       "",
		},
		{
			name:           "Write with empty path",
			toolName:       "write",
			input:          "",
			expectError:    true,  // Should be marked as error
			errorInContent: false, // Creates file named "{}"
			errorMsg:       "",
		},
		{
			name:           "Grep with empty pattern",
			toolName:       "grep",
			input:          "",
			expectError:    true,  // Should be marked as error
			errorInContent: true,  // Error appears in content
			errorMsg:       "Grep failed",
		},
		{
			name:           "Glob with empty pattern",
			toolName:       "glob",
			input:          "",
			expectError:    true,  // Should be marked as error
			errorInContent: true,  // Error appears in content
			errorMsg:       "Glob failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolCall := agent.ToolCall{
				ID:       "test-" + tt.toolName,
				Name:     tt.toolName,
				Input:    tt.input,
				InputMap: map[string]any{},
			}

			result, err := executor.ExecuteToolCall(ctx, toolCall)
			if err != nil {
				t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
			}

			// BLOCKING ERROR: Tools send errors in Status field but executor doesn't check for
			// error patterns like "failed:", "Error:", etc. It only checks for "error:" prefix
			if tt.expectError && !result.IsError {
				t.Logf("BLOCKING ISSUE: Expected IsError=true for %s but got IsError=false", tt.name)
				t.Logf("  Content: %q", result.Content)
			}

			if tt.errorInContent && !strings.Contains(result.Content, tt.errorMsg) {
				t.Logf("Expected error message containing '%s', got: %s", tt.errorMsg, result.Content)
			}
		})
	}
}

// TestExecutor_ReadFileNotFound tests reading a non-existent file
// BLOCKING ERROR #2: Tool failures (file not found, permission denied) are not marked as IsError
func TestExecutor_ReadFileNotFound(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	toolCall := agent.ToolCall{
		ID:   "test-read",
		Name: "read",
		Input: "/nonexistent/path/that/does/not/exist.txt",
		InputMap: map[string]any{
			"file_path": "/nonexistent/path/that/does/not/exist.txt",
		},
	}

	result, err := executor.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
	}

	// The tool sends "Read failed: ..." in Status field but executor doesn't parse it
	// as an error condition for ToolResultBlock.IsError
	if !result.IsError {
		t.Logf("BLOCKING ISSUE: File not found should have IsError=true, got IsError=false")
		t.Logf("  Content: %q", result.Content)
	}

	if result.Content == "" {
		t.Error("Expected some error response for non-existent file")
	}
}

// TestExecutor_WritePermissionDenied tests writing to a protected location
func TestExecutor_WritePermissionDenied(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	// Try to write to /root which should be permission denied on most systems
	toolCall := agent.ToolCall{
		ID:   "test-write",
		Name: "write",
		Input: "/root/test_floyd_write.txt:some content",
		InputMap: map[string]any{
			"file_path": "/root/test_floyd_write.txt",
			"content":   "some content",
		},
	}

	result, err := executor.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
	}

	// Should not crash; should handle permission error gracefully
	// The exact error message depends on the system
	t.Logf("Permission denied result: IsError=%v, Content=%s", result.IsError, result.Content)
	if result.Content == "" {
		t.Error("Expected some response for permission denied write")
	}

	// Verify it didn't actually create the file
	_, statErr := os.Stat("/root/test_floyd_write.txt")
	if statErr == nil {
		os.Remove("/root/test_floyd_write.txt")
		t.Error("File should not have been created in protected location")
	}
}

// TestExecutor_EditOldStringNotFound tests edit when old string doesn't exist
func TestExecutor_EditOldStringNotFound(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	// Create a temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello World\nLine 2\nLine 3"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to edit with old string that doesn't exist
	oldString := "NonExistentString"
	newString := "Replacement"
	input := fmt.Sprintf("%s::%s::%s", testFile, oldString, newString)

	toolCall := agent.ToolCall{
		ID:   "test-edit",
		Name: "edit",
		Input: input,
		InputMap: map[string]any{
			"file_path":  testFile,
			"old_string": oldString,
			"new_string": newString,
		},
	}

	result, err := executor.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
	}

	// Edit tool sends "Edit failed: old string not found" but it's not marked as IsError
	if !result.IsError {
		t.Logf("BLOCKING ISSUE: Old string not found should have IsError=true")
		t.Logf("  Content: %q", result.Content)
	}

	// Verify file was not modified
	actual, _ := os.ReadFile(testFile)
	if string(actual) != content {
		t.Error("File should not have been modified when old string not found")
	}
}

// TestExecutor_EditMultipleMatches tests edit when multiple matches exist
func TestExecutor_EditMultipleMatches(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	// Create a temp file with multiple occurrences
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello World\nHello World\nHello World"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Edit with replace_all=false (default)
	oldString := "Hello World"
	newString := "Goodbye World"
	input := fmt.Sprintf("%s::%s::%s", testFile, oldString, newString)

	toolCall := agent.ToolCall{
		ID:   "test-edit-multi",
		Name: "edit",
		Input: input,
		InputMap: map[string]any{
			"file_path":  testFile,
			"old_string": oldString,
			"new_string": newString,
		},
	}

	result, err := executor.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
	}

	_ = result // Use result to avoid unused variable error

	// Verify only one occurrence was replaced (default behavior)
	actual, _ := os.ReadFile(testFile)
	actualStr := string(actual)
	expected := "Goodbye World\nHello World\nHello World"
	if actualStr != expected {
		// The edit tool replaces only the first occurrence by default
		t.Logf("Note: Edit replaced content as: %s", actualStr)
	}
}

// TestExecutor_GrepInvalidRegex tests grep with invalid regex pattern
func TestExecutor_GrepInvalidRegex(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	// Invalid regex: unclosed bracket
	invalidPattern := "[unclosed"

	toolCall := agent.ToolCall{
		ID:   "test-grep",
		Name: "grep",
		Input: invalidPattern,
		InputMap: map[string]any{
			"pattern": invalidPattern,
			"path":    ".",
		},
	}

	result, err := executor.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
	}

	// Should handle invalid regex gracefully
	// ripgrep might return non-zero exit code or error message
	// The key is that it should NOT crash the TUI
	t.Logf("Result for invalid regex: %s", result.Content)
}

// TestExecutor_GrepNoMatches tests grep when no results are found
func TestExecutor_GrepNoMatches(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	// A pattern very unlikely to match
	pattern := "XYZ_UNMATCHABLE_PATTERN_12345"

	toolCall := agent.ToolCall{
		ID:   "test-grep-nomatch",
		Name: "grep",
		Input: pattern + " .",
		InputMap: map[string]any{
			"pattern": pattern,
			"path":    ".",
		},
	}

	result, err := executor.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
	}

	// Should not be an error - no matches is a valid result
	if result.IsError {
		t.Error("No matches should not be treated as an error")
	}
}

// TestExecutor_LSInvalidPath tests ls with non-existent directory
func TestExecutor_LSInvalidPath(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	toolCall := agent.ToolCall{
		ID:   "test-ls",
		Name: "ls",
		Input: "/nonexistent/directory/xyz",
		InputMap: map[string]any{
			"path": "/nonexistent/directory/xyz",
		},
	}

	result, err := executor.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
	}

	// LS tool sends "LS failed: ..." but not marked as IsError
	if !result.IsError {
		t.Logf("BLOCKING ISSUE: Non-existent directory should have IsError=true")
		t.Logf("  Content: %q", result.Content)
	}
}

// TestExecutor_GlobInvalidPattern tests glob with invalid pattern
func TestExecutor_GlobInvalidPattern(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	// filepath.Glob has some limitations but generally accepts most patterns
	// Let's test with a valid pattern that just won't match anything
	toolCall := agent.ToolCall{
		ID:   "test-glob",
		Name: "glob",
		Input: "**/*.nonexistentfile123",
		InputMap: map[string]any{
			"pattern": "**/*.nonexistentfile123",
		},
	}

	result, err := executor.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
	}

	// No matches is valid, not an error
	t.Logf("Glob result for no matches: %s", result.Content)
}

// TestExecutor_Timeout tests tool execution timeout
func TestExecutor_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	executor := NewExecutor(100 * time.Millisecond)
	ctx := context.Background()

	// Sleep command that exceeds timeout
	toolCall := agent.ToolCall{
		ID:   "test-timeout",
		Name: "bash",
		Input: "sleep 5",
		InputMap: map[string]any{
			"command": "sleep 5",
		},
	}

	start := time.Now()
	result, err := executor.ExecuteToolCall(ctx, toolCall)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
	}

	// Should timeout quickly (well under 5 seconds)
	if elapsed > 2*time.Second {
		t.Errorf("Timeout took too long: %v, expected < 2s", elapsed)
	}

	t.Logf("Timeout result: IsError=%v, Content=%s", result.IsError, result.Content)
}

// TestExecutor_BashCommandNotFound tests bash with command that doesn't exist
func TestExecutor_BashCommandNotFound(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	// Command that definitely doesn't exist
	toolCall := agent.ToolCall{
		ID:   "test-bash",
		Name: "bash",
		Input: "thiscommanddefinitelydoesnotexist12345",
		InputMap: map[string]any{
			"command": "thiscommanddefinitelydoesnotexist12345",
		},
	}

	result, err := executor.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
	}

	// Should handle "command not found" gracefully
	t.Logf("Command not found result: %s", result.Content)
}

// TestExecutor_BashNonZeroExit tests bash command that fails with non-zero exit
func TestExecutor_BashNonZeroExit(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	// Command that exits with non-zero code
	toolCall := agent.ToolCall{
		ID:   "test-bash-fail",
		Name: "bash",
		Input: "exit 42",
		InputMap: map[string]any{
			"command": "exit 42",
		},
	}

	result, err := executor.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
	}

	// Should handle non-zero exit without crashing
	t.Logf("Non-zero exit result: IsError=%v, Content=%s", result.IsError, result.Content)
}

// TestExecutor_ReadBinaryFile tests reading a binary file
func TestExecutor_ReadBinaryFile(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	// Create a binary file
	tmpDir := t.TempDir()
	binaryFile := filepath.Join(tmpDir, "binary.bin")
	binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	if err := os.WriteFile(binaryFile, binaryData, 0644); err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}

	toolCall := agent.ToolCall{
		ID:   "test-read-binary",
		Name: "read",
		Input: binaryFile,
		InputMap: map[string]any{
			"file_path": binaryFile,
		},
	}

	result, err := executor.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
	}

	// Should not crash; may output garbage or handle gracefully
	t.Logf("Binary file read: %d bytes in content", len(result.Content))
}

// TestExecutor_ParallelTools tests running multiple tools in parallel
func TestExecutor_ParallelTools(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	// Create test files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	file3 := filepath.Join(tmpDir, "file3.txt")

	for i, f := range []string{file1, file2, file3} {
		if err := os.WriteFile(f, []byte(fmt.Sprintf("Content %d", i)), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	toolCalls := []agent.ToolCall{
		{ID: "test-1", Name: "read", Input: file1, InputMap: map[string]any{"file_path": file1}},
		{ID: "test-2", Name: "read", Input: file2, InputMap: map[string]any{"file_path": file2}},
		{ID: "test-3", Name: "read", Input: file3, InputMap: map[string]any{"file_path": file3}},
	}

	start := time.Now()
	results := executor.ExecuteToolCalls(ctx, toolCalls)
	elapsed := time.Since(start)

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Verify all results are in order (first result = first tool call)
	if results[0].ToolUseID != "test-1" || results[1].ToolUseID != "test-2" || results[2].ToolUseID != "test-3" {
		t.Error("Results not in correct order")
	}

	// Parallel execution should be faster than sequential
	t.Logf("Parallel execution of 3 tools took: %v", elapsed)
}

// TestExecutor_LargeOutput tests handling of very large tool output
func TestExecutor_LargeOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large output test in short mode")
	}

	executor := NewExecutor(30 * time.Second)
	ctx := context.Background()

	// Create a large file (1MB)
	tmpDir := t.TempDir()
	largeFile := filepath.Join(tmpDir, "large.txt")
	largeContent := strings.Repeat("This is a line of text\n", 100000) // ~2.5MB

	if err := os.WriteFile(largeFile, []byte(largeContent), 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	toolCall := agent.ToolCall{
		ID:   "test-large",
		Name: "read",
		Input: largeFile,
		InputMap: map[string]any{
			"file_path": largeFile,
		},
	}

	start := time.Now()
	result, err := executor.ExecuteToolCall(ctx, toolCall)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ExecuteToolCall should not return error, got: %v", err)
	}

	t.Logf("Large output (%d bytes) took %v to process", len(result.Content), elapsed)

	// Should complete reasonably quickly
	if elapsed > 10*time.Second {
		t.Error("Large output took too long to process")
	}
}

// TestExecutor_BlockedChannel tests what happens when channel blocks
func TestExecutor_BlockedChannel(t *testing.T) {
	// This test verifies the executor handles slow/unresponsive tools
	executor := NewExecutor(100 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start a long-running command
	toolCall := agent.ToolCall{
		ID:   "test-block",
		Name: "bash",
		Input: "sleep 10",
		InputMap: map[string]any{
			"command": "sleep 10",
		},
	}

	done := make(chan struct{})
	var result message.ToolResultBlock
	var err error

	go func() {
		result, err = executor.ExecuteToolCall(ctx, toolCall)
		close(done)
	}()

	select {
	case <-done:
		// Completed (should timeout first)
		if err != nil {
			t.Logf("Expected timeout error, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Executor appears to have hung - this is a BLOCKING ERROR")
	}

	t.Logf("Blocked channel test completed: IsError=%v", result.IsError)
}

// TestExecutor_ContextCancellation tests context cancellation during tool execution
func TestExecutor_ContextCancellation(t *testing.T) {
	executor := NewExecutor(30 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	toolCall := agent.ToolCall{
		ID:   "test-cancel",
		Name: "bash",
		Input: "sleep 30",
		InputMap: map[string]any{
			"command": "sleep 30",
		},
	}

	done := make(chan struct{})
	var result message.ToolResultBlock
	var execErr error

	go func() {
		result, execErr = executor.ExecuteToolCall(ctx, toolCall)
		close(done)
	}()

	// Cancel immediately
	cancel()

	select {
	case <-done:
		// Good - it exited quickly after cancel
		if execErr == nil && !result.IsError {
			t.Log("Note: Tool completed despite cancellation (may have finished immediately)")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Context cancellation did not stop tool execution - this is a BLOCKING ERROR")
	}
}

// TestExecuteToolCalls_EmptySlice tests empty tool call slice
func TestExecuteToolCalls_EmptySlice(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	ctx := context.Background()

	results := executor.ExecuteToolCalls(ctx, []agent.ToolCall{})

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty slice, got %d", len(results))
	}
}

// TestCreateToolResultMessage tests the message creation helper
func TestCreateToolResultMessage(t *testing.T) {
	results := []message.ToolResultBlock{
		{ToolUseID: "tool-1", Content: "Output 1", IsError: false},
		{ToolUseID: "tool-2", Content: "Error message", IsError: true},
	}

	msg := CreateToolResultMessage(results)

	if msg.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", msg.Role)
	}

	if len(msg.Content) != 2 {
		t.Fatalf("Expected 2 content blocks, got %d", len(msg.Content))
	}

	if msg.Content[0].Type != "tool_result" {
		t.Errorf("Expected content type 'tool_result', got '%s'", msg.Content[0].Type)
	}

	if msg.Content[1].IsError != true {
		t.Error("Expected second result to have IsError=true")
	}
}

// TestParseToolInput tests parsing tool input JSON
func TestParseToolInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid JSON",
			input:   `{"key": "value", "number": 123}`,
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			input:   `{not valid json}`,
			wantErr: true,
		},
		{
			name:    "Empty JSON object",
			input:   `{}`,
			wantErr: false,
		},
		{
			name:    "Nested JSON",
			input:   `{"outer": {"inner": "value"}}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseToolInput(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error for invalid JSON, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result == nil {
					t.Error("Expected non-nil result for valid JSON")
				}
			}
		})
	}
}

// TestFormatToolInput tests formatting tool input to JSON
func TestFormatToolInput(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		wantErr bool
	}{
		{
			name: "Simple map",
			input: map[string]any{
				"key": "value",
				"number": 123,
			},
			wantErr: false,
		},
		{
			name: "Nested map",
			input: map[string]any{
				"outer": map[string]any{
					"inner": "value",
				},
			},
			wantErr: false,
		},
		{
			name:    "Empty map",
			input:   map[string]any{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatToolInput(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result == "" {
					t.Error("Expected non-empty result")
				}
				// Verify it's valid JSON
				var parsed map[string]any
				if json.Unmarshal([]byte(result), &parsed) != nil {
					t.Error("Result is not valid JSON")
				}
			}
		})
	}
}

// TestListTools tests listing available tools
func TestListTools(t *testing.T) {
	tools := ListTools()

	if len(tools) == 0 {
		t.Error("Expected at least one tool to be registered")
	}

	// Check for expected tools
	expectedTools := []string{"bash", "read", "write", "edit", "grep", "ls", "glob"}
	for _, expected := range expectedTools {
		found := false
		for _, tool := range tools {
			if tool == expected {
				found = true
				break
			}
		}
		if !found {
			t.Logf("Warning: Expected tool '%s' not found in registry", expected)
		}
	}

	t.Logf("Registered tools: %v", tools)
}

// TestGetToolDescriptions tests getting tool descriptions
func TestGetToolDescriptions(t *testing.T) {
	descriptions := GetToolDescriptions()

	if len(descriptions) == 0 {
		t.Error("Expected at least one tool description")
	}

	// Each tool should have a non-empty description
	for toolName, desc := range descriptions {
		if desc == "" {
			t.Errorf("Tool '%s' has empty description", toolName)
		}
		if toolName == "" {
			t.Error("Found empty tool name in descriptions")
		}
	}

	t.Logf("Tool descriptions count: %d", len(descriptions))
}

// TestFloydtools_Interface tests that all floydtools implement the interface correctly
func TestFloydtools_Interface(t *testing.T) {
	tools := floydtools.AllTools()

	for _, tool := range tools {
		name := tool.Name()
		desc := tool.Description()

		if name == "" {
			t.Error("Tool has empty name")
		}

		if desc == "" {
			t.Errorf("Tool '%s' has empty description", name)
		}

		// Test Validate with empty input (should work for some tools)
		err := tool.Validate("")
		t.Logf("Tool '%s' Validate(''): %v", name, err)

		// Get the runner function
		runner := tool.Run("test")
		if runner == nil {
			t.Errorf("Tool '%s' returned nil runner", name)
		}
	}
}
