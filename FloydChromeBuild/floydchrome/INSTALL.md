# FloydChrome Installation Guide

## Quick Start

### 1. Load Extension in Chrome

1. Open Chrome and navigate to `chrome://extensions/`
2. Enable **Developer mode** (toggle in top right)
3. Click **"Load unpacked"**
4. Select the `floydchrome` directory
5. **Copy the Extension ID** (you'll need it for step 2)

### 2. Install Native Messaging Host

The native messaging host enables communication between FLOYD CLI and the extension.

#### macOS

```bash
cd floydchrome/native-messaging
./install-host.sh
# Enter your Extension ID when prompted
```

The script will:
- Create the native messaging host manifest
- Register it with Chrome
- Set up the connection path

#### Linux

Same process as macOS - the install script detects your OS automatically.

### 3. Build Native Host Binary

The native host binary (`floyd-chrome-host`) needs to be built as part of the FLOYD CLI integration. This binary:

- Receives messages from Chrome via stdin (JSON-RPC with 4-byte length prefix)
- Forwards to FLOYD CLI MCP server
- Returns responses back to Chrome

**Status**: Pending FLOYD CLI integration

### 4. Wire Up FLOYD Agent

The FLOYD agent stub is ready in `agent/floyd.js`. To wire up:

1. Replace the stub `processTask()` method with actual FLOYD agent calls
2. Update initialization to connect to your FLOYD agent instance
3. Handle agent responses and execute browser actions

## Verification

### Test MCP Connection

1. Start FLOYD CLI with `--chrome` flag:
   ```bash
   floyd --chrome
   ```

2. Open the extension side panel (click extension icon)
3. Check status indicator - should show "Connected to FLOYD CLI"

### Test Tools

From FLOYD CLI, try:
```
Navigate to https://example.com
Read the page
Click the "More information" link
```

## Troubleshooting

### Extension Not Loading

- Check `chrome://extensions/` for errors
- Verify manifest.json is valid JSON
- Ensure all required files exist

### Native Messaging Not Working

- Verify native host manifest is in correct location:
  - macOS: `~/Library/Application Support/Google/Chrome/NativeMessagingHosts/`
  - Linux: `~/.config/google-chrome/NativeMessagingHosts/`
- Check Extension ID matches in manifest
- Verify native host binary exists and is executable

### Debugger Attachment Errors

- Some tools require debugger attachment
- Only one debugger can attach per tab
- Extension handles this automatically, but conflicts may occur if DevTools is open

### Connection Issues

- Check Chrome console (background page) for MCP errors
- Verify FLOYD CLI is running with `--chrome` flag
- Check native host binary is running and accessible

## Next Steps

1. **Create Extension Icons**: Add `icon-16.png`, `icon-48.png`, `icon-128.png` to `assets/`
2. **Build Native Host**: Implement native host binary in FLOYD CLI
3. **Wire FLOYD Agent**: Connect actual FLOYD agent to `agent/floyd.js`
4. **Test Workflows**: Verify end-to-end browser automation
5. **Security Audit**: Review safety layer before production use

## Development

### File Structure

```
floydchrome/
├── manifest.json           # Extension config
├── background.js           # Service worker (MCP server)
├── content.js             # Content script (optional)
├── sidepanel/             # Standalone agent UI
├── tools/                 # Browser automation tools
├── mcp/                   # MCP protocol
├── safety/                # Security layer
├── agent/                 # FLOYD agent (STUB)
└── native-messaging/      # Native host setup
```

### Key Files

- `background.js`: Main entry point, initializes MCP server
- `mcp/server.js`: Handles JSON-RPC communication
- `tools/executor.js`: Routes tool calls to implementations
- `agent/floyd.js`: **STUB** - Wire up your FLOYD agent here

## Support

For issues or questions:
1. Check Chrome extension console: `chrome://extensions/` → "Service worker" → "Inspect views"
2. Check FLOYD CLI logs
3. Review native messaging host logs
