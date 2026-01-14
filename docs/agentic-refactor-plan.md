# Agentic Refactor Plan

**The Core Problem:** The current TUI is a sophisticated tool launcher and chat interface, not an agentic system. It's built like a Swiss Army knife where the user selects each tool. Claude Code is a self-driving car where you give it a destination (a goal) and it plans the route and drives itself.

---

## The Core Architectural Gaps

From the codebase, three fundamental design differences explain the "chatbot frame" feeling:

### 1. No Autonomous Agent Loop
The `floyd_mode.go` `beginFloydStream` and `handleFloydStream` functions react to user-initiated tool calls. There's no internal engine that runs a **"Gather → Act → Verify"** cycle. The AI is not in the driver's seat.

### 2. Tools are Manual, Not Discoverable
The `tool.go` interface and `FloydTool` structs are for human selection from a UI list. In an agent system, tools should be discovered and called dynamically by the LLM via a protocol like MCP. The Chrome tool templates are just JSON snippets to copy-paste, not callable functions.

### 3. State is for Display, Not for the Agent
The `Model` in `model.go` holds state for the UI (selected tool, logs). The agent's internal reasoning state, conversation history, and plan are not separate, managed entities that persist across tool calls.

---

## The Refactor: Injecting the Agentic Brain

The TUI doesn't need to be scrapped. A new Agent Core needs to be inserted between the user's input and the tools. The TUI becomes the front-end for this core.

---

## Step 1: Create the Agent Core (Headless)

Create a new Go package, `agent/`, with a central `Engine` that runs the loop. This engine uses the existing `floydtools` via a new MCP Server wrapper.

```go
// agent/engine.go
package agent

type Engine struct {
    session   *SessionState
    mcpClient *MCPClient // Connects to tools
    llmClient *LLMClient // For GLM-4.7 API
}

func (e *Engine) RunTask(userGoal string) (<-chan Event, error) {
    // 1. GATHER CONTEXT: Read .floyd/ files, call cache_retrieve
    // 2. PLAN: Ask LLM with context + goal to generate a plan (list of tool calls)
    // 3. ACT: Execute each tool call via mcpClient
    // 4. VERIFY: Assess result, loop or continue
    // Stream events (PlanGenerated, ToolCalled, ResultReceived) to channel for TUI
}

type Event struct {
    Type string // "plan", "tool_call", "result", "error"
    Data interface{}
}
```

---

## Step 2: Wrap Tools in an MCP Server

The `floydtools` package is perfect. Now expose it via MCP so the Engine can call them. This is the most critical change.

```go
// mcp/server.go
package mcp

import (
    "github.com/Nomadcxx/sysc-Go/tui/floydtools"
)

// StartServer starts an MCP server that exposes floydtools
func StartServer() error {
    // For each tool in floydtools.AllTools(), register an MCP "tool"
    // Listen for JSON-RPC requests like:
    // {"method": "tools/call", "params": {"name": "bash", "arguments": {"input": "ls -la"}}}
    // Route to the appropriate tool's Run method
    // Return result
}
```

---

## Step 3: Refactor floyd_mode.go to Delegate to the Agent

Change the role of the TUI's FLOYD mode. Instead of running tools, it launches and monitors the agent.

**Current Flow:** User selects "Bash" tool → Prompts for command → Runs tool.

**New Agentic Flow:** User types goal in chat → `Engine.RunTask(goal)` → TUI listens to Event stream and displays:
```
[FLOYD] Planning...
[FLOYD] Executing: bash "ls -la"
[FLOYD] Result: (output)...
[FLOYD] Executing: read "package.json"
```

**Key Code Change in floyd_mode.go:**

```go
func (m Model) startFloydAction(kind FloydToolKind, value string) (Model, tea.Cmd) {
    // OLD: Run tool directly
    // NEW: If user input is a natural language goal (not a tool command):
    if isNaturalLanguageGoal(value) {
        // Start the Agent Engine
        return m.beginFloydStream("Task: "+value, func(ch chan<- floydStreamMsg){
            events := agentEngine.RunTask(value)
            for event := range events {
                ch <- convertEventToStreamMsg(event)
            }
        })
    }
    // Keep old tool-launching behavior as a fallback/advanced mode
}
```

---

## Step 4: Connect the Supercache as an MCP Tool

The cache must become a tool the agent can use. Implement the `CacheManager` from the blueprint as an MCP tool within the new server.

```go
// mcp/cache_tool.go
package mcp

func (s *Server) registerCacheTools() {
    s.RegisterTool("cache_store", func(params map[string]interface{}) (interface{}, error) {
        tier := params["tier"].(string)
        key := params["key"].(string)
        // ... call cache logic
    })
    s.RegisterTool("cache_retrieve", ...)
    s.RegisterTool("vault_search", ...)
}
```

---

## 2-Week Refactor Plan

| Week | Task | Outcome |
|------|------|---------|
| **Week 1** | Build the MCP Server & Agent Engine (new `agent/` and `mcp/` packages). Keep TUI unchanged. Test engine from simple CLI. | A headless agent runnable via `go run agent/cli.go "build a login form"` |
| **Week 2** | Integrate Engine into TUI. Refactor `floyd_mode.go` to use engine for natural language inputs. Keep old tool palette as "manual mode." | TUI gains true chat input. Typing "Add auth to my app" triggers full agentic workflow |
| **Week 3** | Connect Supercache. Implement cache MCP tools and update agent instructions to use them for context gathering. | Agent remembers past sessions and reuses solution patterns |
| **Week 4** | Add Chrome Integration. Connect Chrome extension as external MCP server the agent can call. | Agent can autonomously control browser to test UIs or gather data |

---

## The Bottom Line

An excellent tool platform and UI already exists. The missing piece is an autonomous agent that uses those tools on the user's behalf. By building a headless Agent Engine that uses tools via MCP and refactoring the TUI to be its interface, the "chatbot frame" transforms into the "terminal coding tool" envisioned.

---

**Last Updated:** 2026-01-12
**Context:** Architectural analysis for transforming FLOYD from manual tool launcher to agentic system
