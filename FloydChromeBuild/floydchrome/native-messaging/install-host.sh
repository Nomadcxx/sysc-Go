#!/bin/bash
# Install Native Messaging Host for FloydChrome
# This script registers the native messaging host with Chrome

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXTENSION_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Detect OS
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    HOST_MANIFEST_DIR="$HOME/Library/Application Support/Google/Chrome/NativeMessagingHosts"
    HOST_BINARY="$EXTENSION_DIR/native-messaging/floyd-chrome-host"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    HOST_MANIFEST_DIR="$HOME/.config/google-chrome/NativeMessagingHosts"
    HOST_BINARY="$EXTENSION_DIR/native-messaging/floyd-chrome-host"
else
    echo "Unsupported OS: $OSTYPE"
    exit 1
fi

# Create directory if it doesn't exist
mkdir -p "$HOST_MANIFEST_DIR"

# Get extension ID (user will need to install extension first and get ID)
echo "NOTE: You need to install the extension in Chrome first and get the Extension ID"
echo "You can find it at chrome://extensions/"
read -p "Enter Extension ID: " EXTENSION_ID

if [ -z "$EXTENSION_ID" ]; then
    echo "Error: Extension ID is required"
    exit 1
fi

# Create manifest with actual extension ID
MANIFEST_FILE="$HOST_MANIFEST_DIR/com.floyd.chrome.json"
cat > "$MANIFEST_FILE" <<EOF
{
  "name": "com.floyd.chrome",
  "description": "Native messaging host for FloydChrome extension",
  "path": "$HOST_BINARY",
  "type": "stdio",
  "allowed_origins": [
    "chrome-extension://$EXTENSION_ID/"
  ]
}
EOF

# Make host binary executable (if it exists)
if [ -f "$HOST_BINARY" ]; then
    chmod +x "$HOST_BINARY"
fi

echo "Native messaging host installed successfully!"
echo "Manifest: $MANIFEST_FILE"
echo ""
echo "Next steps:"
echo "1. Build the native host binary (floyd-chrome-host)"
echo "2. Install the Chrome extension"
echo "3. Start FLOYD CLI with --chrome flag"
