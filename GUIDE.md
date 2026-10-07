# sysc-Go Developer Guide

Complete guide for using sysc-Go animation library in your Go applications.

sysc-Go provides two types of effects:
- **Animations**: Standalone effects that don't require text input (fire, matrix, rain, fireworks, beams, aquarium)
- **Text Effects**: Effects that animate ASCII text and art (fire-text, matrix-art, rain-art, pour, print, beam-text, ring-text, blackhole)

## Table of Contents
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Available Animations](#available-animations)
- [Color Themes](#color-themes)
- [Integration Examples](#integration-examples)
- [API Reference](#api-reference)

## Quick Start

```bash
# Install the library
go get github.com/Nomadcxx/sysc-Go

# Try the CLI tool
go install github.com/Nomadcxx/sysc-Go/cmd/syscgo@latest
syscgo -effect fire -theme dracula

# Or use the interactive TUI
go install github.com/Nomadcxx/sysc-Go/cmd/syscgo-tui@latest
syscgo-tui
```

## Installation

Add to your project:

```bash
go get github.com/Nomadcxx/sysc-Go
```

Import in your code:

```go
import "github.com/Nomadcxx/sysc-Go/animations"
```

## Basic Usage

### Fire Effect

The classic DOOM PSX-style fire animation.

```go
package main

import (
    "fmt"
    "time"
    "github.com/Nomadcxx/sysc-Go/animations"
)

func main() {
    width, height := 80, 24
    
    // Get color palette for your theme
    palette := animations.GetFirePalette("dracula")
    
    // Create fire effect
    fire := animations.NewFireEffect(width, height, palette)
    
    // Animation loop
    for frame := 0; frame < 200; frame++ {
        fire.Update(frame)
        output := fire.Render()
        
        fmt.Print("\033[H")  // Move cursor to top
        fmt.Print(output)
        time.Sleep(50 * time.Millisecond)
    }
}
```

### Matrix Rain

Digital rain effect with falling character streaks.

```go
package main

import (
    "fmt"
    "time"
    "github.com/Nomadcxx/sysc-Go/animations"
)

func main() {
    width, height := 80, 24
    
    palette := animations.GetMatrixPalette("nord")
    matrix := animations.NewMatrixEffect(width, height, palette)
    
    for frame := 0; frame < 200; frame++ {
        matrix.Update(frame)
        output := matrix.Render()
        
        fmt.Print("\033[H")
        fmt.Print(output)
        time.Sleep(50 * time.Millisecond)
    }
}
```

### Fireworks

Physics-based particle fireworks display.

```go
package main

import (
    "fmt"
    "time"
    "github.com/Nomadcxx/sysc-Go/animations"
)

func main() {
    width, height := 80, 24
    
    palette := animations.GetFireworksPalette("gruvbox")
    fireworks := animations.NewFireworksEffect(width, height, palette)
    
    for frame := 0; frame < 200; frame++ {
        fireworks.Update(frame)
        output := fireworks.Render()
        
        fmt.Print("\033[H")
        fmt.Print(output)
        time.Sleep(50 * time.Millisecond)
    }
}
```

### ASCII Rain

Character-based rain effect.

```go
package main

import (
    "fmt"
    "time"
    "github.com/Nomadcxx/sysc-Go/animations"
)

func main() {
    width, height := 80, 24
    
    palette := animations.GetRainPalette("tokyo-night")
    rain := animations.NewRainEffect(width, height, palette)
    
    for frame := 0; frame < 200; frame++ {
        rain.Update(frame)
        output := rain.Render()

        fmt.Print("\033[H")
        fmt.Print(output)
        time.Sleep(50 * time.Millisecond)
    }
}
```

### Ring Text

Spectacular rotating text convergence animation.

```go
package main

import (
    "fmt"
    "time"
    "github.com/Nomadcxx/sysc-Go/animations"
)

func main() {
    width, height := 100, 30
    text := "SYSC-GO"

    config := animations.RingTextConfig{
        Width:               width,
        Height:              height,
        Text:                text,
        RingColors:          []string{"#bd93f9", "#ff79c6", "#f1fa8c"},
        FinalGradientStops:  []string{"#6272a4", "#bd93f9", "#f8f8f2"},
        FinalGradientSteps:  12,
        FinalGradientDir:    animations.GradientHorizontal,
        RotationFrames:      120,
        ConvergenceFrames:   80,
    }

    ringText := animations.NewRingTextEffect(config)

    for {
        ringText.Update()
        output := ringText.Render()

        fmt.Print("\033[H")
        fmt.Print(output)
        time.Sleep(50 * time.Millisecond)
    }
}
```

### Fire Text

ASCII text consumed by rising flames.

```go
package main

import (
    "fmt"
    "os"
    "time"
    "github.com/Nomadcxx/sysc-Go/animations"
)

func main() {
    width, height := 120, 40

    // Read ASCII art from file
    artData, _ := os.ReadFile("logo.txt")
    artText := string(artData)

    config := animations.FireTextConfig{
        Width:         width,
        Height:        height,
        Text:          artText,
        FireColors:    []string{"#f92672", "#fd971f", "#f4bf75", "#ffffe0"},
        TextColor:     "#f8f8f2",
        BurnSpeed:     2,
        FlameHeight:   8,
    }

    fireText := animations.NewFireTextEffect(config)

    for {
        fireText.Update()
        output := fireText.Render()

        fmt.Print("\033[H")
        fmt.Print(output)
        time.Sleep(50 * time.Millisecond)
    }
}
```

### Rain Art

ASCII art with crystallizing rain effect.

```go
package main

import (
    "fmt"
    "os"
    "time"
    "github.com/Nomadcxx/sysc-Go/animations"
)

func main() {
    width, height := 120, 40

    // Read ASCII art from file
    artData, _ := os.ReadFile("logo.txt")
    artText := string(artData)

    config := animations.RainArtConfig{
        Width:          width,
        Height:         height,
        AsciiArt:       artText,
        RainColor:      "#7aa2f7",
        FrozenColors:   []string{"#7aa2f7", "#bb9af7", "#7dcfff"},
        FreezeChance:   0.9,  // 90% freeze rate for fast crystallization
        MaxDrops:       width * 4,
        SpawnRate:      0.5,
    }

    rainArt := animations.NewRainArtEffect(config)

    for {
        rainArt.Update()
        output := rainArt.Render()

        fmt.Print("\033[H")
        fmt.Print(output)
        time.Sleep(50 * time.Millisecond)
    }
}
```

## Available Animations

### Animations (Standalone)
Effects that don't require text input.

#### Fire Effect
- **Constructor**: `NewFireEffect(width, height int, palette []string) *FireEffect`
- **Palette Function**: `GetFirePalette(theme string) []string`
- **Methods**:
  - `Update(frame int)` - Advance animation
  - `Render() string` - Get current frame
  - `Resize(width, height int)` - Change dimensions
  - `UpdatePalette(palette []string)` - Change colors

#### Logo Spin Family
- **Constructors**: `NewLogoSpinEffect(config LogoSpinConfig) *LogoSpinAnimation`
- **Shapes**: `"sysc"` (block SYSC wordmark), `"cross"` (Latin cross), `"justice"` (faceted justice-cross wireframe), `"sysc-cross"` / `"cross-sysc"` (morph between two while spinning)
- **Styles**: `"wire"` (default, glowing contour rings + extruded edges), `"solid"` (filled beveled plate)
- **Palette Function**: `GetLogoPalette(theme string) []string`
- **Registry IDs**: `sysc-logo`, `cross-logo`, `justice-cross`, `logo-morph`
- **Methods**:
  - `Update()` - Advance one frame (one rotation every 144 frames)
  - `Render() string` - Current frame as truecolor braille
  - `Reset()` - Restart from frame zero
  - `Resize(width, height int)` - Change dimensions
- **Notes**: vector strokes are baked to a signed distance field once per shape,
  extruded into a 3D surface, and rasterized through a braille z-buffer. The
  `justice` shape is defined by its 16-vertex outline loop plus three inner
  facet lines, matching the source wireframe SVG.

#### Matrix Effect
- **Constructor**: `NewMatrixEffect(width, height int, palette []string) *MatrixEffect`
- **Palette Function**: `GetMatrixPalette(theme string) []string`
- **Methods**:
  - `Update(frame int)` - Advance animation
  - `Render() string` - Get current frame
  - `Resize(width, height int)` - Change dimensions

#### Rain Effect
- **Constructor**: `NewRainEffect(width, height int, palette []string) *RainEffect`
- **Palette Function**: `GetRainPalette(theme string) []string`
- **Methods**:
  - `Update(frame int)` - Advance animation
  - `Render() string` - Get current frame

#### Fireworks Effect
- **Constructor**: `NewFireworksEffect(width, height int, palette []string) *FireworksEffect`
- **Palette Function**: `GetFireworksPalette(theme string) []string`
- **Methods**:
  - `Update(frame int)` - Advance animation
  - `Render() string` - Get current frame
  - `Resize(width, height int)` - Change dimensions

#### Beams Effect
Full-screen light beam background animation with sweeping beams.
- **Constructor**: `NewBeamsEffect(config BeamsConfig) *BeamsEffect`
- **Methods**:
  - `Update()` - Advance animation
  - `Render() string` - Get current frame

#### Aquarium Effect
Underwater scene with swimming fish, diver, boat, mermaid, and sea life.
- **Constructor**: `NewAquariumEffect(config AquariumConfig) *AquariumEffect`
- **Methods**:
  - `Update()` - Advance animation
  - `Render() string` - Get current frame

### Text Effects
Effects that require text/ASCII art input.

#### Fire Text Effect
ASCII text consumed by rising flames.
- **Constructor**: `NewFireTextEffect(config FireTextConfig) *FireTextEffect`
- **Methods**:
  - `Update()` - Advance animation
  - `Render() string` - Get current frame
  - `Reset()` - Restart animation

#### Matrix Art Effect
ASCII art revealed by Matrix-style digital streams with 90% freeze rate.
- **Constructor**: `NewMatrixArtEffect(config MatrixArtConfig) *MatrixArtEffect`
- **Methods**:
  - `Update()` - Advance animation
  - `Render() string` - Get current frame

#### Rain Art Effect
ASCII art with crystallizing rain effect - drops fall and freeze into art.
- **Constructor**: `NewRainArtEffect(config RainArtConfig) *RainArtEffect`
- **Methods**:
  - `Update()` - Advance animation
  - `Render() string` - Get current frame

#### Pour Effect
Characters pour into position from different directions (top, bottom, left, right).
- **Constructor**: `NewPourEffect(config PourConfig) *PourEffect`
- **Methods**:
  - `Update()` - Advance animation
  - `Render() string` - Get current frame
  - `Reset()` - Restart animation
  - `Resize(width, height int)` - Change dimensions

#### Print Effect
Classic typewriter-style text rendering with cursor.
- **Constructor**: `NewPrintEffect(config PrintConfig) *PrintEffect`
- **Methods**:
  - `Update()` - Advance animation
  - `Render() string` - Get current frame
  - `Reset()` - Restart animation
  - `Resize(width, height int)` - Change dimensions

#### Beam Text Effect
Text display with animated light beams, auto-sizing, and display mode.
- **Constructor**: `NewBeamTextEffect(config BeamTextConfig) *BeamTextEffect`
- **Methods**:
  - `Update()` - Advance animation
  - `Render() string` - Get current frame
  - `Reset()` - Restart animation
  - `IsComplete() bool` - Check if animation finished

#### Ring Text Effect
Spectacular animation where text rotates and converges into position.
- **Constructor**: `NewRingTextEffect(config RingTextConfig) *RingTextEffect`
- **Methods**:
  - `Update()` - Advance animation
  - `Render() string` - Get current frame
  - `Reset()` - Restart animation

#### Blackhole Effect
Text gets consumed by a swirling blackhole, collapses, and explodes outward.
- **Constructor**: `NewBlackholeEffect(config BlackholeConfig) *BlackholeEffect`
- **Methods**:
  - `Update()` - Advance animation
  - `Render() string` - Get current frame
  - `Reset()` - Restart animation

## Color Themes

All animations support these themes:

| Theme | Description | Style |
|-------|-------------|-------|
| `dracula` | Purple and pink vampiric vibes | Dark, vibrant |
| `gruvbox` | Retro warm colors | Warm, earthy |
| `nord` | Cool arctic palette | Cool, calm |
| `tokyo-night` | Neon Tokyo nights | Dark, neon |
| `catppuccin` | Soothing pastel tones | Soft, pastel |
| `material` | Google Material colors | Clean, modern |
| `solarized` | Classic precision colors | Balanced |
| `monochrome` | Grayscale aesthetic | Minimal |
| `transishardjob` | Trans pride colors | Pink, blue, white |
| `rama` | RAMA keyboard aesthetics | Red, gray, white |
| `eldritch` | Cosmic horror palette | Green, cyan, purple |
| `dark` | Pure black and white minimalism | Monochrome, stark |

Each effect has its own palette function:
- `GetFirePalette(theme)`
- `GetMatrixPalette(theme)`
- `GetFireworksPalette(theme)`
- `GetRainPalette(theme)`

## Integration Examples

### Terminal Size Detection

Get actual terminal dimensions:

```go
import (
    "os"
    "golang.org/x/term"
)

func getTerminalSize() (int, int) {
    width, height, err := term.GetSize(int(os.Stdout.Fd()))
    if err != nil {
        return 80, 24  // Fallback
    }
    return width, height
}
```

### Clean Terminal Setup

Proper terminal setup for animations:

```go
import "fmt"

func setupTerminal() {
    fmt.Print("\033[2J")   // Clear screen
    fmt.Print("\033[H")    // Move cursor to top
    fmt.Print("\033[?25l") // Hide cursor
}

func restoreTerminal() {
    fmt.Print("\033[?25h") // Show cursor
}

func main() {
    setupTerminal()
    defer restoreTerminal()
    
    // Your animation loop here
}
```

### Theme Switching

Switch themes dynamically:

```go
fire := animations.NewFireEffect(80, 24, animations.GetFirePalette("dracula"))

// Switch to gruvbox after 100 frames
for frame := 0; frame < 200; frame++ {
    if frame == 100 {
        fire.UpdatePalette(animations.GetFirePalette("gruvbox"))
    }
    
    fire.Update(frame)
    fmt.Print("\033[H" + fire.Render())
    time.Sleep(50 * time.Millisecond)
}
```

### Window Resize Handling

Handle terminal resize events:

```go
import (
    "os"
    "os/signal"
    "syscall"
    "golang.org/x/term"
)

func main() {
    width, height := getTerminalSize()
    fire := animations.NewFireEffect(width, height, animations.GetFirePalette("dracula"))
    
    // Listen for resize signals
    sigwinch := make(chan os.Signal, 1)
    signal.Notify(sigwinch, syscall.SIGWINCH)
    
    go func() {
        for range sigwinch {
            w, h := getTerminalSize()
            fire.Resize(w, h)
        }
    }()
    
    // Animation loop...
}
```

### With Bubble Tea

Integration with Bubble Tea TUI framework:

```go
package main

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/Nomadcxx/sysc-Go/animations"
    "time"
)

type model struct {
    fire  *animations.FireEffect
    frame int
}

type tickMsg time.Time

func tick() tea.Cmd {
    return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

func (m model) Init() tea.Cmd {
    return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
    case tickMsg:
        m.frame++
        return m, tick()
    case tea.WindowSizeMsg:
        m.fire.Resize(msg.Width, msg.Height)
    }
    return m, nil
}

func (m model) View() string {
    m.fire.Update(m.frame)
    return m.fire.Render()
}

func main() {
    palette := animations.GetFirePalette("dracula")
    fire := animations.NewFireEffect(80, 24, palette)
    
    p := tea.NewProgram(model{fire: fire, frame: 0})
    p.Run()
}
```

## API Reference

### Common Types

```go
type Animation interface {
    Update()
    Render() string
    Reset()
}

type Config struct {
    Width  int
    Height int
    Theme  string
}
```

### Performance Tips

1. **Frame Rate**: 20 FPS (50ms delay) is optimal for most animations
2. **Terminal Size**: Larger terminals need more CPU - consider throttling
3. **Color Depth**: Some terminals handle RGB better than others
4. **Buffer Management**: Animations manage their own buffers efficiently

### Troubleshooting

**Animation looks corrupted:**
- Ensure terminal supports RGB colors
- Try a different theme
- Check terminal size is correct

**Performance issues:**
- Reduce frame rate (increase sleep time)
- Use smaller terminal dimensions
- Switch to simpler animation (rain vs fireworks)

**Colors not showing:**
- Verify terminal supports 24-bit color
- Try `COLORTERM=truecolor` environment variable

## Examples Directory

Check `examples/simple/` for complete working examples:
- `fire.go` - Basic fire effect
- More examples coming soon

## Contributing

Found a bug or want to add an animation? PRs welcome at:
https://github.com/Nomadcxx/sysc-Go

## License

MIT
