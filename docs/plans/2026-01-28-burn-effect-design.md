# Burn Effect Design

**Date:** 2026-01-28
**Effect Names:** `burn`, `burn-text`
**Reference:** Terminal Text Effects effect_burn.py
**Default Theme:** RAMA

## Overview

A terminal animation effect featuring progressive character destruction through fire. Characters ignite and burn away using spanning tree propagation for organic fire spread. The effect includes animated flame morphing, extended color gradients, and rising smoke particles. The `burn-text` variant burns user-provided text with complex character morphing, while `burn` fills the canvas with random characters that burn away.

## Visual Style

**Burning Animation:**
- **Background (`burn`)**: Simplified flame progression using `▁ ▂ ▃ ▄ ▅ ▆ ▇ █` (vertical fill)
- **Text (`burn-text`)**: Complex morph sequence `' . ▖ ▙ █ ▜ ▀ ▝ .` simulating flame consumption
- **Color Gradient**: 7-9 color progression (white hot → gold → orange → red → dark ember)
- **Propagation**: Prim's spanning tree algorithm creates connected, organic fire spread
- **Smoke**: Simplified emission from all burning characters, rises and fades

**Final State:**
- Characters burn away completely to empty space
- Brief pause before reset and new burn cycle

## Animation Architecture

### Effect Structure

```go
type BurnPhase int
const (
    PhaseIgnition BurnPhase = iota  // Sequential ignition
    PhaseBurning                    // Full burning
    PhaseSettle                     // Disappearance
    PhasePause                      // Brief rest before reset
)

type BurnChar struct {
    char          rune
    x, y          int
    ignitionFrame int     // When fire reached this char (-1 = not yet)
    burnStage     int     // Current stage in burn sequence
    colorIndex    int     // Current color in gradient
    stageFrame    int     // Frames on current stage
    colorFrame    int     // Frames on current color
    active        bool    // Still burning vs disappeared
}

type TreeEdge struct {
    from, to int     // Character indices
    weight   float64 // For Prim's algorithm
}

type SmokeParticle struct {
    x, y      float64
    char      rune
    velocity  float64  // Rise speed (0.5-0.8)
    opacity   float64  // 1.0 to 0.0
    lifetime  int      // Frames alive
    maxLife   int      // Frames until full fade (60-80)
}

type BurnAnimation struct {
    width, height int
    palette       []string  // 12 colors (9 burn + 2 smoke + 1 bg)
    theme         string
    withText      bool
    textContent   string

    // Character management
    chars         []BurnChar

    // Spanning tree
    spanningTree  []TreeEdge
    burnOrder     []int        // Ordered char indices
    nextIgnition  int          // Index in burnOrder

    // Smoke
    smokeParticles []SmokeParticle
    smokeSymbols   []rune

    // Phase tracking
    phase         BurnPhase
    frameCount    int
    cycleFrame    int

    // Sequences
    burnSequence  []rune  // For burn-text: morph sequence
    flameSequence []rune  // For burn: flame sequence
}
```

### Dual Mode Operation

**Background Effect (burn):**
```bash
syscgo -effect burn -theme rama -duration 30
```

**Text Destruction (burn-text):**
```bash
syscgo -effect burn-text -file SYSC.txt -theme rama -duration 30
```

Follows existing sysc-go pattern: `fire`/`fire-text`, `smoke`/`smoke-text`

## Spanning Tree Propagation

### Prim's Algorithm for Organic Fire Spread

**Why Spanning Tree:**
Creates connected, organic burn propagation where fire "spreads" from character to character. Each character ignites when fire reaches it from a neighbor, creating natural-looking fire advancement.

**Algorithm:**

```go
func (b *BurnAnimation) generateSpanningTree() {
    // Select origin with center bias
    origin := b.selectCenterBiasedOrigin()

    // Prim's Minimum Spanning Tree
    visited := make(map[int]bool)
    edges := []TreeEdge{}
    priorityQueue := []Edge{{from: origin, to: origin, weight: 0}}

    for len(priorityQueue) > 0 && len(visited) < len(b.chars) {
        edge := extractMin(priorityQueue)

        if visited[edge.to] {
            continue
        }

        visited[edge.to] = true
        edges = append(edges, TreeEdge{from: edge.from, to: edge.to})

        // Add neighbors with random weights
        for _, neighbor := range b.getNeighbors(edge.to) {
            if !visited[neighbor] {
                weight := rand.Float64() // Randomness creates varied spread
                priorityQueue = append(priorityQueue, Edge{
                    from: edge.to,
                    to: neighbor,
                    weight: weight,
                })
            }
        }
    }

    b.spanningTree = edges
    b.burnOrder = b.extractBurnOrder(edges)
}
```

