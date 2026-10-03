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

# sudo's secure_path is a short system PATH. It drops the invoking user's
# PATH, so a Go that works in their shell (official tarball at
# /usr/local/go/bin, ~/sdk/go/bin, asdf/mise/gvm/nix) is invisible and
# `command -v go` fails before clone/build. Rebuild a PATH for go and git
# only; sudoers and the script's own PATH stay unchanged.

invoker_home() {
    if [ -z "${SUDO_USER:-}" ] || [ "$SUDO_USER" = "root" ]; then
        return 0
    fi
    getent passwd "$SUDO_USER" | cut -d: -f6 || true
}

# Login shell so ~/.profile is read (https://go.dev/doc/install adds
# /usr/local/go/bin there). Non-interactive, so interactive bashrc hooks
# such as tmux are not started. stdin is /dev/null so `curl | sudo bash`
# is not consumed.
invoking_user_environment() {
    if [ -z "${SUDO_USER:-}" ] || [ "$SUDO_USER" = "root" ]; then
        return 0
    fi
    sudo -n -u "$SUDO_USER" -H bash -lc 'printf "__SYSC_PATH__%s\n" "$PATH"; printf "__SYSC_MISE_DATA_DIR__%s\n" "${MISE_DATA_DIR:-$HOME/.local/share/mise}"' </dev/null 2>/dev/null || true
}

add_go_dir() {
    local dir="$1"
    if [ -z "$dir" ] || [ ! -x "$dir/go" ]; then
        return 0
    fi
    GO_PREFIX="${GO_PREFIX:+$GO_PREFIX:}$dir"
}

gvm_gos_dir() {
    local home="$1"
    local bin
    if [ -z "$home" ] || [ ! -d "$home/.gvm/gos" ]; then
        return 0
    fi
    bin=$(find "$home/.gvm/gos" -path '*/bin/go' -type f -executable 2>/dev/null | sort -V | tail -n 1 || true)
    if [ -n "$bin" ]; then
        dirname "$bin"
    fi
}

# Directories the error text and common version managers actually use.
# Searched after the user's login PATH, before secure_path.
collect_go_dirs() {
    local home="$1"
    GO_PREFIX=""
    add_go_dir /usr/local/go/bin
    if [ -n "$home" ]; then
        add_go_dir "$home/sdk/go/bin"
        add_go_dir "$home/.local/share/mise/shims"
        add_go_dir "$home/.mise/shims"
        add_go_dir "$home/.asdf/shims"
        if [ -x "$home/.asdf/shims/go" ] && [ -x "$home/.asdf/bin/asdf" ]; then
            GO_PREFIX="${GO_PREFIX:+$GO_PREFIX:}$home/.asdf/bin"
        fi
        add_go_dir "$home/.nix-profile/bin"
        add_go_dir "$home/.gvm/bin"
        add_go_dir "$(gvm_gos_dir "$home")"
    fi
    printf '%s' "$GO_PREFIX"
}

