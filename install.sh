#!/bin/bash
# sysc-Go one-line installer
# Usage: curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-Go/master/install.sh | sudo bash

set -e

TEMP_DIR=""
cleanup() {
    if [ -n "$TEMP_DIR" ]; then
        cd /
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

echo "sysc-Go installer"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    echo "Usage: curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-Go/master/install.sh | sudo bash"
    exit 1
fi

# Check for Go
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Install Go first: https://go.dev/doc/install"
    exit 1
fi

# Check for git before cloning. A minimal image can have Go and not git.
if ! command -v git &> /dev/null; then
    echo "Error: git is not installed"
    echo "Install git first, then re-run this installer"
    exit 1
fi

# Create temp directory
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

echo "Cloning sysc-Go..."
git clone https://github.com/Nomadcxx/sysc-Go.git
cd sysc-Go

echo "Building installer..."
go build -o install-syscgo ./cmd/installer/

echo "Running installer..."
# --yes skips the Bubble Tea welcome screen. curl | bash has no TTY input.
./install-syscgo --yes

echo ""
echo "Installation complete."
echo "Try: syscgo -effect fire -theme dracula"
echo "  or: syscgo-tui"