**Center-Biased Origin:**
- Calculate center region (middle 40% of canvas)
- 70% chance: pick origin from center region
- 30% chance: pick from anywhere
- Creates dramatic radial burns biased toward center

**Neighbor Definition:**
- 4-directional connectivity (up, down, left, right)
- Characters within Manhattan distance of 1
- Creates adjacent spreading pattern

## Phase Details

### Phase 1: Ignition (6 seconds / 120 frames)

**Character Setup:**
- `burn`: Fill canvas with random printable ASCII (A-Z, 0-9, symbols) in muted theme color
- `burn-text`: Parse and center user text, characters in muted theme color

**Spanning Tree Generation:**
- Generate center-biased origin point
- Build spanning tree using Prim's algorithm
- Extract burn order (BFS traversal of tree)

**Sequential Ignition:**
- Ignite 1-2 characters per frame following burn order
- Ignited characters:
  - Begin at color palette[0] (white hot)
  - Start burn stage sequence
  - Begin emitting smoke
- Creates wave of fire spreading organically across canvas

**Timing:** 120 frames with ~500 chars = ~4 chars ignited every 2 frames

### Phase 2: Full Burning (8 seconds / 160 frames)

**All Characters Active:**
- All characters have caught fire
- Each character independently progresses through:
  - **Burn stage sequence**:
    - `burn-text`: 9 stages (`' . ▖ ▙ █ ▜ ▀ ▝ .`), 10-15 frames per stage
    - `burn`: 8 stages (`▁ ▂ ▃ ▄ ▅ ▆ ▇ █`), 15-20 frames per stage
  - **Color gradient**: 9 colors, 12-15 frames per color

**Smoke Emission:**
- Every burning character emits smoke continuously
- Spawn rate: 1 smoke particle every 3-5 frames per character
- Heavy smoke fills upper screen

**Visual State:**
- Maximum visual activity
- Characters at different burn/color stages (staggered ignition)
- Dense smoke layer rising

### Phase 3: Settle (5 seconds / 100 frames)

**Character Completion:**
- Characters finish burn sequences and disappear
- Disappearance order matches ignition order (first ignited = first to vanish)
- Smoke emission stops as characters disappear
- Existing smoke continues rising and fading

**Progressive Clearing:**
- Canvas empties from ignition origin outward
- Smoke dissipates naturally
- By end: clean canvas with fading smoke trails

### Phase 4: Pause (2 seconds / 40 frames)

**Brief Rest:**
- Empty canvas
- Final smoke particles fade out
- Moment of stillness

**Reset Preparation:**
- Frame 20: Generate new spanning tree with new center-biased origin
- Frame 35: Regenerate characters
  - `burn`: New random ASCII characters
  - `burn-text`: Reset text to initial state
- Frame 40: Transition to PhaseIgnition

**Total Cycle:** 15-17 seconds, continuous loop

## Burning Animation Details

### Background Effect (`burn`) - Simplified Flames

**Flame Sequence:**
```go
flameSequence := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
// Bottom fill → full block (8 stages)
```

**Timing:**
- Each flame stage: 15-20 frames
- Total sequence: ~120-160 frames
- Creates smooth vertical "filling" effect

**Initial Characters:**
- Random printable ASCII: letters, numbers, symbols
- Distributed across canvas
- Start in muted theme color (palette[11] or darkest burn color)

### Text Effect (`burn-text`) - Complex Morphing

**Morph Sequence:**
```go
burnSequence := []rune{'\'', '.', '▖', '▙', '█', '▜', '▀', '▝', '.'}
// 9 stages simulating flame shapes consuming character
```

**Timing:**
- Each morph stage: 10-15 frames
- Total sequence: ~90-135 frames
- After final `.`, character disappears

**Visual Effect:**
- Character distorts through flame-like shapes
- Gradual dissolution
- Complete disappearance to empty space

## Color System & Theme Integration

### Palette Function

**GetBurnPalette Structure:**

Returns 12 colors for comprehensive burn theming:

```go
// Returns: [9 burn gradient, 2 smoke colors, 1 background]
// Indices 0-8: Burn progression (white hot → dark ember)
// Indices 9-10: Smoke colors (primary, secondary)
// Index 11: Background/canvas color
func GetBurnPalette(themeName string) []string
```

### Example Theme Palettes

