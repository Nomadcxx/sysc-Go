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
# /usr/local/go/bin, a Nix profile) is invisible and `command -v go` fails
# before clone/build. Rebuild a PATH for go and git only; sudoers and the
# script's own PATH stay unchanged.
#
# That PATH is the user's, and this script runs as root. A git or go found
# there runs with EUID 0, so a shim the user (or anything running as them)
# can write would run code as root (#183). Only a tool that root alone can
# change is used: see trusted_path.

# Owners whose files root may execute. Tests widen it; nothing else should.
TRUSTED_UIDS="0"

invoker_home() {
    if [ -z "${SUDO_USER:-}" ] || [ "$SUDO_USER" = "root" ]; then
        return 0
    fi
    getent passwd "$SUDO_USER" | cut -d: -f6 || true
}

# Login shell so ~/.profile is read (https://go.dev/doc/install adds
# /usr/local/go/bin there). Non-interactive, so interactive bashrc hooks
# such as tmux are not started. stdin is /dev/null so `curl | sudo bash`
# is not consumed. It runs as the user, so it only yields search paths.
invoking_user_path() {
    if [ -z "${SUDO_USER:-}" ] || [ "$SUDO_USER" = "root" ]; then
        return 0
    fi
    sudo -n -u "$SUDO_USER" -H bash -lc 'printf "__SYSC_PATH__%s\n" "$PATH"' </dev/null 2>/dev/null || true
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

# True when "$1" is owned by a trusted uid and no one else can write it. A
# sticky directory such as /tmp is writable by others, but they cannot
# replace an entry they do not own, so it passes as a directory.
root_controlled() {
    local info owner mode
    info="$(stat -c '%u %a' -- "$1" 2>/dev/null)" || return 1
    owner=${info% *}
    mode=${info#* }
    case " $TRUSTED_UIDS " in
        *" $owner "*) ;;
        *) return 1 ;;
    esac
    if (( (8#$mode & 8#022) == 0 )); then
        return 0
    fi
    [ "$2" = dir ] && (( (8#$mode & 8#1000) != 0 ))
}

# Prints the fully resolved "$1" if it and every directory above it are
# root-controlled. The resolved path is what runs, so a symlink that is
# later repointed cannot swap the binary.
trusted_path() {
    local real dir
    real="$(readlink -f -- "$1")" || return 1
    root_controlled "$real" file || return 1
    dir="$real"
    while [ "$dir" != "/" ]; do
        dir="$(dirname -- "$dir")"
        root_controlled "$dir" dir || return 1
    done
    printf '%s\n' "$real"
}

# Sets FOUND to the resolved path of the first trusted executable named "$1"
# on the colon-separated path "$2", and SKIPPED when it passed one over.
# Requires the target to be executable, so a dangling or non-executable
# symlink does not shadow a later toolchain.
find_trusted() {
    local name="$1" dir cand real rest="${2}:"
    FOUND=""
    SKIPPED=""
    while [ -n "$rest" ]; do
        dir=${rest%%:*}
        rest=${rest#*:}
        cand="$dir/$name"
        if [ -z "$dir" ] || [ ! -f "$cand" ] || [ ! -x "$cand" ]; then
            continue
        fi
        if real="$(trusted_path "$cand")"; then
            FOUND="$real"
            return 0
        fi
        echo "Skipping $cand: it or a directory above it is not root-owned"
        SKIPPED=1
    done
    return 1
}

untrusted_help() {
    echo "This installer runs as root and only runs a $1 that only root can change."
    echo "Install $1 system-wide (a distribution package, or Go's tarball in /usr/local/go),"
    echo "or use the README's interactive installer, which runs the toolchain you choose."
}

INVOKER_HOME="$(invoker_home || true)"
USER_PATH="$(sanitize_path "$(invoking_user_path | sed -n 's/^__SYSC_PATH__//p' | tail -n 1)")"
GO_DIRS="/usr/local/go/bin${INVOKER_HOME:+:$INVOKER_HOME/.nix-profile/bin}"
# Login PATH, then known install locations, then the sanitized secure_path.
CANDIDATE="$(sanitize_path "${USER_PATH:+$USER_PATH:}$GO_DIRS:$PATH")"

find_trusted go "$CANDIDATE" || true
GO_BIN="$FOUND"
if [ -z "$GO_BIN" ]; then
    if [ -n "$SKIPPED" ]; then
        echo "Error: no Go that root can safely run"
        untrusted_help Go
    else
        echo "Error: Go is not installed or not visible on PATH"
        echo "sudo resets PATH (secure_path), which omits /usr/local/go/bin."
        echo "Install Go first: https://go.dev/doc/install"
    fi
    exit 1
fi

find_trusted git "$CANDIDATE" || true
GIT_BIN="$FOUND"
if [ -z "$GIT_BIN" ]; then
    if [ -n "$SKIPPED" ]; then
        echo "Error: no git that root can safely run"
        untrusted_help git
    else
        echo "Error: git is not installed"
        echo "Install git first, then re-run this installer"
    fi
    exit 1
fi

echo "Using $GO_BIN and $GIT_BIN"

# The installer looks go up on PATH again, so put the trusted toolchain
# first. The script's own PATH (mktemp, rm) stays on secure_path.
TOOL_PATH="$(sanitize_path "$(dirname -- "$GO_BIN"):$PATH")"

# Create temp directory
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

echo "Cloning sysc-Go..."
"$GIT_BIN" clone https://github.com/Nomadcxx/sysc-Go.git
cd sysc-Go

echo "Building installer..."
PATH="$TOOL_PATH" "$GO_BIN" build -o install-syscgo ./cmd/installer/

echo "Running installer..."
# --yes skips the Bubble Tea welcome screen. curl | bash has no TTY input.
PATH="$TOOL_PATH" ./install-syscgo --yes

echo ""
echo "Installation complete."
echo "Try: syscgo -effect fire -theme dracula"
echo "  or: syscgo-tui"
