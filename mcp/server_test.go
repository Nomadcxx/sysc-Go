package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	s := NewServer()
	if s == nil {
		t.Fatal("NewServer returned nil")
	}

	// Should have at least 9 tools registered
	count := s.ToolCount()
	if count < 9 {
		t.Errorf("Expected at least 9 tools, got %d", count)
	}

	// Check for expected tools
	expectedTools := []string{"bash", "read", "write", "edit", "multiedit", "grep", "ls", "glob", "cache"}
	for _, name := range expectedTools {
		if !s.HasTool(name) {
			t.Errorf("Expected tool %s not found", name)
		}
	}
}

func TestListTools(t *testing.T) {
	s := NewServer()
	tools := s.ListTools()

	if len(tools) < 9 {
		t.Errorf("Expected at least 9 tools, got %d", len(tools))
	}

	// Verify each tool has required fields
	for _, tool := range tools {
		if tool.Name == "" {
			t.Error("Tool has empty name")
		}
		if tool.Description == "" {
			t.Errorf("Tool %s has empty description", tool.Name)
		}
		if tool.InputSchema == nil {
			t.Errorf("Tool %s has nil schema", tool.Name)
		}
	}
}

func TestGetTool(t *testing.T) {
	s := NewServer()

	tool, ok := s.GetTool("bash")
	if !ok {
		t.Fatal("bash tool not found")
	}
	if tool.Name != "bash" {
		t.Errorf("Expected tool name 'bash', got '%s'", tool.Name)
	}

	// Check schema has required properties
	props, ok := tool.InputSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("bash schema properties not a map")
	}
	if _, ok := props["command"]; !ok {
		t.Error("bash schema missing 'command' property")
	}

	// Test non-existent tool
	_, ok = s.GetTool("nonexistent")
	if ok {
		t.Error("Expected false for non-existent tool")
	}
}

func TestCallToolBash(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	call := ToolCall{
		Name: "bash",
		Arguments: map[string]interface{}{
			"command": "echo hello",
		},
	}

	result := s.CallTool(ctx, call)
	if result.Error != "" {
		t.Errorf("Unexpected error: %s", result.Error)
	}
	if !strings.Contains(result.Content, "hello") {
		t.Errorf("Expected output to contain 'hello', got: %s", result.Content)
	}
}

func TestCallToolBashTimeout(t *testing.T) {
	s := NewServer()

	// Create a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	call := ToolCall{
		Name: "bash",
		Arguments: map[string]interface{}{
			"command": "sleep 10", // This should timeout
		},
	}

	result := s.CallTool(ctx, call)
	if result.Error == "" {
		t.Error("Expected timeout error")
	}
}

func TestCallToolReadWrite(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	// Create temp directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, MCP!"

	// Write
	writeCall := ToolCall{
		Name: "write",
		Arguments: map[string]interface{}{
			"file_path": testFile,
			"content":   testContent,
		},
	}

	writeResult := s.CallTool(ctx, writeCall)
	if writeResult.Error != "" {
		t.Fatalf("Write failed: %s", writeResult.Error)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("File was not created")
	}

	// Read
	readCall := ToolCall{
		Name: "read",
		Arguments: map[string]interface{}{
			"file_path": testFile,
		},
	}

	readResult := s.CallTool(ctx, readCall)
	if readResult.Error != "" {
		t.Fatalf("Read failed: %s", readResult.Error)
	}

	if !strings.Contains(readResult.Content, testContent) {
		t.Errorf("Expected content '%s', got: %s", testContent, readResult.Content)
	}
}

func TestCallToolEdit(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "edit_test.txt")
	initialContent := "foo bar baz"
	updatedContent := "foo qux baz"

	// Write initial content
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Edit: replace "bar" with "qux"
	editCall := ToolCall{
		Name: "edit",
		Arguments: map[string]interface{}{
			"file_path":  testFile,
			"old_string": "bar",
			"new_string": "qux",
		},
	}

	result := s.CallTool(ctx, editCall)
	if result.Error != "" {
		t.Fatalf("Edit failed: %s", result.Error)
	}

	// Verify edit
	data, _ := os.ReadFile(testFile)
	if string(data) != updatedContent {
		t.Errorf("Expected '%s', got '%s'", updatedContent, string(data))
	}
}

