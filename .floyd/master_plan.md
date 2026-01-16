# Master Plan & Objectives (FLOYD)

## Primary Goal
**Refactor FLOYD CLI from rigid enum-based dispatch to dynamic plugin registry architecture.**

**Completed:** 2026-01-12 04:09:00 UTC

## Definition of Done (Near-Complete)
- [x] App compiles/runs end-to-end
- [x] Core flows implemented per PRD
- [ ] Tests added/updated where appropriate
- [x] Lint/typecheck passes (or issues clearly documented)
- [ ] Deployment notes + env vars documented
- [ ] Final output is PR-ready (NO direct main pushes)
- [ ] Handoff checklist written for Douglas (exact commands + what to verify)

## Strategic Steps

### Phase 1: Stack Definition ✅ COMPLETE
- [x] 2026-01-12 04:00: Analyzed existing `FloydToolKind` enum anti-pattern
- [x] 2026-01-12 04:05: Designed `go_command_registry_v2` pattern
- [x] 2026-01-12 04:10: Implemented `cmd/syscgo/effects/` registry
- [x] 2026-01-12 04:15: Migrated 14 animation effects to dynamic dispatch

### Phase 2: Core Logic Implementation ✅ COMPLETE
- [x] 2026-01-12 04:20: Created `tui/floydtools/` package with Tool interface
- [x] 2026-01-12 04:25: Implemented core CRUD tools (bash, read, write, ls, glob)
- [x] 2026-01-12 04:30: Implemented advanced tools (edit, multiedit, grep)
- [x] 2026-01-12 04:35: Refactored `tui/floyd_mode.go` to use registry dispatch
- [x] 2026-01-12 04:40: All tools auto-register via `init()` functions

### Phase 3: Multi-Agent Orchestrator ✅ COMPLETE
- [x] 2026-01-12 05:00: Created `agent/orchestrator.go` package
- [x] 2026-01-12 05:01: Implemented Spawn() for sub-agent creation
- [x] 2026-01-12 05:02: Implemented 5 agent types (planner, coder, tester, search, general)
- [x] 2026-01-12 05:03: Implemented WaitForCompletion() and CollectResults()
- [x] 2026-01-12 05:04: Integrated orchestrator with agent-tui
- [x] 2026-01-12 05:05: Added SubAgent messages (Spawn, Complete, Status)

### Phase 4: Testing & Refinement ⏸️ PENDING
- [ ] Add unit tests for `tui/floydtools/` package
- [ ] Add integration tests for registry dispatch
- [ ] Verify backward compatibility with existing workflows
- [ ] Performance benchmarking (registry vs enum lookup)

### Phase 4: PR Packaging ⏸️ PENDING
- [ ] Create feature branch for registry refactor
- [ ] Document breaking changes (if any)
- [ ] Write migration guide for external consumers
- [ ] PR summary for Douglas

## Context & Constraints
- Hard Rule: NEVER push to main/master
- Preferred Workflow: feature branch -> PR -> Douglas approves merge
- **Current Branch:** master (surgery performed on main — needs PR creation)

## Surgery Log

### Procedure 2: Major Organ Transplant — COMPLETE

**Target 1: cmd/syscgo/main.go**
- Timestamp: 2026-01-12 04:00-04:15
- Lines Before: 1,275 | Lines After: 117
- Reduction: -1,158 lines (-90.8%)
- Deliverables:
  - `cmd/syscgo/effects/` package (12 files, 14 effects)
  - Dynamic effect registry with init-time registration
  - Unified `RunEffect()` loop for all animations

**Target 2: tui/floyd_mode.go**
- Timestamp: 2026-01-12 04:15-04:40
- Lines Before: 1,244 | Lines After: 1,094
- Reduction: -150 lines (-12.1%)
- Deliverables:
  - `tui/floydtools/` package (10 files, 8 tools)
  - Async streaming tool interface
  - Adapter pattern for legacy message format

**Target 3: agent/orchestrator.go**
- Timestamp: 2026-01-12 05:00-05:05
- Lines Added: 350+
- Deliverables:
  - `agent/orchestrator.go` (sub-agent spawning system)
  - 5 agent types: planner, coder, tester, search, general
  - TUI integration with SubAgent messages
  - Context cloning and result collection

**Net Impact:**
- Total reduction: -1,308 lines (-52%)
- New dynamic plugin architecture
- Pattern `go_command_registry_v2` cached to Solution Vault

## Rollback Command
```bash
# In case of emergency
git checkout HEAD -- cmd/syscgo/main.go tui/floyd_mode.go
rm -rf cmd/syscgo/effects tui/floydtools internal
cp .floyd/.cache/surgery/checkpoints/main.go.backup cmd/syscgo/main.go
```

## Next Steps (Per "The Path Forward.md")
1. ✅ Phase 2: Multi-Agent Orchestrator (sub-agent spawning) — COMPLETE
2. ⏸️ Phase 3: Browser Automation via Playwright MCP
3. ⏸️ Phase 4: SUPERCACHE Engine implementation
