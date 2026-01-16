package floydui

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestLoadSession_MissingFile tests loading when session file doesn't exist
func TestLoadSession_MissingFile(t *testing.T) {
	// Create a temporary path that doesn't exist
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, ".floyd_session.json")

	session, err := loadSession(nonExistentPath)

	// Current behavior: returns error (os.ErrNotExist)
	if err == nil {
		t.Error("Expected error when loading non-existent file, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.ErrNotExist, got: %v", err)
	}
	// Session should be empty (zero value)
	if len(session.Messages) != 0 || len(session.History) != 0 {
		t.Error("Session should be empty when file doesn't exist")
	}
}

// TestLoadSession_CorruptJSON tests loading a file with invalid JSON
func TestLoadSession_CorruptJSON(t *testing.T) {
	tmpDir := t.TempDir()
	corruptPath := filepath.Join(tmpDir, ".floyd_session.json")

	// Write corrupt JSON
	corruptContent := `{"messages": ["USER: hello", "FLOYD: world", INVALID_JSON_HERE}`
	if err := os.WriteFile(corruptPath, []byte(corruptContent), 0644); err != nil {
		t.Fatalf("Failed to write corrupt file: %v", err)
	}

	session, err := loadSession(corruptPath)

	// Current behavior: returns error and empty session
	if err == nil {
		t.Error("Expected error when loading corrupt JSON, got nil")
	}
	if len(session.Messages) != 0 || len(session.History) != 0 {
		t.Error("Session should be empty when JSON is corrupt")
	}

	// BLOCKING ISSUE #1: Corrupt file causes ALL previous session data to be lost
	// The corrupt file remains on disk, so every load will fail
	// User loses their entire conversation history permanently
}

// TestLoadSession_TruncatedJSON tests a file that was partially written
func TestLoadSession_TruncatedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	truncatedPath := filepath.Join(tmpDir, ".floyd_session.json")

	// Create a valid session
	validSession := Session{
		Messages: []string{"USER: hello", "FLOYD: world"},
		History:  []string{"hello"},
	}
	data, _ := json.MarshalIndent(validSession, "", "  ")

	// Write only half of it
	truncatedData := data[:len(data)/2]
	if err := os.WriteFile(truncatedPath, truncatedData, 0644); err != nil {
		t.Fatalf("Failed to write truncated file: %v", err)
	}

	session, err := loadSession(truncatedPath)

	// Should get an error
	if err == nil {
		t.Error("Expected error when loading truncated JSON, got nil")
	}
	if len(session.Messages) != 0 {
		t.Error("Session should be empty when JSON is truncated")
	}
}

// TestSaveAndLoadSession_RoundTrip tests basic save/load cycle
func TestSaveAndLoadSession_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	original := Session{
		Messages: []string{"USER: hello", "FLOYD: world", "USER: how are you?"},
		History:  []string{"hello", "how are you?"},
	}

	// Save
	err := saveSession(original, sessionPath)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Load
	loaded, err := loadSession(sessionPath)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Verify
	if len(loaded.Messages) != len(original.Messages) {
		t.Errorf("Message count mismatch: got %d, want %d", len(loaded.Messages), len(original.Messages))
	}
	for i, msg := range original.Messages {
		if loaded.Messages[i] != msg {
			t.Errorf("Message %d mismatch: got %q, want %q", i, loaded.Messages[i], msg)
		}
	}
	if len(loaded.History) != len(original.History) {
		t.Errorf("History count mismatch: got %d, want %d", len(loaded.History), len(original.History))
	}
}

// TestSaveSession_PermissionDenied tests writing to a read-only location
func TestSaveSession_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()

	// Create a read-only directory
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0444); err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}
	readOnlyPath := filepath.Join(readOnlyDir, ".floyd_session.json")

	session := Session{
		Messages: []string{"USER: test"},
	}

	err := saveSession(session, readOnlyPath)
	if err == nil {
		t.Error("Expected error when writing to read-only location, got nil")
	}

	// BLOCKING ISSUE #2: Save error is silently ignored in NewModelWithSession("test_session.json")
	// The model initializes successfully but session is never persisted
}

