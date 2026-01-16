# FLOYD AGENT

**Last Updated:** 2026-01-12

## Goal

Build a GLM-4.7 powered coding agent that competes with Claude Code - using a $270/year unlimited GLM Mac Code plan instead of monthly $20 Claude subscriptions.

## Current Status

### ✅ Working
- **Supercache Protocol** - Installed at `.floyd/AGENT_INSTRUCTIONS.md`
- **3-Tier Cache Backend** - `cache/` package with reasoning (5min), project (24h), vault (7d)
- **Tool Registry** - `bash`, `read`, `write`, `edit`, `multiedit`, `grep`, `ls`, `glob`, `cache`
- **Agent Loop** - `agent/loop/` handles tool calling and streaming
- **Protocol Manager** - `agent/floyd/` manages external memory and safety rules
- **Global Install** - `floyd` and `pink-floyd` in `/usr/local/bin/`
- **Agent API Connection** - Successfully connects and streams responses

### ✅ Fixed Issues (Critical)

#### 1. Agent Not Responding (API 422 Error)
**Symptom:** Agent would show "Thinking..." but never produce any response.

**Root Cause:** The Anthropic-compatible API (`api.z.ai`) rejects messages with `role: "system"`. The API only accepts `"user"` or `"assistant"` roles. System prompts must be passed via the **`System` field** on the request, not as a message.

**Error:** `API error (status 422): Input should be 'user' or 'assistant', input: "system"`

**Fix:** Updated `agent/floyd/protocol.go` `EnhancedChatRequest` to:
- Use `prompt.LoadSystemPrompt(mode)` which returns a string
- Set `req.System = systemPrompt` (correct API format)
- Filter out any `role: "system"` messages from the conversation

#### 2. TUI Panic on Channel Close
**Symptom:** TUI would panic with "send on closed channel" after first agent response.

**Root Cause:** `TokenStreamChan` was being closed by a goroutine after each request, but subsequent requests would try to write to the closed channel.

**Fix:** Implemented persistent channel pattern:
- `TokenStreamChan` is created once in `NewModel` and **never closed**
- Use sentinel value `\x00DONE` to signal stream completion
- Added `waitForToken()` command loop to continuously read from channel

#### 3. Token Reading Loop Not Started
**Symptom:** Agent connects but tokens are never displayed in the UI.

**Root Cause:** When `thinkingStartedMsg` was received, it didn't return a `waitForToken()` command, so nothing was reading from the channel.

**Fix:** Added `return m, tea.Batch(Tick(...), m.waitForToken())` to the `thinkingStartedMsg` handler.

#### 4. strings.Builder Copy Panic
**Symptom:** TUI panics with `"strings: illegal use of non-zero Builder copied by value"`.

**Root Cause:** `strings.Builder` cannot be copied by value once it has been written to. Bubble Tea copies the Model struct on each `Update` call, which triggers this panic.

**Fix:** Changed `CurrentResponse` from `strings.Builder` to `*strings.Builder` (pointer) in the Model struct, and initialized it as `&strings.Builder{}` in `NewModel`.

#### 5. Other Fixed Issues
- **TUI Stability** - Fixed crashes on resize and small screens
- **Input Flickering** - Fixed cursor strobing on updates  
- **Timeouts** - Increased limits for reliable tool execution (5m) and thinking (120s)
- **Concurrency** - Fixed `copylocks` issues in Model struct

### 🔄 In Progress
- Simple CLI mode (`cmd/floyd-cli/` started but not integrated)

## Project Structure

