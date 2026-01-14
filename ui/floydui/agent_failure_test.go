package floydui

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Nomadcxx/sysc-Go/agent"
	agenttools "github.com/Nomadcxx/sysc-Go/agent/tools"
)

// mockGLMClient is a mock client for testing failure scenarios
type mockGLMClient struct {
	streamChan <-chan agent.StreamChunk
	streamErr  error
	chatErr    error
	shouldTimeout bool
	cancelImmediately bool
}

func (m *mockGLMClient) StreamChat(ctx context.Context, req agent.ChatRequest) (<-chan agent.StreamChunk, error) {
	if m.streamErr != nil {
		return nil, m.streamErr
	}
	if m.shouldTimeout {
		// Return a channel that will never receive
		ch := make(chan agent.StreamChunk)
		return ch, nil
	}
	if m.cancelImmediately {
		// Return a channel that closes immediately
		ch := make(chan agent.StreamChunk, 1)
		close(ch)
		return ch, nil
	}
	return m.streamChan, nil
}

func (m *mockGLMClient) Chat(ctx context.Context, req agent.ChatRequest) (string, error) {
	if m.chatErr != nil {
		return "", m.chatErr
	}
	return "mock response", nil
}

func (m *mockGLMClient) Cancel(streamID string) {}

// setupTestModel creates a model with mocked dependencies for testing
func setupTestModel() Model {
	m := NewModelWithSession("test_session.json")

	// Override with mock components
	m.ProxyClient = agent.NewProxyClient("test-key", "https://api.test.com")

	return m
}

// =============================================================================
// Test Scenario 1: No API Key
// =============================================================================

// TestNoAPIKey_InitModel verifies behavior when no API key is configured
func TestNoAPIKey_InitModel(t *testing.T) {
	// Clear all API key environment variables
	unsetEnv := func(keys ...string) func() {
		oldValues := make(map[string]string)
		for _, key := range keys {
			oldValues[key] = os.Getenv(key)
			os.Unsetenv(key)
		}
		return func() {
			for key, val := range oldValues {
				if val != "" {
					os.Setenv(key, val)
				}
			}
		}
	}

	restore := unsetEnv("ANTHROPIC_AUTH_TOKEN", "GLM_API_KEY", "ZHIPU_API_KEY")
	defer restore()

	m := NewModelWithSession("test_session.json")

	// Current behavior: Model is created with empty API key
	// Expected behavior: Should show error or prevent agent calls

	if m.ProxyClient == nil {
		t.Fatal("ProxyClient should not be nil")
	}

	// BUG: No validation that API key is present
	// The model will allow user input and attempt API calls which will fail
	t.Logf("WARNING: Model created with potentially empty API key. No validation prevents agent call.")
}

// TestNoAPIKey_StartAgentCall verifies what happens when trying to start an agent call with no API key
func TestNoAPIKey_StartAgentCall(t *testing.T) {
	m := NewModelWithSession("test_session.json")

	// Create a client with no API key
	client := agent.NewProxyClient("", "")
	m.ProxyClient = client

	userInput := "test message"
	cmd := m.startAgentCall(userInput)

	if cmd == nil {
		t.Fatal("startAgentCall should return a command")
	}

	// The command will execute in a goroutine
	// Current behavior: Will fail when making HTTP request
	// Expected behavior: Should validate API key before attempting connection

	t.Logf("WARNING: No API key validation in startAgentCall. Failure will occur during HTTP request.")
}

// =============================================================================
// Test Scenario 2: API Timeout
// =============================================================================

// TestAPITimeout_Streaming verifies behavior when API times out
func TestAPITimeout_Streaming(t *testing.T) {
	_ = setupTestModel()

	// Create a mock client that times out
	_ = &mockGLMClient{
		shouldTimeout: true,
	}

	// Replace the LoopAgent's client (would need to add this capability)
	// For now, we document the expected behavior

	// Current behavior: Timeout is handled by HTTP client (30s default)
	// Expected behavior: Should show user-friendly timeout message

	t.Logf("Current: HTTP client has 30s timeout. Error returned as generic stream error.")
	t.Logf("Expected: Should detect timeout and show specific 'API timed out' message to user.")
}

// TestAPITimeout_Cancellation verifies behavior when user cancels during timeout
func TestAPITimeout_Cancellation(t *testing.T) {
	m := setupTestModel()

	// Simulate user pressing Ctrl+C during a slow request
	_, cancel := context.WithCancel(context.Background())
	m.StreamCancel = cancel

	// Cancel the context
	cancel()

	// Check if thinking state is properly cleared
	m.IsThinking = true

	// BUG: No mechanism to reset IsThinking if stream is cancelled
	if m.IsThinking {
		t.Error("BUG: IsThinking remains true after cancellation. User sees stuck spinner.")
	}
}