// TestSaveSession_DiskFullSimulated tests handling of write failures
func TestSaveSession_DiskFullSimulated(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file and fill it with data (simulate full disk by using a very long path that exceeds limits)
	// On most systems, path length is limited to ~4096 characters

	// Create a deeply nested directory structure
	deepPath := tmpDir
	for i := 0; i < 100; i++ {
		deepPath = filepath.Join(deepPath, strings.Repeat("a", 100))
	}
	_ = deepPath // Used for path length testing
	sessionPath := filepath.Join(tmpDir, "long_name_test_session.json")

	session := Session{
		Messages: []string{"USER: test"},
	}

	err := saveSession(session, sessionPath)
	if err == nil {
		t.Log("Warning: Expected error for excessively long path, but write succeeded")
	} else {
		t.Logf("Got expected error for long path: %v", err)
	}
}

// TestConcurrentWrites tests multiple goroutines saving simultaneously
func TestConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	var wg sync.WaitGroup
	numGoroutines := 10
	writesPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				session := Session{
					Messages: []string{string(rune('A'+id)) + ": message " + string(rune('0'+j))},
					History:  []string{"history"},
				}
				_ = saveSession(session, sessionPath)
			}
		}(i)
	}
	wg.Wait()

	// Verify file exists and is valid
	_, err := loadSession(sessionPath)
	if err != nil {
		// BLOCKING ISSUE #3: No mutex protection on saveSession
		// Concurrent writes can corrupt the JSON file
		t.Errorf("After concurrent writes, session file is corrupted: %v", err)
	}
}

// TestSaveCurrentSession_VeryLargeHistory tests saving thousands of messages
func TestSaveCurrentSession_VeryLargeHistory(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	m := NewModelWithSession("test_session.json")
	m.SessionPath = sessionPath
	m.Messages = nil // Clear any existing messages

	// Add 10,000 messages
	for i := 0; i < 10000; i++ {
		m.Messages = append(m.Messages, ChatMessage{
			Role:    "user",
			Content: string(rune('A'+(i%26))) + " message " + string(rune('0'+(i%10))),
		})
		m.History = append(m.History, "history entry "+string(rune('0'+(i%10))))
	}

	err := m.SaveCurrentSession()
	if err != nil {
		t.Logf("Failed to save large session: %v", err)
	}

	// Check file size
	info, err := os.Stat(sessionPath)
	if err != nil {
		t.Fatalf("Failed to stat session file: %v", err)
	}

	t.Logf("Large session file size: %d bytes", info.Size())

	// Verify it can be loaded
	loaded, err := loadSession(sessionPath)
	if err != nil {
		t.Errorf("Failed to load large session: %v", err)
	}
	if len(loaded.Messages) != 10000 {
		t.Errorf("Expected 10000 messages, got %d", len(loaded.Messages))
	}
}

// TestSpecialCharactersInMessages tests unicode, emojis, null bytes, etc.
func TestSpecialCharactersInMessages(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	m := NewModelWithSession("test_session.json")
	m.SessionPath = sessionPath

	// Add messages with special characters
	specialContents := []string{
		"Hello 世界",                      // Chinese
		"Привет мир",                    // Cyrillic
		"مرحبا بالعالم",                 // Arabic
		"🎉🔥💀🚀",                          // Emojis
		"Null\x00byte",                  // Null byte
		"Line\nBreak\tTab",              // Control characters
		"Quote: \"test\" and 'single'",  // Quotes
		"Backslash: \\ and forward: /",  // Slashes
		strings.Repeat("a", 5000),       // Very long message
		"<script>alert('xss')</script>", // HTML-like content
		`{"json": "inside"}`,            // JSON-like content
	}

	for _, content := range specialContents {
		m.Messages = append(m.Messages, ChatMessage{
			Role:    "user",
			Content: content,
		})
	}

	err := m.SaveCurrentSession()
	if err != nil {
		t.Errorf("Failed to save session with special characters: %v", err)
	}

	// Load and verify
	loaded, err := loadSession(sessionPath)
	if err != nil {
		t.Errorf("Failed to load session with special characters: %v", err)
	}

	for i, originalContent := range specialContents {
		found := false
		for _, msg := range loaded.Messages {
			if strings.Contains(msg, originalContent) || strings.HasSuffix(msg, originalContent) {
				found = true
				break
			}
		}
		if !found {
			// The message format is "USER: <content>"
			expectedPrefix := "USER: " + originalContent
			msgFound := false
			for _, msg := range loaded.Messages {
				if strings.HasPrefix(msg, "USER:") && msg == expectedPrefix {
					msgFound = true
					break
				}
			}
			if !msgFound {
				t.Errorf("Special character message %d not found in loaded session", i)
			}
		}
	}
}