**RAMA Theme (Warm Reds):**
```go
case "rama":
    return []string{
        // Burn gradient
        "#ffffff", "#fff5e1", "#ffebcd", "#ffd700",
        "#ff8c00", "#ff4500", "#dc143c", "#8b0000", "#2b1a1a",
        // Smoke
        "#6b6b6b", "#4a4a4a",
        // Background
        "#2b2d42",
    }
```

**Dracula Theme (Cool Purples/Pinks):**
```go
case "dracula":
    return []string{
        // Burn gradient (purple/pink fire)
        "#f8f8f2", "#f5f5dc", "#dda0dd", "#da70d6",
        "#ba55d3", "#9370db", "#8b008b", "#4b0082", "#1a0033",
        // Smoke
        "#6272a4", "#44475a",
        // Background
        "#282a36",
    }
```

**Nord Theme (Ice Fire - Blues/Cyans):**
```go
case "nord":
    return []string{
        // Burn gradient (cool fire)
        "#eceff4", "#e5e9f0", "#d8dee9", "#88c0d0",
        "#81a1c1", "#5e81ac", "#4c566a", "#3b4252", "#2e3440",
        // Smoke
        "#4c566a", "#3b4252",
        // Background
        "#2e3440",
    }
```

**Gruvbox Theme (Warm Earthy Fire):**
```go
case "gruvbox":
    return []string{
        // Burn gradient (classic warm fire)
        "#fbf1c7", "#ebdbb2", "#fabd2f", "#fe8019",
        "#d65d0e", "#fb4934", "#cc241d", "#9d0006", "#282828",
        // Smoke
        "#665c54", "#504945",
        // Background
        "#282828",
    }
```

Support all existing themes: dracula, gruvbox, nord, tokyo-night, catppuccin, material, solarized, monochrome, transishardjob, rama, eldritch, dark

### Color Application

**Burn Color Progression:**
1. Character ignites: Start at palette[0] (white hot)
2. Progress through gradient: 12-15 frames per color
3. 9 total colors = ~108-135 frames for full gradient
4. Independent progression per character (staggered ignition creates varied colors)
5. Final: palette[8] (dark ember) before disappearing

**Smoke Colors:**
- Primary: palette[9] (main smoke color)
- Secondary: palette[10] (variation)
- Randomly alternate per particle
- Apply opacity multiplier: `color * opacity`

**Background:**
- palette[11] for canvas background
- Burnt-away areas show background
- Maintains theme aesthetic

### Theme Characteristics

**Cool Themes (Nord, Tokyo Night):**
- "Ice fire" with blues, cyans, cool tones
- Cool gray smoke
- Ethereal, otherworldly feel

**Warm Themes (Gruvbox, Material, RAMA):**
- Traditional fire with oranges, reds, warm tones
- Warm gray/brown smoke
- Classic burning aesthetic

**Vibrant Themes (Dracula, Catppuccin):**
- Colorful fire with purples, magentas
- Accent-colored smoke
- Unique, stylized burning

**Monochrome:**
- Grayscale fire progression
- Simple gray smoke
- Clean, minimalist

## Smoke Particle System

### Theme-Aware Smoke Symbols

```go
func (b *BurnAnimation) getSmokeSymbols() []rune {
    switch strings.ToLower(b.theme) {
    case "monochrome", "dark":
        return []rune{'▒', '░', '▓'}  // Block characters
    case "dracula", "catppuccin", "rama":
        return []rune{'·', '˙', '°'}  // Small dots
    case "nord", "gruvbox", "tokyo-night":
        return []rune{'.', '\'', ',', '·'}  // Mixed punctuation
    case "material", "solarized":
        return []rune{'░', '·', '.'}  // Mix of blocks and dots
    default:
        return []rune{'·', '.', '░'}  // Default mix
    }
}
```

### Smoke Behavior

**Simplified Emission:**
- Every active burning character emits smoke
- No probability check (simpler than TTE)
- Spawn rate: 1 particle every 3-5 frames per character
- Spawns from character's position

**Rising Motion:**
- Velocity: 0.5-0.8 cells per frame (medium rise)
- Random velocity per particle within range
- Straight upward movement (no horizontal drift)
- Consistent, predictable rising

**Fading:**
- Start opacity: 1.0 (fully visible)
- Linear fade over lifetime
- Max lifetime: 60-80 frames (~3-4 seconds)
- Opacity formula: `1.0 - (lifetime / maxLife)`
- Disappears at opacity < 0.05 or when off-screen