```
├── agent/              # Core agent logic
│   ├── client.go       # API client (GLM-4.7 via Anthropic-compatible endpoint)
│   ├── floyd/          # FLOYD protocol manager
│   ├── loop/           # Tool calling loop
│   ├── message/        # Message types
│   ├── prompt/         # System prompt loader
│   └── tools/          # Tool executor and schema
├── cache/              # 3-tier SUPERCACHE backend
│   ├── frame.go        # Tier 1: reasoning (5min)
│   ├── chronicle.go    # Tier 2: project (24h)
│   ├── vault.go        # Tier 3: vault (7d)
│   └── manager.go      # Unified cache manager
├── cmd/                # Entry points
│   ├── floyd/          # Main CLI/TUI (uses shared ui/floyd)
│   ├── pink-floyd/     # TUI variant (same as floyd)
│   └── floyd-cli/      # Simple CLI (in progress)
├── tui/                # TUI components
│   ├── floyd_mode.go   # Chrome tool templates (not implemented)
│   └── floydtools/     # Tool implementations
├── ui/floydui/         # Shared UI package
│   ├── theme.go        # Theme definitions
│   ├── styles.go       # Lipgloss styles
│   ├── assets.go       # Banners, thinking phrases
│   ├── model.go        # Core model
│   ├── update.go       # Update logic
│   └── view.go         # View rendering
└── .floyd/             # External memory
    ├── AGENT_INSTRUCTIONS.md  # Full FLOYD protocol
    ├── master_plan.md         # Project goals
    ├── progress.md            # Execution log
    ├── scratchpad.md          # Error notes
    ├── stack.md               # Tech stack
    └── branch.md              # PR checklist
```

## FLOYD-S SUPERCACHE Protocol

The 3-tier cache is defined in `.floyd/AGENT_INSTRUCTIONS.md`:

```xml
<floyd_s_supercache>
  <tier_1_reasoning>
    <tier_id>reasoning</tier_id>
    <ttl>5 minutes</ttl>
    <description>Current conversation context</description>
  </tier_1_reasoning>

  <tier_2_project>
    <tier_id>project</tier_id>
    <ttl>24 hours</ttl>
    <description>Project-specific context</description>
  </tier_2_project>

  <tier_3_vault>
    <tier_id>vault</tier_id>
    <ttl>7 days</ttl>
    <description>Reusable solutions</description>
  </tier_3_vault>
</floyd_s_supercache>
```

## Tools

| Tool | Description | Status |
|------|-------------|--------|
| `bash` | Run shell commands | ✅ |
| `read` | Read file contents | ✅ |
| `write` | Write files | ✅ |
| `edit` | Find/replace in files | ✅ |
| `multiedit` | Multiple edits at once | ✅ |
| `grep` | Search files with regex | ✅ |
| `ls` | List directories | ✅ |
| `glob` | Find files by pattern | ✅ |
| `cache` | Manage SUPERCACHE tiers | ✅ |
| `Chrome: navigate` | Browser navigation | ❌ Template only |
| `Chrome: computer` | Browser automation | ❌ Template only |

## API Configuration

- **Endpoint:** `https://api.z.ai/api/anthropic`
- **Model:** `claude-opus-4` → maps to GLM-4.7
- **Format:** Anthropic API compatible
- **Streaming:** Supported

### Critical API Requirements

1. **System prompts use `System` field**, NOT messages with `role: "system"`
2. Messages array may only contain `role: "user"` or `role: "assistant"`
3. The `EnhancedChatRequest` function in `agent/floyd/protocol.go` handles this correctly

## Environment Variables

```
ANTHROPIC_AUTH_TOKEN  # API key (primary)
GLM_API_KEY           # API key (fallback)
ZHIPU_API_KEY         # API key (fallback)
```

## Troubleshooting

### Agent shows "Thinking..." but never responds
1. Check API key is set: `echo $ANTHROPIC_AUTH_TOKEN` or `echo $GLM_API_KEY`
2. Run test: `go run ./cmd/test_agent/main.go`
3. Look for API error in output (e.g., 401, 422, 500)

### TUI panics with "send on closed channel"
This was fixed. If it recurs, check that:
- `TokenStreamChan` is never closed (use `\x00DONE` sentinel instead)
- See `ui/floydui/update.go` and `ui/floydui/model.go` for the pattern

### TUI shows "Terminal too small"
Resize your terminal to at least 20 lines height.

### Agent times out (120s)
- Check network connectivity to `api.z.ai`
- Increase timeout in `ui/floydui/update.go` if needed

## Next Steps

1. ~~Fix TUI~~ ✅ Fixed
2. Implement Chrome tools (if needed)
3. Add sub-agent orchestration (spawn specialists)
4. Improve error handling and logging
5. Complete simple CLI mode
