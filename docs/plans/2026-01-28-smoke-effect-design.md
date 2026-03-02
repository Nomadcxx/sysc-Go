# Smoke Effect Design

**Date:** 2026-01-28
**Effect Names:** `smoke`, `smoke-text`
**Reference:** Terminal Text Effects smoke.py
**Default Theme:** RAMA

## Overview

A terminal animation effect featuring animated smoke that floods across the canvas using a breadth-first algorithm. The smoke effect has a breathing cycle (flood in → hold → recede → repeat) with phase-based color transitions. The `smoke-text` variant reveals text character-by-character as smoke passes over it, with characters transitioning through smoke gradients before revealing in final colors.

## Visual Style

**Smoke Animation:**
- Symbols: `░ ▒ ▓ ▒ ░` (light → medium → heavy → medium → light shading)
- Slow cycling: 15-20 frames per symbol for meditative atmosphere
- Breadth-first flood from random origin points
- Cool colors during flood-in, warm colors during recede

**Text Reveal (smoke-text only):**
- Text starts in theme's darkest/muted color
- Smoke floods with smart padding (5-10 chars around text)
- Characters transition through smoke gradient, then to final gradient colors
- Final text displayed with horizontal gradient using brightest theme colors

## Animation Architecture

### Effect Structure

```go
type SmokePhase int
const (
    PhaseFloodIn SmokePhase = iota
    PhaseHold
    PhaseRecede
    PhasePause
)

type SmokeCell struct {
    x, y          int
    symbolIndex   int      // Current position in symbol cycle (0-4)
    symbolFrame   int      // Frames on current symbol
    floodStep     int      // When this cell was activated
    active        bool     // Is this cell currently smoking
    phase         SmokePhase
}

type SmokeAnimation struct {
    width, height int
    palette       []string
    theme         string
    withText      bool
    textContent   string

    // Smoke state
    smokeCells    []SmokeCell
    floodMap      [][]int      // BFS flood order
    originX, originY int        // Current flood origin

    // Text state (smoke-text only)
    textChars     []TextChar
    textBounds    BoundingBox

    // Phase tracking
    phase         SmokePhase
    frameCount    int
    cycleFrame    int          // Frame within current cycle
}
```

### Dual Mode Operation

**Background Effect (smoke):**
```bash
syscgo -effect smoke -theme rama -duration 30
```

**Text Integration (smoke-text):**
```bash
syscgo -effect smoke-text -file SYSC.txt -theme rama -duration 0
```

Follows existing sysc-go pattern: `fire`/`fire-text`, `matrix`/`matrix-art`, `rain`/`rain-art`

## Phase Details

### Smoke Background Effect - Medium Breathing Cycles

**Phase 1: Flood In (4-5 seconds / 80-100 frames)**
- Smoke floods from random origin point using breadth-first algorithm
- Each frame, smoke expands to neighboring cells based on flood order
- Smoke symbols (`░ ▒ ▓ ▒ ░`) cycle slowly (15-20 frames per symbol)
- **Colors:** Cool/darker colors from theme palette (indices 1-3)
- Smoke gradually fills entire canvas

**Phase 2: Hold (2-3 seconds / 40-60 frames)**
- All smoke cells active and cycling symbols in place
- No expansion or movement, just animated symbol cycling
- Creates moment of full atmospheric density

**Phase 3: Recede (4-5 seconds / 80-100 frames)**
- Smoke disappears in reverse flood order (last flooded = first to vanish)
- **Colors:** Warm/brighter colors from theme palette (indices 4-6)
- Creates color shift from cool to warm during breathing cycle

**Phase 4: Pause (1-2 seconds / 20-40 frames)**
- Brief pause before next cycle
- New random origin selected
- Flood map regenerated

**Total cycle:** ~10-12 seconds, continuous loop

### Smoke Text Effect - Partial Loop

**Phase 1: Flood & Reveal (4-5 seconds / 80-100 frames)**
- Text starts in theme's darkest/most muted color (palette[0])
- Smoke floods with smart padding (5-10 chars around text boundaries)
- As smoke reaches each character:
  - Character transitions through smoke gradient (cool colors, 20-30 frames)
  - Then animates to final gradient color (30-40 frames)
