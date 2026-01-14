# Technology Stack (FLOYD)

## Core Frameworks
- [x] Go 1.21+
- [x] Bubbletea TUI framework
- [x] Lipgloss styling
- [ ] GLM-4.7 API (for agent orchestration)

## Dynamic Registry Pattern
- [x] `go_command_registry_v2` pattern (validated)
- [x] Effect registry (cmd/syscgo/effects/)
- [x] Tool registry (tui/floydtools/)
- [x] Init-time auto-registration

## File Operations
- [x] Bash (shell execution)
- [x] Read (file reading)
- [x] Write (file writing)
- [x] Edit (string replacement)
- [x] MultiEdit (atomic multi-edit)
- [x] LS (directory listing)
- [x] Glob (pattern matching)
- [x] Grep (regex search via ripgrep)

## Animation System
- [x] Fire effects
- [x] Matrix effects
- [x] Rain effects
- [x] Fireworks effect
- [x] Pour effect
- [x] Print effect
- [x] Beams effects
- [x] Ring-text effect
- [x] Blackhole effect
- [x] Aquarium effect

## Browser Automation (Pending)
- [ ] Playwright MCP integration
- [ ] Chrome tool templates (placeholders ready)

## Sub-Agent System (Phase 2)
- [x] Agent orchestrator
- [x] Specialist spawning (planner, coder, tester)
- [x] Context cloning for sub-agents

## SUPERCACHE Engine (Phase 4)
- [x] .floyd/.cache/ directory structure
- [x] Reasoning frame persistence (cache/frame.go)
- [x] Project chronicle (cache/chronicle.go)
- [x] Solution vault indexing (cache/vault.go)
- [x] CacheManager tool (floydtools/cache.go)