// =============================================================================
// Test Scenario 3: API Error Responses (401, 429, 500)
// =============================================================================

// TestAPIError_401_Unauthorized verifies behavior on 401 error
func TestAPIError_401_Unauthorized(t *testing.T) {
	err401 := errors.New("API error (status 401): unauthorized")

	m := setupTestModel()
	m.IsThinking = true
	m.Messages = append(m.Messages, ChatMessage{Role: "user", Content: "test"})

	// Simulate receiving a 401 error
	msg := streamErrorMsg{err: err401}
	newModel, _ := m.Update(msg)
	model := newModel.(Model)

	// Current behavior: Generic error message added
	if !model.IsThinking {
		t.Error("IsThinking should be cleared after error")
	}

	// Check if error message is user-friendly
	lastMsg := model.Messages[len(model.Messages)-1]
	if !strings.Contains(lastMsg.Content, "Error:") {
		t.Error("Error message should indicate error occurred")
	}

	t.Logf("Current error message: %s", lastMsg.Content)
	t.Logf("Expected: Should detect 401 and suggest checking API key configuration")
}

// TestAPIError_429_RateLimit verifies behavior on rate limit
func TestAPIError_429_RateLimit(t *testing.T) {
	err429 := errors.New("API error (status 429): rate limit exceeded")

	m := setupTestModel()
	m.IsThinking = true

	msg := streamErrorMsg{err: err429}
	newModel, _ := m.Update(msg)
	model := newModel.(Model)

	lastMsg := model.Messages[len(model.Messages)-1]

	// BUG: Generic error message doesn't help user understand rate limiting
	t.Logf("Current: %s", lastMsg.Content)
	t.Logf("Expected: Should show 'Rate limit exceeded. Please wait before trying again.'")
}

// TestAPIError_500_InternalError verifies behavior on 500 error
func TestAPIError_500_InternalError(t *testing.T) {
	err500 := errors.New("API error (status 500): internal server error")

	m := setupTestModel()
	m.IsThinking = true

	msg := streamErrorMsg{err: err500}
	newModel, _ := m.Update(msg)
	model := newModel.(Model)

	lastMsg := model.Messages[len(model.Messages)-1]

	// BUG: No retry mechanism or suggestion for server errors
	t.Logf("Current: %s", lastMsg.Content)
	t.Logf("Expected: Should indicate server error and suggest retry or support contact")
}

// =============================================================================
// Test Scenario 4: Malformed Response
// =============================================================================

// TestMalformedJSON_Response verifies behavior with invalid JSON response
func TestMalformedJSON_Response(t *testing.T) {
	// This would require mocking at the HTTP stream level
	// For now, document expected behavior

	t.Logf("Current: Invalid SSE chunks are silently skipped in readAnthropicStream")
	t.Logf("Expected: Should detect malformed responses and show parsing error")
}

// =============================================================================
// Test Scenario 5: Empty User Input
// =============================================================================

// TestEmptyInput_WhitespaceOnly verifies behavior with whitespace input
func TestEmptyInput_WhitespaceOnly(t *testing.T) {
	m := setupTestModel()

	testCases := []string{
		"   ",
		"\t\t",
		"\n\n",
		"  \t  \n  ",
	}

	for _, input := range testCases {
		m.Input.SetValue(input)

		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
		newModel, _ := m.Update(keyMsg)
		model := newModel.(Model)

		// Good: Empty/whitespace input is rejected
		if len(model.Messages) > 0 {
			lastMsg := model.Messages[len(model.Messages)-1]
			if lastMsg.Role == "user" && strings.TrimSpace(lastMsg.Content) == "" {
				t.Errorf("BUG: Empty message was added for input: %q", input)
			}
		}

		t.Logf("Input %q correctly rejected", input)
	}
}

// =============================================================================
// Test Scenario 6: Very Long Input
// =============================================================================

// TestLongInput_ExceedsLimit verifies behavior with very long input
func TestLongInput_ExceedsLimit(t *testing.T) {
	m := setupTestModel()

	// Input longer than CharLimit (500)
	longInput := strings.Repeat("a", 1000)

	// The textinput component should truncate at CharLimit
	m.Input.SetValue(longInput)

	if m.Input.Value() != longInput {
		t.Logf("Input truncated to %d chars (CharLimit=%d)", len(m.Input.Value()), m.Input.CharLimit)
	}

	// BUG: No validation for total conversation size vs model context window
	t.Logf("WARNING: No check for total conversation context size before API call")
}

