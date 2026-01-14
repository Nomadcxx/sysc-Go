# Changelog

## [1.0.0] - 2026-01-12

### Added
- Initial release of FloydChrome extension
- MCP Server implementation for FLOYD CLI integration
- Phase 1 toolset (MVP):
  - `navigate` - Navigate to URLs
  - `read_page` - Get accessibility tree
  - `get_page_text` - Extract visible text
  - `find` - Find elements by natural language
  - `click` - Click elements
  - `type` - Type text
  - `tabs_create` - Create new tabs
  - `get_tabs` - List open tabs
- Safety layer with prompt injection protection
- Side panel UI for standalone agent mode
- Native messaging host setup scripts
- FLOYD agent integration stub (ready for wiring)
- Audit logging for all tool calls
- Content sanitization before sending to AI

### Pending
- Native host binary implementation (FLOYD CLI integration)
- FLOYD agent wiring (stub ready)
- Extension icons (placeholder needed)
- Phase 2 tools (console, network, screenshots, etc.)

### Known Issues
- Native messaging requires native host binary (pending FLOYD CLI)
- Extension icons need to be created
- Some tools may conflict with Chrome DevTools debugger attachment
