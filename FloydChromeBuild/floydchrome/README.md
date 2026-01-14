# FloydChrome Extension

Browser automation extension for FLOYD AI coding agent. Enables FLOYD to see, navigate, and interact with web pages through Chrome.

## Architecture

- **MCP Server Mode**: Extension exposes browser tools via Model Context Protocol; FLOYD CLI/TUI acts as client
- **Standalone Agent Mode**: Extension runs its own AI agent in a Chrome side panel

## Installation

### 1. Load Extension in Chrome

1. Open Chrome and navigate to `chrome://extensions/`
2. Enable "Developer mode" (toggle in top right)
3. Click "Load unpacked"
4. Select the `floydchrome` directory
5. Note the Extension ID (you'll need it for native messaging setup)

### 2. Install Native Messaging Host

```bash
cd native-messaging
./install-host.sh
# Enter your Extension ID when prompted
```

### 3. Build Native Host Binary

The native host binary (`floyd-chrome-host`) needs to be built as part of the FLOYD CLI integration. This will bridge Chrome Native Messaging to the FLOYD CLI MCP server.

## Usage

### MCP Server Mode (FLOYD CLI)

Start FLOYD CLI with Chrome integration:
```bash
floyd --chrome
```

FLOYD can now use browser tools:
- `navigate` - Navigate to URL
- `read_page` - Get page accessibility tree
- `get_page_text` - Extract visible text
- `find` - Find elements by natural language
- `click` - Click elements
- `type` - Type text
- `tabs_create` - Open new tab
- `get_tabs` - List open tabs

### Standalone Agent Mode

1. Click the FloydChrome extension icon
2. Side panel opens with agent interface
3. Enter tasks in natural language
4. Agent executes browser automation (when FLOYD agent is wired up)

## Development

### Project Structure

```
floydchrome/
├── manifest.json           # Extension configuration
├── background.js           # MCP server + tool executor
├── content.js              # Page context bridge (optional)
├── sidepanel/              # Standalone agent UI
├── tools/                  # Tool implementations
├── mcp/                    # MCP protocol handling
├── safety/                 # Security layer
├── agent/                  # FLOYD agent integration (STUB)
└── native-messaging/       # Native messaging host setup
```

### FLOYD Agent Integration

The FLOYD agent is currently stubbed in `agent/floyd.js`. To wire up:

1. Replace the stub implementation with actual FLOYD agent calls
2. Update `processTask()` to call the real agent
3. Handle agent responses and execute browser actions

## Safety Features

- **Prompt Injection Protection**: Sanitizes all page content before sending to AI
- **User Confirmation**: Destructive actions require approval
- **Audit Logging**: All tool calls logged locally
- **Permission Minimization**: Only requests necessary permissions

## Status

- ✅ Phase 1 MVP tools implemented
- ✅ MCP server infrastructure
- ✅ Safety layer
- ✅ Side panel UI
- ⏳ Native host binary (pending FLOYD CLI integration)
- ⏳ FLOYD agent wiring (stub ready)

## License

Part of the FLOYD project.