- Horizontal gradient applied across revealed text

**Phase 2: Hold (15-20 seconds / 300-400 frames)**
- Text fully revealed in final gradient colors
- Smoke symbols continue cycling over revealed text
- Extended hold so users can read/appreciate the text

**Phase 3: Recede & Fade (4-5 seconds / 80-100 frames)**
- Smoke recedes in reverse order (warm colors)
- Text fades through smoke gradient back to starting muted color
- Synchronized with smoke disappearing

**Phase 4: Pause (1-2 seconds / 20-40 frames)**
- Brief pause before next cycle
- New random origin selected

**Total cycle:** ~25-30 seconds for full loop

## Breadth-First Flood Algorithm

### Flood Map Generation

**Initialization:**
- Select random origin point (x, y) on canvas
- Use breadth-first search (BFS) to calculate flood order for all cells
- Each cell gets a "flood step" number indicating when it will be reached
- Store in 2D flood map for quick lookup during animation

**BFS Algorithm:**
```go
func (s *SmokeAnimation) generateFloodMap(originX, originY int) {
    queue := []FloodPoint{{originX, originY, 0}}
    visited := make(map[Coord]bool)
    s.floodMap = make([][]int, s.height)

    for y := 0; y < s.height; y++ {
        s.floodMap[y] = make([]int, s.width)
        for x := 0; x < s.width; x++ {
            s.floodMap[y][x] = -1 // Not reached
        }
    }

    for len(queue) > 0 {
        point := queue[0]
        queue = queue[1:]

        coord := Coord{point.x, point.y}
        if visited[coord] {
            continue
        }

        s.floodMap[point.y][point.x] = point.step
        visited[coord] = true

        // Add neighbors (up, down, left, right)
        neighbors := []Coord{
            {point.x, point.y - 1}, // up
            {point.x, point.y + 1}, // down
            {point.x - 1, point.y}, // left
            {point.x + 1, point.y}, // right
        }

        for _, n := range neighbors {
            if n.x >= 0 && n.x < s.width && n.y >= 0 && n.y < s.height {
                if !visited[n] {
                    queue = append(queue, FloodPoint{n.x, n.y, point.step + 1})
                }
            }
        }
    }
}
```

**Flood Progression:**
- Frame-by-frame, activate cells whose `floodMap[y][x] <= currentFloodStep`
- `currentFloodStep` increments each frame (or every N frames for speed control)
- Creates smooth wave-like expansion from origin

**Reverse Order (Recede Phase):**
- Find maximum flood step in map
- Count down from max to 0
- Deactivate cells whose `floodMap[y][x] >= currentRecedingStep`
- Creates symmetric reverse flooding effect

### Smoke Cell Animation

**Symbol Cycling:**
- Symbols: `░ ▒ ▓ ▒ ░` (indices 0-4)
- Each symbol displays for 15-20 frames before advancing
- Cycle wraps: after index 4 → back to 0
- Each cell cycles independently (no synchronization needed)

**Color Selection:**
- Flood phase: cool colors from palette (indices 1-3)
- Recede phase: warm colors from palette (indices 4-6)
- Interpolate between colors based on flood depth/distance from origin

## Theme Integration & Color System

### Palette Function

**GetSmokePalette Structure:**
Returns 8 colors for comprehensive theming:

```go
// Returns: [muted_text, cool1, cool2, cool3, warm1, warm2, warm3, bright_final]
// Index 0: Starting text color (muted/dark)
// Indices 1-3: Cool colors for flood-in phase
// Indices 4-6: Warm colors for recede phase
// Index 7: Brightest accent for final text
```

**Example Palettes:**