**Update Logic:**
```go
func (b *BurnAnimation) updateSmoke() {
    // Update existing particles
    active := []SmokeParticle{}
    for _, smoke := range b.smokeParticles {
        smoke.y -= smoke.velocity  // Rise upward (negative Y)
        smoke.lifetime++
        smoke.opacity = 1.0 - (float64(smoke.lifetime) / float64(smoke.maxLife))

        // Keep if still visible
        if smoke.y >= 0 && smoke.opacity > 0.05 {
            active = append(active, smoke)
        }
    }
    b.smokeParticles = active

    // Spawn from burning characters
    for i, char := range b.chars {
        if char.active && b.shouldSpawnSmoke(i) {
            b.spawnSmoke(char.x, char.y)
        }
    }
}

func (b *BurnAnimation) shouldSpawnSmoke(charIndex int) bool {
    // Spawn every 3-5 frames per character
    return (b.frameCount + charIndex) % (3 + rand.Intn(3)) == 0
}
```

## Implementation Details

### File Structure

**Files to Create:**
- `animations/burn.go` - Core burn animation
- Add `GetBurnPalette(theme string)` to `animations/palettes.go`

**CLI Integration (`cmd/syscgo/main.go`):**
- `runBurn(width, height, theme, frames)` - Background effect
- `runBurnText(width, height, theme, file, frames)` - Text destruction
- Add to effect switch statement
- Update help text

### Core Implementation Components

**Spanning Tree Generation:**
- `generateSpanningTree()` - Prim's MST from center-biased origin
- `selectCenterBiasedOrigin()` - 70% center, 30% random
- `getNeighbors(charIndex)` - 4-directional neighbors
- `extractBurnOrder(edges)` - BFS traversal for ignition sequence

**Phase State Machine:**
```go
func (b *BurnAnimation) Update() {
    b.frameCount++
    b.cycleFrame++

    switch b.phase {
    case PhaseIgnition:
        b.updateIgnition()
    case PhaseBurning:
        b.updateBurning()
    case PhaseSettle:
        b.updateSettle()
    case PhasePause:
        b.updatePause()
    }

    b.updateSmoke()
}
```

**Rendering Pipeline:**
1. Create buffer (width × height rune array)
2. Create color buffer
3. Render background (palette[11])
4. Render burning characters (current stage rune + color)
5. Render smoke particles (symbol + color with opacity)
6. Use lipgloss for color styling with opacity
7. Build final frame string

### Performance Considerations

**Spanning Tree:**
- Generate once per cycle (every ~340 frames)
- O(n log n) complexity, one-time cost
- Cache burn order for ignition phase
- Typical chars: 500-2000 depending on canvas size

**Active Character Management:**
- Track all chars in slice
- Mark as inactive when disappeared
- Filter inactive during render (skip if !active)
- Typical active during burning: 100-1000 chars

**Smoke Particles:**
- Cap at ~300-500 particles
- Remove when off-screen or faded (opacity < 0.05)
- Spawn rate controlled per character (every 3-5 frames)
- Prevents particle overflow

**Memory Usage:**
- Spanning tree: ~O(n) edges = ~4KB for 1000 chars
- Burn chars: ~1000 chars × 40 bytes = ~40KB
- Smoke particles: ~500 × 48 bytes = ~24KB
- Total: ~70KB, minimal footprint

### Terminal Compatibility

- Handle small terminal sizes gracefully
- Implement `Resize(width, height int)` to regenerate spanning tree
- Ensure burn symbols render correctly across terminals
- Test with various terminal sizes (80×24 to 200×60)

## Success Criteria

- [ ] Spanning tree creates organic, connected fire propagation
- [ ] Characters morph through burn sequences smoothly
- [ ] Color gradient transitions are smooth (12-15 frames per color)
- [ ] Center-biased origin creates dramatic radial burns
- [ ] Smoke rises and fades naturally (medium speed)
- [ ] Theme-aware colors work for all themes (cool/warm/vibrant)
- [ ] Slow burn timing feels dramatic (~15-17s total cycle)
- [ ] Seamless loop with new origin each cycle
- [ ] Performance maintains 20 FPS with ~500 active elements
- [ ] Background `burn` fills with random chars correctly
- [ ] Text `burn-text` burns user text correctly
- [ ] Compatible with both CLI and TUI

## Configuration Philosophy

Following YAGNI principle, keep configuration minimal:
- Effect name determines behavior (`burn` vs `burn-text`)
- Theme determines all colors (via GetBurnPalette)
- File input for text (burn-text only)
- Duration for how long to run

No additional flags needed. Spanning tree origin is center-biased random by design.
