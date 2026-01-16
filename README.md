# FLOYD CLI

**File-Logged Orchestrator Yielding Deliverables**

A GLM-4.7 powered coding agent designed to compete with Claude Code - at a fraction of the cost.

## What is FLOYD?

FLOYD is an AI coding assistant that:
- Reads and writes code
- Runs commands and tools
- Remembers context across sessions (FLOYD-S SUPERCACHE)
- Follows the FLOYD AGENT protocol for structured development
- Uses your GLM Mac Code unlimited plan instead of monthly subscriptions

## Installation

```bash
# Build from source
go build -o floyd ./cmd/floyd
go build -o pink-floyd ./cmd/pink-floyd

# Install globally
sudo cp floyd /usr/local/bin/floyd
sudo cp pink-floyd /usr/local/bin/pink-floyd
```

## Configuration

FLOYD reads your API key from:
1. `ANTHROPIC_AUTH_TOKEN` environment variable
2. `GLM_API_KEY` environment variable
3. `ZHIPU_API_KEY` environment variable
4. `~/.claude/settings.json`

## Reliability Configuration

The following internal settings are configured for stability:
- **UI Watchdog:** 120s timeout (waits for agent thinking)
- **Tool Execution:** 5m timeout (allows long builds/tests)
- **Streaming:** No HTTP timeout (prevents disconnects during long generations)

## Usage

```bash
# Start FLOYD (TUI mode)
floyd

# Start pink-floyd (TUI mode)
pink-floyd

# Workspace commands (init .floyd/ directory)
floyd  # then type /init
```

## Commands

| Command | Description |
|---------|-------------|
| `/help` | Show help |
| `/init` | Initialize .floyd/ workspace |
| `/status` | Show workspace status |
| `/tools` | List available tools |
| `/clear` | Clear chat history |
| `/theme <name>` | Change theme |
| `/exit` | Quit |

## FLOYD-S SUPERCACHE

FLOYD uses a 3-tier caching system:

| Tier | Purpose | TTL |
|------|---------|-----|
| `reasoning` | Current conversation context | 5 min |
| `project` | Project-specific context | 24 hours |
| `vault` | Reusable solutions | 7 days |

## Available Tools

- `bash` - Run shell commands
- `read` - Read files
- `write` - Write files
- `edit` - Edit files (find/replace)
- `multiedit` - Multiple edits at once
- `grep` - Search files
- `ls` - List directories
- `glob` - Find files by pattern
- `cache` - Manage SUPERCACHE tiers

## Project Structure

```
├── agent/           # Core agent logic
├── cache/           # 3-tier cache backend
├── cmd/             # Entry points
│   ├── floyd/       # Main CLI/TUI
│   └── pink-floyd/  # TUI variant
├── tui/             # TUI components
└── ui/floyd/        # Shared UI package
```

## Status

- ✅ Supercache protocol installed
- ✅ 3-tier cache backend implemented
- ✅ Tools registered and working
- ✅ TUI mode working (fixed 2026-01-12)
- ✅ Agent responds to user input
- ⏳ Simple CLI mode in progress

## Quick Test

Verify the agent is working:

```bash
# Build and test
go build -o floyd ./cmd/floyd
go run ./cmd/test_agent/main.go

# Expected output:
# ✓ Client created
# ✓ Stream started
# Response: Hello! I am FLOYD.
# ✓ Done! Received X tokens
```

## Troubleshooting

If the agent doesn't respond:
1. Check API key: `echo $ANTHROPIC_AUTH_TOKEN` or `echo $GLM_API_KEY`
2. Run test: `go run ./cmd/test_agent/main.go`
3. See `docs/agents.md` for detailed troubleshooting

## License

MIT

## Acknowledgments

- Forked from [sysc-Go](https://github.com/Nomadcxx/sysc-Go)
- Inspired by [Claude Code](https://claude.ai/claude-code)
- Uses [Bubble Tea](https://github.com/charmbracelet/bubbletea) for TUI