```go
case "rama":
    return []string{
        "#2b2d42",  // Muted starting text
        "#8d99ae",  // Cool gray (flood)
        "#5a7fa6",  // Cool blue-gray (flood)
        "#3d5a7f",  // Deeper cool (flood)
        "#d90429",  // Warm red (recede)
        "#ef233c",  // Brighter red (recede)
        "#ff6b7a",  // Light warm (recede)
        "#edf2f4",  // Brightest white (final text)
    }

case "dracula":
    return []string{
        "#282a36",  // Background (muted text)
        "#44475a",  // Current line (cool)
        "#6272a4",  // Comment (cool)
        "#8be9fd",  // Cyan (cool)
        "#ffb86c",  // Orange (warm)
        "#ff79c6",  // Pink (warm)
        "#ff5555",  // Red (warm)
        "#f8f8f2",  // Foreground (brightest)
    }
```

Support all existing themes: dracula, gruvbox, nord, tokyo-night, catppuccin, material, solarized, monochrome, transishardjob, rama, eldritch, dark

### Color Application

**Smoke Background Effect:**
- **Flood phase:** Smoke cells use palette[1-3] (cool colors)
- **Hold phase:** Maintain flood colors
- **Recede phase:** Smoke cells transition to palette[4-6] (warm colors)
- Gradient within each phase based on flood depth/distance from origin

**Smoke Text Effect:**
- **Starting state:** Text rendered in palette[0] (muted, barely visible)
- **During flood:** Characters transition through palette[1-3] as smoke passes (20-30 frames)
- **Final reveal:** Text uses palette[7] + brightest 2-3 colors for horizontal gradient (30-40 frames)
- **During recede:** Text fades through palette[4-6] back to palette[0]

## Text Integration & Smart Padding

### Text Parsing & Positioning

**Text Layout:**
- Parse input text into character positions (like existing text effects)
- Center text horizontally and vertically in terminal
- Store character positions with initial state (muted color)

**Smart Boundary Detection:**

```go
type BoundingBox struct {
    minX, maxX int
    minY, maxY int
}

func (s *SmokeAnimation) calculateTextBounds() BoundingBox {
    // Find min/max X and Y of text characters
    bounds := calculateBounds(s.textChars)

    // Add 5-10 character padding
    return BoundingBox{
        minX: max(0, bounds.minX - 7),
        maxX: min(s.width-1, bounds.maxX + 7),
        minY: max(0, bounds.minY - 5),
        maxY: min(s.height-1, bounds.maxY + 5),
    }
}

func (s *SmokeAnimation) isInFloodBounds(x, y int) bool {
    return x >= s.textBounds.minX && x <= s.textBounds.maxX &&
           y >= s.textBounds.minY && y <= s.textBounds.maxY
}
```

Flood algorithm only activates cells within padded boundaries, creating effect of smoke coming from outside the text area.

### Character Reveal Mechanics

**Three-Stage Color Transition:**

1. **Pre-reveal (until smoke reaches):**
   - Character in starting color (palette[0])
   - Character is visible but muted

2. **Smoke transition (20-30 frames as smoke passes):**
   - Character cycles through smoke gradient (cool colors palette[1-3])
   - Creates "smoking" effect on the character itself
   - Example: palette[1] (10 frames) → palette[2] (10 frames) → palette[3] (10 frames)

3. **Final reveal (30-40 frame transition):**
   - Character smoothly transitions to final gradient color
   - Gradient color based on character's X position (horizontal gradient)
   - Smoke symbols may fade out or remain as subtle overlay

**Transition Implementation:**

```go
func (s *SmokeAnimation) getCharacterColor(char TextChar, floodStep int) string {
    charFloodStep := s.floodMap[char.y][char.x]

    if charFloodStep < 0 || floodStep < charFloodStep {
        // Not yet reached by smoke
        return s.palette[0] // Muted starting color
    }

    framesSinceReached := floodStep - charFloodStep

    if framesSinceReached < 30 {
        // Smoke gradient phase (0-30 frames)
        progress := float64(framesSinceReached) / 30.0
        return interpolateSmokGradient(progress, s.palette[1:4])
    } else if framesSinceReached < 60 {
        // Final reveal phase (30-60 frames)
        progress := float64(framesSinceReached-30) / 30.0
        finalColor := s.getTextGradientColor(char.x)
        return interpolateToFinal(s.palette[3], finalColor, progress)
    } else {
        // Fully revealed
        return s.getTextGradientColor(char.x)
    }
}
```

