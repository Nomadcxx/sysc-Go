# FLOYD CLI - AI Coding Agent Architecture

**File-Logged Orchestrator Yielding Deliverables**

---

## Overview

FLOYD is an autonomous AI coding agent built in Go, designed to compete with Claude Code while leveraging the GLM-4.7 API through an Anthropic-compatible proxy. It provides a terminal-based interface for AI-assisted software development with full filesystem access, intelligent caching, and structured protocol management.

### Components

| Component | Status | Description |
|-----------|--------|-------------|
| **FLOYD CLI** | ✅ Complete | Terminal-based TUI agent |
| **FloydChrome** | ✅ Built | Chrome extension for browser automation |
| **3-Tier Cache** | ✅ Complete | SUPERCACHE with reasoning/project/vault tiers |
| **Protocol Manager** | ✅ Complete | Safety rules and context injection |

---

## Core Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         GLM-4.7 API                              │
│                    (Anthropic-Compatible)                        │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                        FLOYD AGENT                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Prompt    │  │    Tool     │  │    Protocol Manager     │  │
│  │   Loader    │  │   Loop      │  │  (Safety + Context)     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
         │                  │                      │
         ▼                  ▼                      ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────────────────┐
│  BASH         │  │  FILESYSTEM   │  │  3-TIER SUPERCACHE        │
│  - Commands   │  │  - read       │  │  ┌─────────────────────┐  │
│  - Builds     │  │  - write      │  │  │ Reasoning (5 min)   │  │
│  - Tests      │  │  - edit       │  │  ├─────────────────────┤  │
│  - Git        │  │  - grep       │  │  │ Project (24 hours)  │  │
│               │  │  - ls         │  │  ├─────────────────────┤  │
│               │  │  - glob       │  │  │ Vault (7 days)      │  │
└───────────────┘  └───────────────┘  └───────────────────────────┘
```

---

## Tool Capabilities

### 1. **Bash** - Shell Command Execution
Execute any shell command directly on the user's system.
```json
{"command": "go build ./...", "timeout": 30000}
```
**Use Cases:**
- Running builds and tests
- Git operations
- Package management (npm, pip, cargo)
- System diagnostics

### 2. **Read** - File Content Access
Read files with optional search and line limiting.
```json
{"file_path": "/path/to/file.go", "query": "func Error", "limit": 100}
```
**Use Cases:**
- Understanding codebases
- Reviewing configurations
- Finding specific functions

### 3. **Write** - File Creation/Overwrite
Create new files or completely replace existing ones.
```json
{"file_path": "/path/to/new_file.go", "content": "package main\n..."}
```
**Use Cases:**
- Creating new source files
- Writing configuration files
- Generating documentation

### 4. **Edit** - Surgical File Modification
Find and replace specific text in files.
```json
{"file_path": "main.go", "old_string": "func old()", "new_string": "func new()"}
```
**Use Cases:**
- Fixing bugs
- Renaming functions
- Updating imports

### 5. **MultiEdit** - Batch Modifications
Apply multiple non-contiguous edits in a single operation.
```json
{
  "file_path": "main.go",
  "edits": [
    {"old_string": "error1", "new_string": "fixed1"},
    {"old_string": "error2", "new_string": "fixed2"}
  ]
}
```

### 6. **Grep** - Pattern Search
Search across files using regex patterns.
```json
{"pattern": "TODO|FIXME", "path": ".", "glob": "*.go", "output_mode": "content"}
```
**Use Cases:**
- Finding code patterns
- Locating function definitions
- Tracking TODOs

### 7. **Ls** - Directory Listing
Explore project structure with smart filtering.
```json
{"path": ".", "ignore": [".git", "node_modules", "vendor"]}
```

### 8. **Glob** - File Pattern Matching
Find all files matching a glob pattern.
```json
{"pattern": "**/*.test.go"}
```

### 9. **Cache** - SUPERCACHE Management
Store and retrieve context across sessions.
```json
{"action": "store", "tier": "project", "key": "arch_decisions", "value": "..."}
```

---

## 3-Tier SUPERCACHE System

| Tier | Name | TTL | Purpose |
|------|------|-----|---------|
| 1 | **Reasoning** | 5 minutes | Current conversation context, working memory |
| 2 | **Project** | 24 hours | Project-specific context, decisions, progress |
| 3 | **Vault** | 7 days | Reusable solutions, patterns, learned approaches |

---

## Protocol Manager

### Safety Rules
- ❌ NEVER push to main/master
- ❌ NEVER delete files without confirmation
- ❌ NEVER run destructive commands without warning
- ✓ Use feature branches
- ✓ Prepare PR-ready changes

### Context Injection
FLOYD automatically loads from `.floyd/` directory:
- `master_plan.md` - Project goals and objectives
- `progress.md` - Execution log
- `stack.md` - Technology stack definition
- `scratchpad.md` - Error notes and working memory

### Runtime Context
Each request includes:
- Current time and timezone
- Working directory and repo name
- Git branch information
- Project type detection (Go, Node.js, Python, etc.)
- User and shell environment

---

## TUI Features

- **Streaming responses** - Real-time token display
- **Multiple themes** - Dark Side, Classic Pink, Neon Vapor, etc.
- **Tool visualization** - Shows tool calls and results
- **Progress indicators** - Animated thinking states
- **Session persistence** - Conversation history saved to disk

---

## Design Philosophy

1. **Execute, Don't Advise** - FLOYD acts on tasks rather than just describing solutions
2. **Verify Everything** - Builds and tests run after every change
3. **Context is King** - External memory in `.floyd/` provides persistent project knowledge
4. **Safety First** - Hard rules prevent destructive operations
5. **Professional Output** - Clean formatting with tables, boxes, and checkmarks

---

## FloydChrome Extension

Browser automation extension for FLOYD, enabling web research and interaction capabilities.

### Location
```
FloydChromeBuild/floydchrome/
```

### Architecture
```
┌─────────────────────────────────────────────────────────────────┐
│                     FLOYD CLI                                    │
│                  (Native Messaging)                              │
└─────────────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────────────┐
│                  FloydChrome Extension                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ Background  │  │    MCP      │  │      Safety Layer       │  │
│  │  Service    │  │   Server    │  │  (Sanitizer + Perms)    │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────────────┐
│                     Browser (Chrome)                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │  Navigate   │  │    Read     │  │       Interact          │  │
│  │   URLs      │  │   Pages     │  │    Click/Type           │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### Phase 1 Tools (8 Implemented)

