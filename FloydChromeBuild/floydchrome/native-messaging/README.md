# Native Messaging Host Setup

The Native Messaging host enables communication between the FLOYD CLI and the FloydChrome extension.

## Installation

### macOS

1. Install the Chrome extension and note the Extension ID from `chrome://extensions/`
2. Run the install script:
   ```bash
   ./install-host.sh
   ```
3. Build the native host binary (see below)

### Linux

Same as macOS, but paths will be different.

## Native Host Binary

The native host binary (`floyd-chrome-host`) is a bridge that:
- Receives messages from Chrome via stdin
- Forwards them to FLOYD CLI via MCP protocol
- Returns responses from FLOYD CLI back to Chrome

### Implementation Notes

The native host should:
1. Read JSON-RPC messages from stdin (4-byte length prefix + JSON)
2. Forward to FLOYD CLI MCP server
3. Return responses in same format

This will be implemented as part of the FLOYD CLI integration.

## Testing

To test the connection:
1. Install extension
2. Install native host manifest
3. Start FLOYD CLI with `--chrome` flag
4. Extension should show "Connected" status in side panel