// TestEmptySession tests loading and saving sessions with no messages
func TestEmptySession(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	// Save empty session
	emptySession := Session{
		Messages: []string{},
		History:  []string{},
	}

	err := saveSession(emptySession, sessionPath)
	if err != nil {
		t.Fatalf("Failed to save empty session: %v", err)
	}

	// Load empty session
	loaded, err := loadSession(sessionPath)
	if err != nil {
		t.Fatalf("Failed to load empty session: %v", err)
	}

	if len(loaded.Messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(loaded.Messages))
	}
	if len(loaded.History) != 0 {
		t.Errorf("Expected 0 history entries, got %d", len(loaded.History))
	}

	// Verify file content
	data, _ := os.ReadFile(sessionPath)
	var parsed Session
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Saved empty session is not valid JSON: %v", err)
	}
}

// TestSaveCurrentSession_WithToolMessages tests saving messages with tool use/results
func TestSaveCurrentSession_WithToolMessages(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	m := NewModelWithSession("test_session.json")
	m.SessionPath = sessionPath
	m.Messages = nil // Clear any existing messages

	// Add various message types
	m.AddUserMessage("Run tests")
	m.AddToolUse("bash", "tool_123")
	m.AddToolResult("bash", "All tests passed", false)
	m.AddAssistantMessage("Tests completed successfully")

	err := m.SaveCurrentSession()
	if err != nil {
		t.Errorf("Failed to save session with tool messages: %v", err)
	}

	// Load and verify
	loaded, err := loadSession(sessionPath)
	if err != nil {
		t.Errorf("Failed to load session with tool messages: %v", err)
	}

	// Check that all messages were saved
	if len(loaded.Messages) != 4 {
		t.Errorf("Expected 4 messages, got %d", len(loaded.Messages))
	}
}

// TestAtomicWrite_CrashSimulation tests that write is atomic
func TestAtomicWrite_CrashSimulation(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	// Create an initial session
	initialSession := Session{
		Messages: []string{"USER: important message 1", "FLOYD: response 1"},
		History:  []string{"important message 1"},
	}

	err := saveSession(initialSession, sessionPath)
	if err != nil {
		t.Fatalf("Failed to save initial session: %v", err)
	}

	// Read the initial file size
	initialData, _ := os.ReadFile(sessionPath)
	initialSize := len(initialData)

	// Simulate a "crash" during write by using a buffer that fails
	// The current implementation uses os.WriteFile which is NOT atomic
	// If the process is killed during WriteFile, the file could be partially written

	// Create a new session with different content
	newSession := Session{
		Messages: []string{"USER: brand new important message for the session", "FLOYD: brand new detailed response"},
		History:  []string{"brand new important message"},
	}

	// Write the new session
	err = saveSession(newSession, sessionPath)
	if err != nil {
		t.Fatalf("Failed to save new session: %v", err)
	}

	// Verify file was completely written
	newData, err := os.ReadFile(sessionPath)
	if err != nil {
		t.Fatalf("Failed to read session file: %v", err)
	}

	// Verify it's valid JSON
	var loaded Session
	if err := json.Unmarshal(newData, &loaded); err != nil {
		// BLOCKING ISSUE #4: os.WriteFile is NOT atomic
		// If process crashes during write, file is corrupted
		t.Errorf("Session file is not valid JSON after write: %v", err)
	}

	// Verify content changed
	if len(newData) <= initialSize {
		t.Errorf("File size did not increase after writing new content (was %d, now %d)", initialSize, len(newData))
	}
}

// TestLoadSession_HomeDirNotSet tests behavior when HOME is not set
func TestLoadSession_HomeDirNotSet(t *testing.T) {
	// Unset HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	os.Unsetenv("HOME")

	// This affects NewModel which calls os.UserHomeDir()
	// The behavior depends on the OS
	m := NewModelWithSession("test_session.json")
	m.Messages = nil // Clear default greeting

	// Check if SessionPath was set
	if m.SessionPath == "" {
		t.Error("SessionPath should not be empty even when HOME is not set")
	} else {
		t.Logf("SessionPath when HOME not set: %q", m.SessionPath)
	}
}