| Tool | File | Description |
|------|------|-------------|
| `navigate` | tools/navigation.js | Navigate to URLs |
| `read_page` | tools/reading.js | Extract page text |
| `find_elements` | tools/reading.js | Query DOM elements |
| `click` | tools/interaction.js | Click elements |
| `type` | tools/interaction.js | Type into inputs |
| `get_tabs` | tools/tabs.js | List open tabs |
| `switch_tab` | tools/tabs.js | Switch active tab |
| `close_tab` | tools/tabs.js | Close a tab |

### Safety Features
- **Prompt Injection Protection** - `safety/sanitizer.js`
- **Permission Validation** - `safety/permissions.js`
- **Manifest V3 Compliant** - Modern Chrome extension standards

### Integration Points
1. **Native Messaging** - `native-messaging/` host manifests
2. **MCP Protocol** - `mcp/server.js` for FLOYD communication
3. **Agent Stub** - `agent/floyd.js` ready for wiring

### Installation
```bash
cd FloydChromeBuild/floydchrome
# Load as unpacked extension in Chrome
# Then run:
./native-messaging/install-host.sh <EXTENSION_ID>
```

---

## API Configuration

| Setting | Value |
|---------|-------|
| Endpoint | `https://api.z.ai/api/anthropic` |
| Model | `claude-opus-4` → GLM-4.7 |
| Format | Anthropic API compatible |
| Streaming | Supported |

---

## Quick Start

```bash
# Build
go build -o floyd ./cmd/floyd

# Run
./floyd

# Initialize workspace
/init

# Get help
/help
```

---

*FLOYD: Building complete software, not MVPs.*
