# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

sysc-Go is a terminal animation library for Go that provides multiple visual effects (fire, matrix rain, fireworks, rain, decrypt, pour, print, beams, skull) with customizable color themes. The project includes both a Go library and a CLI tool (`syscgo`).

## Build, Test, and Development Commands

### Building the CLI
```bash
# Build the CLI binary
go build -o syscgo ./cmd/syscgo/

# Build all packages
go build ./...

# Install CLI tool globally
go install github.com/Nomadcxx/sysc-Go/cmd/syscgo@latest
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific test
go test -run TestName ./path/to/package

# Run with verbose output
go test -v ./...
```

### Linting
```bash
# Run linter
golangci-lint run
```

### Running Effects
```bash
# Run with default settings (fire effect, dracula theme, 10 seconds)
./syscgo

# Run specific effect with theme
./syscgo -effect matrix -theme nord -duration 30

# Run decrypt effect with custom file
./syscgo -effect decrypt -theme tokyo-night -file message.txt -duration 15

# Run skull effect with ash particles
./syscgo -effect skull -theme rama -duration 30

# Run skull with text integration
./syscgo -effect skull-text -file art.txt -theme rama -duration 0

# Available effects: fire, matrix, rain, fireworks, decrypt, pour, print, beams, skull, skull-text
# Available themes: dracula, gruvbox, nord, tokyo-night, catppuccin, material, solarized, monochrome, transishardjob
```

## Architecture

### Project Structure
```
sysc-Go/
├── animations/          # Core animation library
│   ├── common.go       # Animation interface and common types
│   ├── palettes.go     # Theme color definitions
│   ├── fire.go         # DOOM PSX-style fire effect
│   ├── skull.go        # Skull drip with ash particles and accent illumination
│   ├── matrix.go       # Matrix digital rain effect
│   ├── fireworks.go    # Particle-based fireworks
│   ├── rain.go         # ASCII rain effect
│   ├── decrypt.go      # Text decryption animation
│   ├── pour.go         # Text pouring animation
│   ├── print.go        # Typewriter print effect
│   └── beams.go        # Light beam effect
├── cmd/syscgo/         # CLI application
│   └── main.go         # CLI entry point with effect runners
├── examples/           # Example usage code
└── docs/               # Documentation
```

### Core Concepts

**Animation Interface**: All effects implement the `Animation` interface:
```go
type Animation interface {
    Update()           // Advance animation by one frame
    Render() string    // Return current frame as string
    Reset()            // Restart animation
}
```

**Palette System**: Each effect has a `Get*Palette(theme string) []string` function that returns theme-specific hex colors. The palette system in `animations/palettes.go` centralizes all theme definitions.

**Effect Configuration**: Effects use dedicated config structs (e.g., `DecryptConfig`, `PourConfig`, `PrintConfig`) passed to constructors for complex initialization.

**Terminal Integration**: Effects render to strings that are printed with ANSI escape codes. The CLI handles terminal setup (cursor hiding, screen clearing) and frame timing (50ms delays for ~20fps).

**Styling**: All effects use `github.com/charmbracelet/lipgloss/v2` for color rendering and terminal styling.

### Animation Implementation Pattern

When implementing new effects:
1. Create effect struct with width, height, palette, and internal state
2. Implement `Animation` interface methods
3. Add palette function to `palettes.go` with theme support
4. Add effect runner to `cmd/syscgo/main.go`
5. Update CLI help text and switch statement
6. Handle terminal resizing if needed (implement `Resize(width, height int)`)

### Key Dependencies
- `github.com/charmbracelet/lipgloss/v2` - Terminal styling and colors
- `golang.org/x/term` - Terminal size detection
- `gonum.org/v1/gonum` - Mathematical operations for effects

## Effect-Specific Notes

**Fire Effect**: Uses PSX DOOM algorithm with heat propagation and random decay. Bottom row is heat source, flames spread upward with horizontal randomness.

**Matrix Effect**: Character streams fall at varying speeds with fade trails. Uses configurable character sets and brightness decay.

**Fireworks Effect**: Physics-based particle system with gravity, velocity, and particle lifetimes. Multiple fireworks spawn at intervals.

**Decrypt Effect**: Multi-phase animation (typing, decrypting, gradient) that cycles random characters before revealing final text.

**Pour Effect**: Text "pours" into position from configurable directions with physics-based movement and gradient transitions.

**Print Effect**: Typewriter-style rendering with customizable print head symbol and trailing characters.

**Skull Effect**: ASCII skull drips into place with gravity simulation. Features sequential accent illumination (top → eyes → cheekbones → teeth) and a two-layer ash particle system with wind drift and density breathing. The `skull-text` variant adds user text that slides in from edges with an easeOutBack bounce effect.

## Common Development Patterns

### Adding a New Theme
1. Add theme case to all palette functions in `animations/palettes.go`
2. Define 3-9 colors progressing from background to highlights
3. Ensure color progression makes sense for each effect type
4. Add theme to CLI help text in `cmd/syscgo/main.go`

### Frame Rate and Timing
- Target 20 FPS (50ms sleep between frames)
- Duration parameter: frames = duration * 20 (0 = infinite)
- Use `time.Sleep(50 * time.Millisecond)` consistently

### Terminal Handling
```go
// Setup
fmt.Print("\033[2J\033[H")   // Clear screen
fmt.Print("\033[?25l")       // Hide cursor

// Per frame
fmt.Print("\033[H")          // Move to top
fmt.Print(output)            // Render frame

// Cleanup
fmt.Print("\033[?25h")       // Show cursor
```

### Text Wrapping
The CLI includes a `wrapText(text, width)` function for handling text files that need to fit terminal width. Use this for text-based effects (decrypt, pour, print).

## Git Commit and Attribution Policy

**IMPORTANT**: When making commits to this repository:

1. **Commit only when explicitly requested** by the repository owner
2. **Author attribution MUST be**: 
   - Username: `Nomadcxx`
   - Email: `noovie@gmail.com`
3. **NEVER add AI/Claude attribution**:
   - No Co-authored-by tags mentioning AI/Claude
   - No commit messages referencing AI assistance
   - No tags or metadata that would add AI as a contributor
4. All commits must appear as if written directly by the repository owner
5. Pushing to remote is allowed as long as attribution rules are followed

## Code Style

- Use `PascalCase` for exported types/functions, `camelCase` for unexported
- Document all exported functions, types, and methods
- Group imports: standard library, third-party, internal
- Handle errors explicitly; wrap with context using `fmt.Errorf`
- Animation timing should be frame-based, not time-based
- Configuration should use dedicated structs, not function parameters
- Effects must gracefully handle small terminal sizes