**Text Final Gradient:**

Horizontal gradient using brightest colors:
```go
// For RAMA theme
finalGradient := []string{
    palette[7],    // #edf2f4 (brightest)
    palette[6],    // #ff6b7a (warm bright)
    palette[4],    // #d90429 (accent red)
}
```

Map character X position to gradient color (left to right).

### Reverse Process (Recede Phase)

- Characters fade from final gradient to warm smoke colors (palette[4-6])
- Then fade to starting muted color (palette[0])
- Synchronized with smoke receding in reverse flood order
- Timing: reverse of reveal (60 frames total, 30 for warm transition, 30 for fade to muted)

## Implementation Details

### File Structure

**Files to Create:**
- `animations/smoke.go` - Core smoke animation
- Add `GetSmokePalette(theme string)` to `animations/palettes.go`

**CLI Integration (`cmd/syscgo/main.go`):**
- `runSmoke(width, height, theme, frames)` - Background effect
- `runSmokeText(width, height, theme, file, frames)` - Text reveal
- Add to effect switch statement
- Update help text

### Core Implementation Components

**Flood Map Generation:**
- `generateFloodMap(originX, originY)` - BFS calculation
- Called at start of each cycle with new random origin
- Returns 2D array of flood step numbers
- Cached for reuse during recede phase (iterate in reverse)

**Smoke Cell Management:**
- `updateSmokeCells()` - Activate/deactivate cells based on phase and flood progress
- `updateSymbolCycling()` - Advance symbol animation for active cells
- Manage active cell list for efficient rendering

**Phase State Machine:**
```go
func (s *SmokeAnimation) Update() {
    s.frameCount++
    s.cycleFrame++

    switch s.phase {
    case PhaseFloodIn:
        s.updateFloodIn()
    case PhaseHold:
        s.updateHold()
    case PhaseRecede:
        s.updateRecede()
    case PhasePause:
        s.updatePause()
    }

    s.updateSymbolCycling()
}
```

**Rendering Pipeline:**
1. Create buffer (width × height rune array)
2. Render smoke cells as background layer with phase-appropriate colors
3. Render text characters (if smoke-text mode) with current transition colors
4. Use lipgloss for color styling
5. Build final frame string

### Performance Considerations

**Flood Map Caching:**
- Only regenerate flood map when starting new cycle
- Reuse during flood/recede phases (just iterate in reverse)
- Prevents per-frame BFS recalculation

**Active Cell Management:**
- Only update cells that are currently active (flooded but not receded)
- Typical active count: 200-800 cells (manageable at 20 FPS)
- Use slice for active cells, not map for iteration efficiency

**Memory Usage:**
- Flood map: `width * height * 4 bytes` (~8KB for 80x24 terminal)
- Smoke cells: ~800 cells * 32 bytes = ~25KB
- Text characters: ~500 chars * 40 bytes = ~20KB
- Total: ~50KB, minimal memory footprint

### Terminal Compatibility

- Handle small terminal sizes gracefully
- Implement `Resize(width, height int)` to regenerate flood map
- Ensure smoke symbols render correctly across terminals

## Success Criteria

- [ ] Smoke floods smoothly with breadth-first expansion
- [ ] Symbol cycling creates animated smoke effect (15-20 frames/symbol)
- [ ] Cool-to-warm color shift during breathing cycle
- [ ] Text reveals character-by-character with smooth transitions (smoke-text)
- [ ] Smart padding creates natural reveal boundaries
- [ ] All themes supported with appropriate cool/warm colors
- [ ] Performance maintains 20 FPS target (200-800 active cells)
- [ ] Seamless cycle transitions with new random origins
- [ ] Reverse flood during recede creates symmetrical effect
- [ ] Compatible with both CLI and TUI

## Configuration Philosophy

Following YAGNI principle, keep configuration minimal:
- Effect name determines behavior (`smoke` vs `smoke-text`)
- Theme determines all colors (via GetSmokePalette)
- File input for text (smoke-text only)
- Duration for how long to run

No additional flags needed. If users want customization, they can modify the code or create custom themes.
