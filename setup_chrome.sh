#!/bin/bash

# Define paths
SOURCE_FILE="/Volumes/Storage/FLOYD_CLI/INK/floyd-cli/com.floyd.chrome.json"
DEST_DIR="$HOME/Library/Application Support/Google/Chrome/NativeMessagingHosts"
DEST_FILE="$DEST_DIR/com.floyd.chrome.json"

echo "🔧 Setting up FloydChrome Native Messaging Host..."

# Check if source exists
if [ ! -f "$SOURCE_FILE" ]; then
    echo "❌ Error: Source file not found at $SOURCE_FILE"
    exit 1
fi

# Create destination directory
echo "📂 Creating directory: $DEST_DIR"
mkdir -p "$DEST_DIR"

# Copy file
echo "Vk Copying configuration file..."
cp "$SOURCE_FILE" "$DEST_DIR/"

# Verify
if [ -f "$DEST_FILE" ]; then
    echo "✅ Success! The Native Messaging Host is installed."
    echo ""
    echo "👉 NEXT STEP:"
    echo "   Run the CLI command separately:"
    echo "   cd INK/floyd-cli && node dist/cli.js --chrome"
else
    echo "❌ Error: Failed to copy file."
    exit 1
fi
