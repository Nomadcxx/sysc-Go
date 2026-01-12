![sysc-Go](assets/sysc-go.png)

Terminal animation library for Go. Pure Go animations ready to use in your TUI applications.

![Matrix Rain Showcase](assets/matrix.gif)

## FLOYD Agent TUI

FLOYD is a Bubble Tea-based terminal UI for interacting with GLM-4.7 (via GLM Coding Plan). It provides an async chat interface with streaming responses and vim-style input.

### Installation

```bash
go install github.com/Nomadcxx/sysc-Go/cmd/agent-tui@latest
```

### Configuration

FLOYD reads your API key from `~/.claude/settings.json`:

```json
{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "your-api-key-here"
  }
}
```

Or set the `ANTHROPIC_AUTH_TOKEN` environment variable.

### Usage

```bash
# Start the agent TUI
agent-tui

# With a specific theme
agent-tui --theme catppuccin

# With a specific API key
agent-tui --api-key your-key-here
```

### Keybindings

| Key | Action |
|-----|--------|
| `esc` | Switch to normal mode |
| `i` | Switch to insert mode (from normal) |
| `enter` | Send message |
| `ctrl+j/k` | Navigate input history |
| `ctrl+l` | Clear conversation |
| `ctrl+r` | Refresh viewport |
| `ctrl+s` | Toggle auto-scroll |
| `ctrl+c` | Quit (or cancel current operation) |
| `q` | Quit (from normal mode) |

### Features

- **Streaming responses** - Real-time token streaming from GLM-4.7
- **Reasoning mode** - GLM-4.7 shows thought process before responding
- **Vim-style input** - Normal/insert modes with vim keybindings
- **Conversation history** - Persistent chat history with scrollback
- **Auto-scroll** - Automatically follows new responses
- **Multiple themes** - catppuccin, dracula, gruvbox, nord, tokyo-night, and more

### Model

FLOYD uses the GLM Coding Plan via the Anthropic-compatible API endpoint:
- Endpoint: `https://api.z.ai/api/anthropic`
- Model: `claude-opus-4` → maps to GLM-4.7
- Format: Anthropic API with streaming support

## Effects

### Animations
Standalone effects that don't require text input.

- **Fire** - DOOM PSX-style fire animation
- **Matrix Rain** - Classic Matrix digital rain
- **Rain** - ASCII character rain effect
- **Fireworks** - Particle-based fireworks display
- **Beams** - Full-screen light beam background animation
- **Aquarium** - Underwater scene with fish, diver, boat, and sea life

### Text Effects
Effects that animate ASCII text and art (requires `-file` flag).

- **Fire Text** - ASCII text consumed by rising flames
- **Matrix Art** - ASCII art with Matrix-style digital streams
- **Rain Art** - ASCII art with crystallizing rain effect
- **Pour** - Characters pour into position from different directions
- **Print** - Typewriter-style text rendering
- **Beam Text** - Text display with animated light beams and auto-sizing
- **Ring Text** - Text rotates and converges in spectacular ring animation
- **Blackhole** - Text gets consumed by a swirling blackhole and explodes

## Installation

### CLI Tools

**Via one-line installer (fastest):**
```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-Go/master/install.sh | sudo bash
```

**Via interactive installer:**
```bash
# Clone and run the TUI installer (requires sudo for system-wide installation)
# Installs both syscgo and syscgo-tui binaries
git clone https://github.com/Nomadcxx/sysc-Go.git
cd sysc-Go
sudo go run ./cmd/installer/
```

**Via AUR (Arch Linux):**
```bash
yay -S syscgo
```

**Via Go:**
```bash
go install github.com/Nomadcxx/sysc-Go/cmd/syscgo@latest
go install github.com/Nomadcxx/sysc-Go/cmd/syscgo-tui@latest
```

> **Note:** Installing via `go install` does not include fonts or assets. The TUI will work but the BIT editor font browser will be empty. For full functionality, use the installer or AUR package, or manually download fonts from `fonts/` to `~/.local/share/syscgo/fonts/`.

### As Library

```bash
go get github.com/Nomadcxx/sysc-Go
```

## Quick Start

### Interactive TUI

The easiest way to browse and select animations is through the interactive TUI:

```bash
syscgo-tui
```

![TUI Showcase](assets/showcase_tui.gif)

