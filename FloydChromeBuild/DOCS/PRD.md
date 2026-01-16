# FloydChrome Product Requirements Document (PRD)

**Project:** FloydChrome - Browser Automation Extension for FLOYD Agent
**Status:** Planning / Design
**Version:** 1.0
**Date:** 2026-01-12
**Owner:** FLOYD Project

---

## Executive Summary

FloydChrome is a Chrome extension that enables the FLOYD AI coding agent to see, navigate, and interact with web pages. It serves as a browser automation bridge between the FLOYD CLI/TUI agent and Google Chrome, enabling "build-test-debug" workflows similar to Claude in Chrome, but powered by GLM-4.7 at a fraction of the cost.

### Key Differentiator vs. Standard Browser Automation
Unlike Playwright/Puppet which require *imperative scripting* (click here, type there), FloydChrome enables *agentic automation* where the AI interprets natural language goals and autonomously plans/executes browser actions.

---

## Problem Statement

Developers using FLOYD for coding tasks currently lack seamless browser automation capabilities. The existing workflow requires:
1. Writing code in the terminal with FLOYD
2. Manually switching to browser to test
3. Copy-pasting console errors back to FLOYD
4. Repeating the cycle

This context switching breaks flow and limits FLOYD's ability to autonomously verify its own code changes.

---

## Solution Overview

FloydChrome extends FLOYD's capabilities with browser automation through a **dual-mode architecture**:

| Mode | Description | Use Case |
|------|-------------|----------|
| **MCP Server Mode** | Extension exposes browser tools via Model Context Protocol; FLOYD CLI/TUI acts as client | Build-test-debug loops from terminal |
| **Standalone Agent Mode** | Extension runs its own AI agent in a Chrome side panel | Quick browser-only tasks without terminal |

---

## Core Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                    FLOYD CLI / TUI Agent                        │
│                  (GLM-4.7 powered coding agent)                  │
└──────────────────────────┬──────────────────────────────────────┘
                           │ MCP Protocol (JSON-RPC over stdio)
                           │ Native Messaging
┌──────────────────────────▼──────────────────────────────────────┐
│                   FloydChrome Extension                         │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │              Background Service Worker                     │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌───────────────────┐  │ │
│  │  │   MCP       │  │   Tool      │  │    Safety         │  │ │
│  │  │   Server    │  │   Executor  │  │    Layer          │  │ │
│  │  └─────────────┘  └─────────────┘  └───────────────────┘  │ │
│  └───────────────────────────────────────────────────────────┘ │
└──────────────────────────┬──────────────────────────────────────┘
                           │ Chrome Extension APIs
┌──────────────────────────▼──────────────────────────────────────┐
│                      Google Chrome                              │
│  (debugger, scripting, tabs, runtime, storage)                  │
└─────────────────────────────────────────────────────────────────┘
```

### Communication Protocol

**Native Messaging (Chrome) + MCP (Application Layer)**

1. FLOYD CLI starts with `--chrome` flag
2. CLI discovers FloydChrome's Native Messaging host
3. Bidirectional JSON-RPC channel established
4. FLOYD can now call browser tools: `read_page()`, `click()`, `type()`, etc.

---

## Tool Specifications

### Phase 1 Toolset (MVP)

| Tool | Description | Chrome API Used |
|------|-------------|-----------------|
| `navigate` | Navigate to URL | `chrome.tabs.update` |
| `read_page` | Get semantic accessibility tree | `chrome.debugger.sendCommand('DOM.getFlattenedDocument')` |
| `get_page_text` | Extract visible text content | `chrome.scripting.executeScript` |
| `find` | Locate element by natural language query | Accessibility tree search + fuzzy matching |
| `click` | Click element at coordinates | `chrome.debugger.sendCommand('Input.dispatchMouseEvent')` |
| `type` | Type text into focused element | `chrome.debugger.sendCommand('Input.insertText')` |
| `tabs_create` | Open new tab | `chrome.tabs.create` |
| `get_tabs` | List open tabs | `chrome.tabs.query` |

### Phase 2 Toolset (Enhanced)

| Tool | Description | Chrome API Used |
|------|-------------|-----------------|
| `read_console_messages` | Extract console logs/errors | `chrome.debugger.sendCommand('Runtime.consoleAPICalled')` |
| `read_network_requests` | Monitor network activity | `chrome.debugger.sendCommand('Network.getResponseBody')` |
| `screenshot` | Capture page screenshot | `chrome.tabs.captureVisibleTab` |
| `form_input` | Fill form fields intelligently | Form detection + `type` |
| `javascript_tool` | Execute arbitrary JS in page context | `chrome.scripting.executeScript` |
| `gif_creator` | Record workflow as animated GIF | `chrome.tabs.captureVisibleTab` (frame sequence) |

---

## Safety & Security

### Critical Injection Defenses

Given that FLOYD will interact with untrusted web content, the extension MUST implement:

1. **Prompt Injection Protection**
   - Treat all page content as potentially adversarial
   - Sanitize accessibility tree before sending to AI
   - Flag suspicious patterns (e.g., "ignore previous instructions")

2. **User Confirmation Requirements**
   - Destructive actions (form submission, payment) require user approval
   - Unknown domains trigger safety warning
   - Download actions must be confirmed

3. **Permission Minimization**
   - Request only necessary permissions
   - ActiveTab-only mode where possible
   - Clear permission manifest for user review

4. **Audit Logging**
   - All tool calls logged locally
   - User can review extension activity
   - Exportable session history

---

## Integration with FLOYD

### CLI Integration

```bash
# Start FLOYD with Chrome integration
floyd --chrome