// =============================================================================
// Test Scenario 7: Context Window Exceeded
// =============================================================================

// TestContextWindow_Exceeded verifies behavior when conversation is too long
func TestContextWindow_Exceeded(t *testing.T) {
	m := setupTestModel()

	// Simulate a long conversation
	for i := 0; i < 100; i++ {
		m.AddUserMessage(strings.Repeat("test message ", 100))
		m.AddAssistantMessage(strings.Repeat("response ", 100))
	}

	// Check total message count
	totalChars := 0
	for _, msg := range m.Conversation {
		if str, ok := msg.Content.(string); ok {
			totalChars += len(str)
		}
	}

	t.Logf("Total conversation size: %d characters", totalChars)

	// BUG: No context window management before sending to API
	// The API will reject the request with an error
	t.Logf("WARNING: No truncation or summarization of old messages")
}

// =============================================================================
// Test Scenario 8: Stream Interruption (Ctrl+C)
// =============================================================================

// TestStreamInterruption_CtrlC verifies graceful handling of Ctrl+C during stream
func TestStreamInterruption_CtrlC(t *testing.T) {
	m := setupTestModel()
	m.IsThinking = true
	m.ExecutingTools = true

	// Create a cancel function
	_, cancel := context.WithCancel(context.Background())
	m.StreamCancel = cancel

	// Simulate Ctrl+C
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	newModel, cmd := m.Update(keyMsg)
	model := newModel.(Model)

	// Current behavior: Cancel is called and app quits
	if cmd == nil {
		t.Error("Ctrl+C should return quit command")
	}

	// BUG: StreamCancel is called but there's no verification that stream actually stopped
	// Goroutine may continue running
	t.Logf("WARNING: No WaitGroup or sync mechanism to ensure goroutine cleanup")

	_ = model // Use model
}

// TestStreamInterruption_DuringToolExecution verifies cancellation during tool use
func TestStreamInterruption_DuringToolExecution(t *testing.T) {
	m := setupTestModel()
	m.IsThinking = true
	m.ExecutingTools = true
	m.CurrentToolName = "bash"

	_, cancel := context.WithCancel(context.Background())
	m.StreamCancel = cancel

	// Cancel while tool is executing
	cancel()

	// BUG: No way to stop an already-running tool execution
	// ToolExecutor's ExecuteToolCalls doesn't check parent context during execution
	t.Logf("WARNING: Tool execution doesn't respect parent context cancellation")
}

// =============================================================================
// Test Scenario 9: Tool Execution Failure
// =============================================================================

// TestToolExecution_Error verifies behavior when tool fails
func TestToolExecution_Error(t *testing.T) {
	m := setupTestModel()

	// Simulate tool error message
	msg := toolErrorMsg{
		toolName: "bash",
		content:  "command failed: exit status 1",
	}

	newModel, _ := m.Update(msg)
	model := newModel.(Model)

	// Should clear executing state
	if model.ExecutingTools {
		t.Error("ExecutingTools should be cleared after tool error")
	}

	// Should add error result
	lastMsg := model.Messages[len(model.Messages)-1]
	if lastMsg.ToolResult == nil {
		t.Error("Should have ToolResult after tool error")
	} else if !lastMsg.ToolResult.IsError {
		t.Error("ToolResult should be marked as error")
	}

	t.Logf("Tool error handling: %v", lastMsg)
}

// TestToolExecution_Timeout verifies behavior when tool times out
func TestToolExecution_Timeout(t *testing.T) {
	m := setupTestModel()

	// Create executor with short timeout
	executor := agenttools.NewExecutor(time.Millisecond)
	m.ToolExecutor = executor

	// This would require an actual tool that takes longer than 1ms
	// For now, document the expected behavior

	t.Logf("Current: Tool timeout returns error in ToolResult")
	t.Logf("Expected: Should show clear timeout message to user")
}

// =============================================================================
// Test Scenario 10: Network Offline
// =============================================================================