# Drop empty components so a trailing ":" in the user PATH cannot mean ".".
sanitize_path() {
    local part rest="$1:" out=""
    while [ -n "$rest" ]; do
        part=${rest%%:*}
        rest=${rest#*:}
        if [ -n "$part" ]; then
            out="${out:+$out:}$part"
        fi
    done
    printf '%s' "$out"
}

# First executable file named "$1" on the colon-separated path "$2".
# Requires the target to be executable, so a dangling or non-executable
# symlink does not shadow a later toolchain.
first_executable() {
    local name="$1" dir cand rest="${2}:"
    while [ -n "$rest" ]; do
        dir=${rest%%:*}
        rest=${rest#*:}
        cand="$dir/$name"
        if [ -n "$dir" ] && [ -f "$cand" ] && [ -x "$cand" ]; then
            printf '%s\n' "$cand"
            return 0
        fi
    done
    return 1
}

INVOKER_HOME="$(invoker_home || true)"
INVOKER_ENV="$(invoking_user_environment)"
USER_PATH="$(sanitize_path "$(printf '%s\n' "$INVOKER_ENV" | sed -n 's/^__SYSC_PATH__//p' | tail -n 1)")"
USER_MISE_DATA_DIR="$(printf '%s\n' "$INVOKER_ENV" | sed -n 's/^__SYSC_MISE_DATA_DIR__//p' | tail -n 1)"
GO_DIRS="$(collect_go_dirs "$INVOKER_HOME")"
# Login PATH, then known install locations, then the sanitized secure_path.
CANDIDATE="$(sanitize_path "${USER_PATH:+$USER_PATH:}${GO_DIRS:+$GO_DIRS:}$PATH")"

GO_BIN="$(first_executable go "$CANDIDATE" || true)"
GIT_BIN="$(first_executable git "$CANDIDATE" || true)"

# Check for Go
if [ -z "$GO_BIN" ]; then
    echo "Error: Go is not installed or not visible on PATH"
    echo "sudo resets PATH (secure_path), which omits /usr/local/go/bin and version managers."
    echo "Install Go first: https://go.dev/doc/install"
    exit 1
fi

if [ -z "$GIT_BIN" ]; then
    echo "Error: git is not installed"
    echo "Install git first, then re-run this installer"
    exit 1
fi

# Directory of the resolved toolchain, plus the helper bin a shim execs.
# The script's own PATH (mktemp, rm) stays on secure_path.
toolchain_path() {
    local extra=""
    extra="$(dirname "$GO_BIN")"
    if [ -n "$INVOKER_HOME" ] && [ "$GO_BIN" = "$INVOKER_HOME/.asdf/shims/go" ] && [ -x "$INVOKER_HOME/.asdf/bin/asdf" ]; then
        extra="$extra:$INVOKER_HOME/.asdf/bin"
    fi
    if [ -n "$INVOKER_HOME" ] && [ -x "$INVOKER_HOME/.local/bin/mise" ]; then
        case "$GO_BIN" in
            "$INVOKER_HOME/.local/share/mise/shims/go"|"$INVOKER_HOME/.mise/shims/go")
                extra="$extra:$INVOKER_HOME/.local/bin"
                ;;
        esac
    fi
    sanitize_path "$extra:$PATH"
}

TOOL_PATH="$(toolchain_path)"

# Create temp directory
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

echo "Cloning sysc-Go..."
"$GIT_BIN" clone https://github.com/Nomadcxx/sysc-Go.git
cd sysc-Go

echo "Building installer..."
# PATH covers shim helpers (asdf, mise). The binary itself is absolute,
# so a different `go` earlier on PATH cannot replace it.
run_with_toolchain() {
    if [ -n "$INVOKER_HOME" ] && [ -d "$INVOKER_HOME/.asdf" ]; then
        ASDF_DIR="$INVOKER_HOME/.asdf" \
            ASDF_DATA_DIR="${ASDF_DATA_DIR:-$INVOKER_HOME/.asdf}" \
            MISE_DATA_DIR="${USER_MISE_DATA_DIR:-$INVOKER_HOME/.local/share/mise}" \
            PATH="$TOOL_PATH" "$@"
    elif [ -n "$USER_MISE_DATA_DIR" ]; then
        MISE_DATA_DIR="$USER_MISE_DATA_DIR" PATH="$TOOL_PATH" "$@"
    else
        PATH="$TOOL_PATH" "$@"
    fi
}
run_with_toolchain "$GO_BIN" build -o install-syscgo ./cmd/installer/

echo "Running installer..."
# --yes skips the Bubble Tea welcome screen. curl | bash has no TTY input.
run_with_toolchain ./install-syscgo --yes

echo ""
echo "Installation complete."
echo "Try: syscgo -effect fire -theme dracula"
echo "  or: syscgo-tui"