**Features:**
- Visual selector for all animations, themes, and durations
- Built-in ASCII art editor (BIT) with live preview
- 174 block-style fonts for creating ASCII art
- Real-time animation preview
- Export ASCII art to file (Ctrl+S)
- Navigate with arrow keys or vim keybindings (h/j/k/l)
- Instant theme switching

### Command Line

Run any animation directly from command line:

**Animations** (no text input required):
```bash
# Fire effect with Dracula theme (infinite loop)
syscgo -effect fire -theme dracula -duration 0

# Matrix rain with Nord theme for 30 seconds
syscgo -effect matrix -theme nord -duration 30

# Rain effect with Tokyo Night theme
syscgo -effect rain -theme tokyo-night

# Fireworks display
syscgo -effect fireworks -theme catppuccin -duration 20

# Beams effect (full-screen background)
syscgo -effect beams -theme nord -duration 0

# Aquarium effect (infinite)
syscgo -effect aquarium -theme dracula -duration 0
```

**Text Effects** (require `-file` flag with text/ASCII art):
```bash
# Fire text effect with ASCII art
syscgo -effect fire-text -file logo.txt -theme rama -duration 0

# Matrix art effect
syscgo -effect matrix-art -file art.txt -theme eldritch -duration 20

# Rain art effect
syscgo -effect rain-art -file banner.txt -theme tokyo-night -duration 15

# Pour effect with text file
syscgo -effect pour -file message.txt -theme gruvbox -duration 10

# Print effect (typewriter style)
syscgo -effect print -file banner.txt -theme catppuccin -duration 15

# Beam text effect with auto-sizing and display mode
syscgo -effect beam-text -file header.txt -auto -display -theme nord -duration 5

# Ring text effect
syscgo -effect ring-text -file title.txt -theme eldritch -duration 10

# Blackhole effect with text
syscgo -effect blackhole -file text.txt -theme dark -duration 15
```

**Available themes:** dracula, gruvbox, nord, tokyo-night, catppuccin, material, solarized, monochrome, transishardjob, rama, eldritch, dark

**Text Effect Flags:**
- `-auto` - Auto-size canvas to fit text (beam-text only)
- `-display` - Complete once and hold at final state (beam-text only)
- `-file` - Path to text file for text-based effects

## Asset Directories

When using text effects with `-file`, syscgo searches for files in the following order:

1. **Absolute path** - If you provide a full path (e.g., `/home/user/art.txt`)
2. **Current directory** - Files in your working directory
3. **User assets** - `~/.local/share/syscgo/assets/` (where new art is saved)
4. **System assets** - `/usr/local/share/syscgo/assets/` or `/usr/share/syscgo/assets/`

**Where your ASCII art is saved:**

When you create ASCII art using the TUI (syscgo-tui) and export it, files are saved to:

```
~/.local/share/syscgo/assets/
```

**Example usage:**
```bash
# Create art in TUI
syscgo-tui
# Use BIT editor, export as "hyprland.txt"

# Use your exported art
syscgo -effect fire-text -file hyprland.txt -theme rama -duration 0
```

For sysc-walls integration, art exports to `~/.config/sysc-walls/ascii/` instead.

## Effect Showcase

### Animations

#### Fire
![Fire Effect](assets/fire.gif)

#### Matrix Rain
![Matrix Rain](assets/matrix.gif)

#### Rain
![ASCII Rain](assets/rain.gif)

#### Fireworks
![Fireworks](assets/fireworks.gif)

#### Beams
![Beams Effect](assets/beams.gif)

#### Aquarium
![Aquarium Effect](assets/aquarium.gif)

### Text Effects

#### Fire Text
![Fire Text Effect](assets/fire-text.gif)

#### Matrix Art
![Matrix Art Effect](assets/matrix.gif)

#### Rain Art
![Rain Art Effect](assets/rain-text.gif)

#### Beam Text
![Beam Text Effect](assets/beam-text.gif)

#### Ring Text
![Ring Text Effect](assets/ring-text.gif)

#### Blackhole
![Blackhole Effect](assets/blackhole.gif)

## Demo

Run the interactive demo to see all animations:

```bash
cd examples/demo
go run .
```

## Documentation

See [GUIDE.md](GUIDE.md) for detailed CLI usage.

## Acknowledgements

- [terminaltexteffects](https://github.com/ChrisBuilds/terminaltexteffects) - Inspiration for terminal visual effects
- [bit](https://github.com/superstarryeyes/bit) - ASCII art editor and font library integration

## License

MIT