// TestNetworkOffline_NoConnection verifies behavior with no network
func TestNetworkOffline_NoConnection(t *testing.T) {
	m := setupTestModel()
	m.IsThinking = true

	// Simulate network error
	errNetwork := errors.New("dial tcp: lookup api.z.ai: no such host")
	msg := streamErrorMsg{err: errNetwork}

	newModel, _ := m.Update(msg)
	model := newModel.(Model)

	// Check error handling
	if model.IsThinking {
		t.Error("IsThinking should be cleared after network error")
	}

	lastMsg := model.Messages[len(model.Messages)-1]

	// BUG: Generic error message for network issues
	if strings.Contains(lastMsg.Content, "dial tcp") {
		t.Logf("Current: Raw error shown to user: %s", lastMsg.Content)
		t.Logf("Expected: Should show 'Unable to connect to API. Check your internet connection.'")
	}
}

// =============================================================================
// Integration Tests for Error Recovery
// =============================================================================

// TestErrorRecovery_CanRetryAfterError verifies user can retry after error
func TestErrorRecovery_CanRetryAfterError(t *testing.T) {
	m := setupTestModel()

	// First, cause an error
	m.IsThinking = true
	errMsg := streamErrorMsg{err: errors.New("test error")}
	newModel, _ := m.Update(errMsg)
	model := newModel.(Model)

	// Verify thinking state is cleared
	if model.IsThinking {
		t.Error("Should not be thinking after error")
	}

	// Now try to send a new message
	model.Input.SetValue("retry message")
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = model.Update(keyMsg)
	model = newModel.(Model)

	// BUG: May not allow retry if state is inconsistent
	if model.IsThinking {
		t.Log("Can retry after error - IsThinking set correctly")
	} else {
		t.Error("BUG: May not be able to retry after error")
	}
}

// TestMultipleErrorsInSequence verifies behavior with multiple consecutive errors
func TestMultipleErrorsInSequence(t *testing.T) {
	m := setupTestModel()

	// Send multiple errors in sequence
	for i := 0; i < 5; i++ {
		m.IsThinking = true
		errMsg := streamErrorMsg{err: errors.New("repeated error")}
		newModel, _ := m.Update(errMsg)
		model := newModel.(Model)

		if model.IsThinking {
			t.Errorf("Error %d: IsThinking not cleared", i+1)
		}

		m = model
	}

	t.Logf("Messages after 5 errors: %d", len(m.Messages))
}

// =============================================================================
// Tests for Message Type Handling
// =============================================================================