# In-session usage
You: "Test the login form at localhost:3000"
FLOYD: [Navigates] [Reads form] [Fills credentials] [Submits] [Checks console for errors]
```

### TUI Integration

The FLOYD TUI (`pink-floyd`) will include:
- Browser status indicator (connected/disconnected)
- Tab selector dropdown
- Screenshot preview pane
- Console log viewer
- Tool execution log

---

## Technical Requirements

### Extension Structure

```
floydchrome/
├── manifest.json           # Extension configuration
├── background.js           # MCP server + tool executor
├── content.js              # Page context bridge (optional)
├── sidepanel/              # Standalone agent UI
│   ├── index.html
│   ├── panel.js
│   └── panel.css
├── tools/                  # Tool implementations
│   ├── navigation.js
│   ├── interaction.js
│   ├── reading.js
│   └── debugging.js
├── mcp/                    # MCP protocol handling
│   ├── server.js
│   └── messages.js
├── safety/                 # Security layer
│   ├── sanitizer.js
│   └── permissions.js
└── assets/
    └── icon.png
```

### Manifest Configuration

```json
{
  "manifest_version": 3,
  "name": "FloydChrome - Browser Automation for FLOYD",
  "version": "1.0.0",
  "description": "Connect FLOYD AI agent to your browser for automated testing and workflows",
  "permissions": [
    "nativeMessaging",
    "debugger",
    "scripting",
    "tabs",
    "activeTab",
    "storage"
  ],
  "host_permissions": [
    "<all_urls>"
  ],
  "background": {
    "service_worker": "background.js",
    "type": "module"
  },
  "side_panel": {
    "default_path": "sidepanel/index.html"
  },
  "externally_connectable": {
    "ids": ["*"]
  }
}
```

---

## Success Criteria

### Phase 1 (MVP)
- [ ] Extension installs in Chrome without errors
- [ ] Native Messaging connection to FLOYD CLI established
- [ ] Basic tools working: navigate, read_page, click, type
- [ ] FLOYD can complete a simple login flow autonomously
- [ ] Console errors can be read back to FLOYD

### Phase 2 (Enhanced)
- [ ] GIF recording for workflow documentation
- [ ] Network request inspection
- [ ] Form auto-fill with intelligent field detection
- [ ] Side panel standalone agent mode functional

### Phase 3 (Production)
- [ ] Published to Chrome Web Store
- [ ] Full test coverage
- [ ] Security audit completed
- [ ] Performance optimization (context efficiency)

---

## Known Limitations (Inherited from Claude in Chrome)

1. **Chrome-only**: No Brave/Arc/Edge support initially (Chromium compatibility possible later)
2. **No headless mode**: Requires visible browser window
3. **Context efficiency**: Full page snapshots can be large; smart summarization needed
4. **Beta status**: Expect edge cases and gradual stability improvements

---

## Dependencies

| Component | Version/Source | Purpose |
|-----------|----------------|---------|
| GLM-4.7 API | api.z.ai (Anthropic-compatible proxy) | AI model for agent reasoning |
| FLOYD CLI/TUI | This repo | Main agent interface |
| Chrome Extension API | Manifest V3 | Browser automation primitives |
| Model Context Protocol | anthropic-protocols/python | Communication standard |

---

## Open Questions

1. **Native Messaging Host Registration**: How will the host manifest be installed? (Auto-install script vs. manual)
2. **Context Window Management**: Should we implement on-device page summarization to reduce token usage?
3. **Multi-browser Support**: Is Chromium-wide compatibility a Phase 1 or Phase 2 goal?
4. **Standalone Agent LLM**: For side panel mode, which model? (GLM-4.7 via API or local model?)

---

## References

- [Claude in Chrome System Prompt](https://github.com/x1xhlol/system-prompts-and-models-of-ai-tools/blob/main/Anthropic/Claude%20for%20Chrome/Prompt.txt)
- [Claude in Chrome Tools Schema](https://github.com/x1xhlol/system-prompts-and-models-of-ai-tools/blob/main/Anthropic/Claude%20for%20Chrome/Tools.json)
- [Chrome Extension Docs](https://developer.chrome.com/docs/extensions/)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [Native Messaging](https://developer.chrome.com/docs/extensions/mv3/nativeMessaging/)

---

*This PRD is a living document. As development progresses, update success criteria, resolve open questions, and document architectural decisions.*
Build it