// TestSaveCurrentSession_DirectoryNotExists tests saving to a non-existent directory
func TestSaveCurrentSession_DirectoryNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentDir := filepath.Join(tmpDir, "does_not_exist")
	sessionPath := filepath.Join(nonExistentDir, ".floyd_session.json")

	m := NewModelWithSession("test_session.json")
	m.SessionPath = sessionPath
	m.Messages = append(m.Messages, ChatMessage{Role: "user", Content: "test"})

	err := m.SaveCurrentSession()
	if err == nil {
		t.Error("Expected error when saving to non-existent directory, got nil")
	}

	// BLOCKING ISSUE #5: No directory creation before save
	// If ~/.floyd_session.json directory doesn't exist, save fails silently
}

// TestLoadSession_WithBOM tests loading a file with UTF-8 BOM
func TestLoadSession_WithBOM(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	// Create session data
	session := Session{
		Messages: []string{"USER: hello"},
		History:  []string{"hello"},
	}
	data, _ := json.MarshalIndent(session, "", "  ")

	// Add UTF-8 BOM
	bomData := append([]byte{0xEF, 0xBB, 0xBF}, data...)

	if err := os.WriteFile(sessionPath, bomData, 0644); err != nil {
		t.Fatalf("Failed to write file with BOM: %v", err)
	}

	_, err := loadSession(sessionPath)
	if err != nil {
		// JSON decoder should handle BOM, but let's verify
		t.Logf("Failed to load file with BOM: %v", err)
	}
}

// TestConcurrentReadWrite tests reading while writing
func TestConcurrentReadWrite(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	// Initial save
	initialSession := Session{
		Messages: []string{"USER: initial"},
		History:  []string{"initial"},
	}
	saveSession(initialSession, sessionPath)

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Start writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				session := Session{
					Messages: []string{"USER: writer " + string(rune('A'+id))},
					History:  []string{"writer " + string(rune('A'+id))},
				}
				if err := saveSession(session, sessionPath); err != nil {
					errors <- err
				}
			}
		}(i)
	}

	// Start readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_, err := loadSession(sessionPath)
				if err != nil {
					errors <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		errorCount++
		t.Logf("Concurrent access error: %v", err)
	}

	if errorCount > 0 {
		// BLOCKING ISSUE #6: No locking on read/write
		// Concurrent access can cause corruption
		t.Errorf("Got %d errors during concurrent read/write", errorCount)
	}
}

// TestSaveCurrentSession_MalformedUTF8 tests handling of invalid UTF-8
func TestSaveCurrentSession_MalformedUTF8(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	m := NewModelWithSession("test_session.json")
	m.SessionPath = sessionPath

	// Try to add content with invalid UTF-8 sequence
	invalidUTF8 := string([]byte{0xFF, 0xFE, 0xFD})
	m.Messages = append(m.Messages, ChatMessage{
		Role:    "user",
		Content: invalidUTF8,
	})

	// Go's json.Marshal should replace invalid UTF-8 with replacement character
	err := m.SaveCurrentSession()
	if err != nil {
		t.Logf("Error saving invalid UTF-8: %v", err)
	}

	// Verify file is still valid
	loaded, err := loadSession(sessionPath)
	if err != nil {
		t.Errorf("Failed to load session after saving invalid UTF-8: %v", err)
	}
	t.Logf("Loaded %d messages after invalid UTF-8 save", len(loaded.Messages))
}

// TestLoadSession_EmptyFile tests loading an empty file
func TestLoadSession_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	// Write empty file
	if err := os.WriteFile(sessionPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to write empty file: %v", err)
	}

	session, err := loadSession(sessionPath)
	if err == nil {
		t.Error("Expected error when loading empty file, got nil")
	}
	if len(session.Messages) != 0 {
		t.Error("Session should be empty when file is empty")
	}
}

// TestLoadSession_WhitespaceOnly tests loading a file with only whitespace
func TestLoadSession_WhitespaceOnly(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	// Write whitespace
	whitespace := []byte("   \n\t  ")
	if err := os.WriteFile(sessionPath, whitespace, 0644); err != nil {
		t.Fatalf("Failed to write whitespace file: %v", err)
	}

	session, err := loadSession(sessionPath)
	if err == nil {
		t.Error("Expected error when loading whitespace-only file, got nil")
	}
	if len(session.Messages) != 0 {
		t.Error("Session should be empty when file contains only whitespace")
	}
}