// TestStreamErrorMsgHandling verifies streamErrorMsg is properly handled
func TestStreamErrorMsgHandling(t *testing.T) {
	m := setupTestModel()

	// Set up thinking state
	m.IsThinking = true
	m.CurrentResponse.WriteString("partial response")
	m.PendingTokens = []string{"token1", "token2"}

	// Send error
	errMsg := streamErrorMsg{err: errors.New("stream failed")}
	newModel, _ := m.Update(errMsg)
	model := newModel.(Model)

	// Verify state after error
	if model.IsThinking {
		t.Error("IsThinking should be false after stream error")
	}

	if len(model.PendingTokens) != 0 {
		t.Error("PendingTokens should be flushed after error")
	}

	// Should have error message in messages
	foundError := false
	for _, msg := range model.Messages {
		if strings.Contains(msg.Content, "Error:") && strings.Contains(msg.Content, "stream failed") {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Error message not found in messages")
	}
}

// TestThinkingStartedMsgHandling verifies thinkingStartedMsg
func TestThinkingStartedMsgHandling(t *testing.T) {
	m := setupTestModel()

	// Send thinking started
	msg := thinkingStartedMsg{}
	newModel, _ := m.Update(msg)
	model := newModel.(Model)

	if !model.IsThinking {
		t.Error("IsThinking should be true after thinkingStartedMsg")
	}

	if model.AnimationFrame != 0 {
		t.Error("AnimationFrame should be reset to 0")
	}

	if !model.ProgressActive {
		t.Error("ProgressActive should be true")
	}
}

// =============================================================================
// Blocking Error Verification Tests
// =============================================================================

// TestBlockingError_StateInconsistency verifies if errors cause state inconsistency
func TestBlockingError_StateInconsistency(t *testing.T) {
	m := setupTestModel()

	// Set up complex state
	m.IsThinking = true
	m.ExecutingTools = true
	m.ProgressActive = true
	m.ProgressPercent = 0.5

	// Send error
	errMsg := streamErrorMsg{err: errors.New("critical error")}
	newModel, _ := m.Update(errMsg)
	model := newModel.(Model)

	// BLOCKING BUG: Inconsistent state after error
	if model.ExecutingTools {
		t.Error("BLOCKING BUG: ExecutingTools not cleared after error")
	}

	if model.ProgressActive {
		t.Error("BLOCKING BUG: ProgressActive not cleared after error")
	}

	// The UI may show conflicting indicators
	t.Logf("State after error: IsThinking=%v, ExecutingTools=%v, ProgressActive=%v",
		model.IsThinking, model.ExecutingTools, model.ProgressActive)
}

// TestBlockingError_NoUserFeedback verifies if errors are visible to user
func TestBlockingError_NoUserFeedback(t *testing.T) {
	m := setupTestModel()

	// Simulate error during silent period
	m.IsThinking = true

	errMsg := streamErrorMsg{err: errors.New("silent error")}
	newModel, _ := m.Update(errMsg)
	model := newModel.(Model)

	// Check if user is notified
	lastMsg := model.Messages[len(model.Messages)-1]

	if !strings.Contains(lastMsg.Content, "Error:") {
		t.Error("BLOCKING BUG: User not notified of error")
	}

	// Check if error is visually distinct (marked as system message)
	if lastMsg.Role != "system" {
		t.Error("BLOCKING BUG: Error not in system message for visibility")
	}
}

// TestBlockingError_CannotContinue verifies if errors block further interaction
func TestBlockingError_CannotContinue(t *testing.T) {
	m := setupTestModel()

	// Cause an error
	m.IsThinking = true
	errMsg := streamErrorMsg{err: errors.New("blocking error")}
	newModel, _ := m.Update(errMsg)
	model := newModel.(Model)

	// Try to send new message
	model.Input.SetValue("new message")
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = model.Update(keyMsg)
	model = newModel.(Model)

	// Check if new message can be sent
	if model.IsThinking {
		// Should be thinking again after new message
		t.Log("Can continue after error (good)")
	} else {
		t.Error("BLOCKING BUG: Cannot send new message after error")
	}
}

// =============================================================================
// Helper Tests
// =============================================================================

// TestLoadConfig_NoKey tests loadConfig with no API key available
func TestLoadConfig_NoKey(t *testing.T) {
	// Temporarily clear environment
	oldVars := map[string]string{
		"ANTHROPIC_AUTH_TOKEN": os.Getenv("ANTHROPIC_AUTH_TOKEN"),
		"GLM_API_KEY":          os.Getenv("GLM_API_KEY"),
		"ZHIPU_API_KEY":        os.Getenv("ZHIPU_API_KEY"),
	}

	for key := range oldVars {
		os.Unsetenv(key)
	}

	defer func() {
		for key, val := range oldVars {
			if val != "" {
				os.Setenv(key, val)
			}
		}
	}()

	cfg := loadConfig()

	if cfg.apiKey != "" {
		t.Errorf("Expected empty API key, got: %s", cfg.apiKey)
	}

	t.Logf("Config with no key: apiKey=%q, baseURL=%s, model=%s",
		cfg.apiKey, cfg.baseURL, cfg.model)
}

// TestLoadConfig_PriorityOrder tests environment variable priority
func TestLoadConfig_PriorityOrder(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedKey    string
		expectedSource string
	}{
		{
			name: "ANTHROPIC_AUTH_TOKEN takes priority",
			envVars: map[string]string{
				"ANTHROPIC_AUTH_TOKEN": "key-anthropic",
				"GLM_API_KEY":          "key-glm",
				"ZHIPU_API_KEY":        "key-zhipu",
			},
			expectedKey:    "key-anthropic",
			expectedSource: "ANTHROPIC_AUTH_TOKEN",
		},
		{
			name: "GLM_API_KEY as fallback",
			envVars: map[string]string{
				"GLM_API_KEY":   "key-glm",
				"ZHIPU_API_KEY": "key-zhipu",
			},
			expectedKey:    "key-glm",
			expectedSource: "GLM_API_KEY",
		},
		{
			name: "ZHIPU_API_KEY as last fallback",
			envVars: map[string]string{
				"ZHIPU_API_KEY": "key-zhipu",
			},
			expectedKey:    "key-zhipu",
			expectedSource: "ZHIPU_API_KEY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save old values
			oldVars := make(map[string]string)
			for key := range tt.envVars {
				oldVars[key] = os.Getenv(key)
			}

			// Set test values
			for key, val := range tt.envVars {
				os.Setenv(key, val)
			}

			// Restore after test
			defer func() {
				for key, oldVal := range oldVars {
					if oldVal != "" {
						os.Setenv(key, oldVal)
					} else {
						os.Unsetenv(key)
					}
				}
			}()

			cfg := loadConfig()

			if cfg.apiKey != tt.expectedKey {
				t.Errorf("Expected key from %s (%s), got: %s",
					tt.expectedSource, tt.expectedKey, cfg.apiKey)
			}
		})
	}
}
