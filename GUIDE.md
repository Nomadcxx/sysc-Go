# sysc-Go Developer Guide

Complete guide for using sysc-Go animation library in your Go applications.

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

## Available Animations

### Fire Effect
- **Constructor**: `NewFireEffect(width, height int, palette []string) *FireEffect`
- **Palette Function**: `GetFirePalette(theme string) []string`
- **Methods**:
  - `Update(frame int)` - Advance animation
  - `Render() string` - Get current frame
  - `Resize(width, height int)` - Change dimensions
  - `UpdatePalette(palette []string)` - Change colors

### Matrix Effect  
- **Constructor**: `NewMatrixEffect(width, height int, palette []string) *MatrixEffect`
- **Palette Function**: `GetMatrixPalette(theme string) []string`
- **Methods**:
  - `Update(frame int)` - Advance animation
  - `Render() string` - Get current frame
  - `Resize(width, height int)` - Change dimensions

### Fireworks Effect
- **Constructor**: `NewFireworksEffect(width, height int, palette []string) *FireworksEffect`
- **Palette Function**: `GetFireworksPalette(theme string) []string`
- **Methods**:
  - `Update(frame int)` - Advance animation
  - `Render() string` - Get current frame
  - `Resize(width, height int)` - Change dimensions

### Rain Effect
- **Constructor**: `NewRainEffect(width, height int, palette []string) *RainEffect`
- **Palette Function**: `GetRainPalette(theme string) []string`
- **Methods**:
  - `Update(frame int)` - Advance animation
  - `Render() string` - Get current frame

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