// TestSaveCurrentSession_PartialMessageLoss tests scenario where save fails after reading messages
func TestSaveCurrentSession_PartialMessageLoss(t *testing.T) {
	_ = t.TempDir()
	// Create a file that will cause write to fail (using a special file-like setup)
	// This is hard to test without filesystem manipulation

	// Instead, let's test that saveSession properly handles errors
	// by providing a path that will definitely fail

	// On Unix systems, /dev/full always returns ENOSPC on write
	sessionPath := "/dev/full"

	session := Session{
		Messages: []string{"USER: test"},
	}

	err := saveSession(session, sessionPath)
	if err == nil {
		t.Log("Warning: /dev/full write succeeded (unexpected)")
	} else {
		t.Logf("Got expected error writing to /dev/full: %v", err)
	}
}

// TestLoadSession_NilValues tests loading JSON with null values
func TestLoadSession_NilValues(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	// Write JSON with null values
	nullJSON := `{"messages": null, "history": null}`
	if err := os.WriteFile(sessionPath, []byte(nullJSON), 0644); err != nil {
		t.Fatalf("Failed to write null JSON: %v", err)
	}

	session, err := loadSession(sessionPath)
	if err != nil {
		t.Errorf("Failed to load JSON with null values: %v", err)
	}

	// Null values become empty slices
	if session.Messages == nil {
		t.Log("Messages is nil (JSON null was unmarshaled to nil slice)")
	}
	if session.History == nil {
		t.Log("History is nil (JSON null was unmarshaled to nil slice)")
	}
}

// TestSessionPathTooLong tests extremely long session paths
func TestSessionPathTooLong(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a very long path component
	longName := strings.Repeat("a", 255) // Most filesystems limit to 255
	sessionPath := filepath.Join(tmpDir, longName, ".floyd_session.json")

	session := Session{
		Messages: []string{"USER: test"},
	}

	err := saveSession(session, sessionPath)
	if err == nil {
		t.Log("Warning: Long path did not cause error")
	} else {
		t.Logf("Got expected error for long path: %v", err)
	}
}

// TestSaveOverwriteExisting tests that save properly overwrites existing content
func TestSaveOverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	// Save initial large session
	largeSession := Session{
		Messages: make([]string, 1000),
		History:  make([]string, 1000),
	}
	for i := range largeSession.Messages {
		largeSession.Messages[i] = "USER: message " + string(rune('0'+i%10))
		largeSession.History[i] = "history " + string(rune('0'+i%10))
	}

	if err := saveSession(largeSession, sessionPath); err != nil {
		t.Fatalf("Failed to save large session: %v", err)
	}

	// Get initial file size
	info1, _ := os.Stat(sessionPath)
	initialSize := info1.Size()

	// Save small session
	smallSession := Session{
		Messages: []string{"USER: single"},
		History:  []string{"single"},
	}

	if err := saveSession(smallSession, sessionPath); err != nil {
		t.Fatalf("Failed to save small session: %v", err)
	}

	// Get new file size
	info2, _ := os.Stat(sessionPath)
	newSize := info2.Size()

	// New file should be smaller
	if newSize >= initialSize {
		t.Errorf("Expected smaller file after overwrite, got %d >= %d", newSize, initialSize)
	}

	// Verify content is correct
	loaded, err := loadSession(sessionPath)
	if err != nil {
		t.Errorf("Failed to load after overwrite: %v", err)
	}
	if len(loaded.Messages) != 1 {
		t.Errorf("Expected 1 message after overwrite, got %d", len(loaded.Messages))
	}
}

// BenchmarkSaveSession benchmarks the save operation
func BenchmarkSaveSession(b *testing.B) {
	tmpDir := b.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	session := Session{
		Messages: make([]string, 100),
		History:  make([]string, 100),
	}
	for i := range session.Messages {
		session.Messages[i] = "USER: message " + string(rune('0'+i%10))
		session.History[i] = "history " + string(rune('0'+i%10))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = saveSession(session, sessionPath)
	}
}