func TestCallToolMultiEdit(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "multiedit_test.txt")
	initialContent := "alpha beta gamma delta"
	updatedContent := "one two three delta"

	// Write initial content
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	// MultiEdit: replace multiple strings
	multiEditCall := ToolCall{
		Name: "multiedit",
		Arguments: map[string]interface{}{
			"file_path": testFile,
			"edits": []interface{}{
				map[string]interface{}{"old": "alpha", "new": "one"},
				map[string]interface{}{"old": "beta", "new": "two"},
				map[string]interface{}{"old": "gamma", "new": "three"},
			},
		},
	}

	result := s.CallTool(ctx, multiEditCall)
	if result.Error != "" {
		t.Fatalf("MultiEdit failed: %s", result.Error)
	}

	// Verify edits
	data, _ := os.ReadFile(testFile)
	if string(data) != updatedContent {
		t.Errorf("Expected '%s', got '%s'", updatedContent, string(data))
	}
}

func TestCallToolGrep(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "grep_test.txt")
	if err := os.WriteFile(testFile, []byte("hello\nworld\nhello again\n"), 0644); err != nil {
		t.Fatal(err)
	}

	call := ToolCall{
		Name: "grep",
		Arguments: map[string]interface{}{
			"pattern": "hello",
			"path":    tmpDir,
		},
	}

	result := s.CallTool(ctx, call)
	if result.Error != "" {
		t.Fatalf("Grep failed: %s", result.Error)
	}

	// Should find "hello" twice
	helloCount := strings.Count(result.Content, "hello")
	if helloCount < 2 {
		t.Errorf("Expected to find 'hello' at least twice, got %d matches", helloCount)
	}
}

func TestCallToolLs(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tmpDir := t.TempDir()
	// Create some test files
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content"), 0644)

	call := ToolCall{
		Name: "ls",
		Arguments: map[string]interface{}{
			"path": tmpDir,
		},
	}

	result := s.CallTool(ctx, call)
	if result.Error != "" {
		t.Fatalf("Ls failed: %s", result.Error)
	}

	if !strings.Contains(result.Content, "file1.txt") {
		t.Errorf("Expected output to contain 'file1.txt', got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "file2.txt") {
		t.Errorf("Expected output to contain 'file2.txt', got: %s", result.Content)
	}
}

func TestCallToolGlob(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tmpDir := t.TempDir()
	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "test1.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "test2.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("readme"), 0644)

	call := ToolCall{
		Name: "glob",
		Arguments: map[string]interface{}{
			"pattern": "*.go",
			"path":    tmpDir,
		},
	}

	result := s.CallTool(ctx, call)
	if result.Error != "" {
		t.Fatalf("Glob failed: %s", result.Error)
	}

	if !strings.Contains(result.Content, "test1.go") {
		t.Errorf("Expected output to contain 'test1.go', got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "test2.go") {
		t.Errorf("Expected output to contain 'test2.go', got: %s", result.Content)
	}
	if strings.Contains(result.Content, "readme.txt") {
		t.Errorf("Expected output to NOT contain 'readme.txt', got: %s", result.Content)
	}
}

func TestCallToolNotFound(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	call := ToolCall{
		Name:      "nonexistent_tool",
		Arguments: map[string]interface{}{},
	}

	result := s.CallTool(ctx, call)
	if result.Error == "" {
		t.Error("Expected error for non-existent tool")
	}
	if !strings.Contains(result.Error, "tool not found") {
		t.Errorf("Expected 'tool not found' error, got: %s", result.Error)
	}
}

func TestCallToolSync(t *testing.T) {
	s := NewServer()

	call := ToolCall{
		Name: "bash",
		Arguments: map[string]interface{}{
			"command": "echo sync test",
		},
	}

	result := s.CallToolSync(call)
	if result.Error != "" {
		t.Errorf("Unexpected error: %s", result.Error)
	}
	if !strings.Contains(result.Content, "sync test") {
		t.Errorf("Expected output to contain 'sync test', got: %s", result.Content)
	}
}

func TestSchemaValidation(t *testing.T) {
	s := NewServer()
	tools := s.ListTools()

	// Verify all tools have valid schemas
	for _, tool := range tools {
		schema := tool.InputSchema

		// Must be an object type
		typ, ok := schema["type"].(string)
		if !ok || typ != "object" {
			t.Errorf("Tool %s: schema type must be 'object', got %v", tool.Name, typ)
		}

		// Must have properties
		props, ok := schema["properties"].(map[string]interface{})
		if !ok {
			t.Errorf("Tool %s: schema must have properties", tool.Name)
			continue
		}

		// Required fields should exist in properties
		required, _ := schema["required"].([]string)
		for _, req := range required {
			if _, ok := props[req]; !ok {
				t.Errorf("Tool %s: required field '%s' not in properties", tool.Name, req)
			}
		}
	}
}

func BenchmarkCallToolBash(b *testing.B) {
	s := NewServer()
	ctx := context.Background()
	call := ToolCall{
		Name: "bash",
		Arguments: map[string]interface{}{
			"command": "echo test",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.CallTool(ctx, call)
	}
}