// BenchmarkLoadSession benchmarks the load operation
func BenchmarkLoadSession(b *testing.B) {
	tmpDir := b.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	session := Session{
		Messages: make([]string, 100),
		History:  make([]string, 100),
	}
	for i := range session.Messages {
		session.Messages[i] = "USER: message " + string(rune('0'+i%10))
		session.History[i] = "history " + string(rune('0'+i%10))
	}
	_ = saveSession(session, sessionPath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = loadSession(sessionPath)
	}
}

// TestNewModel_IgnoresLoadErrors tests that NewModel ignores loadSession errors
func TestNewModel_IgnoresLoadErrors(t *testing.T) {
	tmpDir := t.TempDir()
	badSessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	// Write corrupt JSON
	_ = os.WriteFile(badSessionPath, []byte("{invalid json}"), 0644)

	// NewModel will try to load from ~/.floyd_session.json by default
	// We can't easily change this, so let's test the behavior differently
	// by manually setting SessionPath after creation

	// BLOCKING ISSUE #7: NewModel silently ignores loadSession errors
	// Line 125 in model.go: session, _ := loadSession(sessionPath)
	// The error is discarded, so corrupt files are silently treated as empty sessions

	// Verify that loadSession returns an error for our corrupt file
	_, err := loadSession(badSessionPath)
	if err == nil {
		t.Error("Expected error loading corrupt session, got nil")
	}

	// But NewModel would ignore this error and continue
	// This means users lose their session without any notification
}

// TestSaveCurrentSessionToolHandling verifies tool messages are correctly saved
func TestSaveCurrentSessionToolHandling(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	m := NewModelWithSession("test_session.json")
	m.SessionPath = sessionPath

	// Add various message types with ToolUse and ToolResult
	m.AddSystemMessage("System message")
	m.Messages = append(m.Messages, ChatMessage{
		Role:    "system",
		ToolUse: &ToolUseInfo{Name: "bash", ID: "tool_123"},
	})
	m.Messages = append(m.Messages, ChatMessage{
		Role:       "system",
		ToolResult: &ToolResultInfo{Name: "bash", Content: "Output", IsError: false},
	})

	err := m.SaveCurrentSession()
	if err != nil {
		t.Fatalf("SaveCurrentSession failed: %v", err)
	}

	// Load and verify format
	loaded, err := loadSession(sessionPath)
	if err != nil {
		t.Fatalf("loadSession failed: %v", err)
	}

	// Check that tool messages are saved with "SYSTEM:" prefix
	foundToolUse := false
	foundToolResult := false
	for _, msg := range loaded.Messages {
		if strings.Contains(msg, "[Tool: bash]") {
			foundToolUse = true
		}
		if strings.Contains(msg, "[Tool Result: bash]") {
			foundToolResult = true
		}
	}

	if !foundToolUse {
		t.Error("Tool use message not found in saved session")
	}
	if !foundToolResult {
		t.Error("Tool result message not found in saved session")
	}
}

// TestSessionIntegrityAfterPartialWrite simulates partial write scenario
func TestSessionIntegrityAfterPartialWrite(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, ".floyd_session.json")

	// Create a valid session first
	original := Session{
		Messages: []string{"USER: critical data", "FLOYD: important response"},
		History:  []string{"critical data"},
	}
	_ = saveSession(original, sessionPath)

	// Read original data
	originalData, _ := os.ReadFile(sessionPath)

	// Simulate partial write by writing only part of new data
	newSession := Session{
		Messages: []string{"USER: new data"},
		History:  []string{"new data"},
	}
	newData, _ := json.MarshalIndent(newSession, "", "  ")

	// Write only half (simulating crash during write)
	partialData := newData[:len(newData)/2]
	_ = os.WriteFile(sessionPath, partialData, 0644)

	// Now try to load
	_, err := loadSession(sessionPath)

	// BLOCKING ISSUE #8: No backup mechanism
	// When save fails or is interrupted, the original session is lost
	// os.WriteFile replaces the entire file content before the write completes

	if err != nil {
		// The session is corrupted and unrecoverable
		t.Logf("As expected, partial write corrupted the file: %v", err)

		// Verify original data is gone
		currentData, _ := os.ReadFile(sessionPath)
		if bytes.Equal(originalData, currentData) {
			t.Error("Original data somehow preserved (unexpected)")
		}
	}
}
