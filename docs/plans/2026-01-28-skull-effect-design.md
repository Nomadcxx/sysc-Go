# Skull Effect Design

**Date:** 2026-01-28
**Effect Names:** `skull`, `skull-text`
**Default Theme:** RAMA

## Overview

A terminal animation effect featuring an ASCII skull that drips into place from the top, with strategic accent points that illuminate in sequence using theme colors. The effect includes a two-layer ash particle system (light and dense smoky ash) with wind drift and subtle density variation. The `skull-text` variant adds user-provided ASCII art that slides in from both edges with a bounce effect.

## Visual Style

**Skull Design:**
- Hybrid geometric outline with internal shading/texture
- Based on reference images: geometric style with added detail
- Positioned in bottom half of terminal (similar to Christmas tree positioning)
- ASCII art to be generated using conversion tools from reference images

**Color Scheme:**
- Skull base: Muted gray tones (first palette color)
- Accent points: Progressive illumination using theme colors (dark → bright)
- Ash particles: Darker theme colors (first 2-3 palette colors)
- Text (skull-text only): Vibrant gradient using brightest 3-4 theme colors

## Animation Architecture

### Effect Structure

```go
type SkullAnimation struct {
    width, height int
    palette       []string
    theme         string
    withText      bool      // Toggle text integration
    textContent   string    // User's ASCII text/header

    // Animation state
    skullBuffer   [][]rune  // Final skull positions
    currentState  [][]cell  // Current render state with colors
    ashParticles  []Particle
    phase         AnimationPhase
    frameCount    int
}

type AnimationPhase int
const (
    PhaseDrip AnimationPhase = iota
    PhaseIllumination
    PhaseTextEntrance  // skull-text only
    PhaseHold
)
```

### Dual Mode Operation

**Background Effect (skull):**
```bash
syscgo -effect skull -theme rama -duration 30
```

**Text Integration (skull-text):**
```bash
syscgo -effect skull-text -file SYSC.txt -theme rama -duration 30
```

Follows existing sysc-go pattern: `fire`/`fire-text`, `matrix`/`matrix-art`, `rain`/`rain-art`

## Phase Details

### Phase 1: Skull Drip (100-120 frames / 5-6 seconds)

**Mechanics:**
- Each skull character starts at random height above final position (-5 to -15 rows)
- Drip speed varies per character: 0.3 to 0.8 cells per frame
- Simple gravity simulation: characters accelerate slightly as they fall
- Characters "lock in" upon reaching final position
- Skull rendered in muted gray tones (first palette color)

**Concurrent:**
- Ash particles begin falling immediately (atmospheric backdrop)

### Phase 2: Illumination Sequence (80 frames / 4 seconds)

**Strategic Accent Points (in order):**
1. **Top details** (frames 111-130): Forehead, crown area
2. **Eyes** (frames 131-155): Eye sockets - longer hold for emphasis
3. **Cheekbones** (frames 156-175): Mid-skull structural elements
4. **Teeth** (frames 176-190): Bottom jaw - final focal point, brightest

**Implementation:**
- Accent regions identified by coordinate ranges
- Illumination uses theme palette progression (darkest → brightest)
- Each region holds its illuminated state after lighting up
- Smooth color transitions using theme colors

### Phase 3: Text Entrance (40 frames / 2 seconds, skull-text only)

**Text Splitting:**
- Split text vertically at terminal center column
- Left half: Spawns off-screen left (x = -5)
- Right half: Spawns off-screen right (x = width + 5)
- Each character tracks final destination

**Easing Function (Out-Back):**
```go
// Overshoots target then settles back
func easeOutBack(t float64) float64 {
    c1 := 1.70158
    c3 := c1 + 1
    return 1 + c3 * math.Pow(t - 1, 3) + c1 * math.Pow(t - 1, 2)
}
```

**Animation:**
- All characters slide simultaneously
- Overshoot final position by ~5-10%
- Bounce back and settle
- Duration: 40 frames (2 seconds)

**Text Styling:**
- Horizontal gradient using 3-4 brightest theme colors
- Centered horizontally and vertically
- May overlap top portion of skull
- Z-order: Ash (back) → Skull (middle) → Text (front)

**Text Gradients per Theme:**
Define in `runSkullText` function similar to `runPour`, `runPrint`, etc. Each theme gets vibrant colors ensuring contrast.

### Phase 4: Hold State (200 frames / 10 seconds)

- All elements static except ash
- Skull fully illuminated
- Text displayed (skull-text only)
- Ash continues falling with density breathing

### Phase 5: Reset (2-3 frames)

- Quick fade or instant reset
- Ash particles continue seamlessly (no visible reset)
- Loop back to Phase 1

**Total Loop Time:**
- `skull`: ~19-20 seconds
- `skull-text`: ~21-22 seconds

## Ash Particle System

### Two-Layer Design

**Light Ash Layer:**
- Characters: `.`, `,`, `'`, `·`
- Fall speed: 0.5-1.5 cells per frame
- Horizontal drift: ±0.1-0.3 cells per frame (sine wave)
- Opacity: 40-70% of theme colors
- Lifecycle: Fade as falling, disappear near bottom

**Dense Smoky Ash Layer:**
- Characters: `▒`, `░`, `·`, `*`
- Fall speed: 0.3-0.8 cells per frame (slower)
- Horizontal drift: ±0.2-0.5 cells per frame (different sine frequency)
- Opacity: 60-90% of theme colors
- Lifecycle: Persist longer, slight accumulation at bottom (2-3 rows)

### Density Variation (Breathing Effect)

- Subtle pulse: 10-15% variation (not 30%)
- Formula: `density = baseDensity * (0.9 + 0.1 * sin(frameCount / 100))`
- Both layers sync to same density wave
- Creates atmospheric breathing without being distracting

### Wind Drift

- Each particle has sine wave horizontal movement
- Different frequencies per particle for organic feel
- Drift amplitude varies per particle type (light vs dense)

### Color Selection

- Use first 2-3 colors from theme palette (darker tones)
- Creates shadowy atmosphere
- Contrasts with bright skull accent illumination

## Implementation Tasks

### 1. Skull ASCII Art Generation

**Task:** Convert reference images to ASCII
- Investigate ASCII conversion tools
- Ensure quality capture of geometric outline + shading
- Generate skull art combining both reference styles
- Validate dimensions work for typical terminal sizes

### 2. Core Animation Implementation

**Files to Create:**
- `animations/skull.go` - Main skull effect
- Add `GetSkullPalette(theme string)` to `animations/palettes.go`

**Implement:**
- Skull drip mechanics with variable speeds and gravity
- Accent point identification and illumination sequencing
- Phase management and state transitions
- Reset logic that preserves ash continuity

### 3. Ash Particle System

**Implement:**
- Two-layer particle spawning and lifecycle
- Wind drift (sine wave horizontal movement)
- Density breathing (subtle 10-15% variation)
- Bottom accumulation for dense ash
- Fading and opacity management

### 4. Text Integration (skull-text)

**Implement:**
- Text splitting at center column
- Out-back easing function
- Simultaneous slide-in animation
- Text gradient application per theme
- Z-order rendering (ash → skull → text)

### 5. CLI Integration

**Update `cmd/syscgo/main.go`:**
- Add `runSkull()` function for background effect
- Add `runSkullText()` function for text variant
- Define text gradients for all themes
- Add to effect switch statement
- Update help text with new effects

### 6. Theme Support

**Add to `animations/palettes.go`:**
- `GetSkullPalette(theme string)` returning colors for:
  - Skull base color (muted)
  - Accent illumination progression (4-5 colors, dark → bright)
  - Ash particle colors (2-3 dark colors)
- Support all existing themes: dracula, gruvbox, nord, tokyo-night, catppuccin, material, solarized, monochrome, transishardjob, rama, eldritch, dark

### 7. Testing & Polish

- Test on various terminal sizes
- Verify timing feels right (drip speed, illumination, hold)
- Ensure ash looks atmospheric not distracting
- Validate text slides smoothly with bounce
- Test all themes for color contrast
- Verify TUI compatibility

## Technical Considerations

### Skull Positioning
- Bottom half of terminal like Christmas tree
- Calculate vertical offset based on skull height and terminal height
- Center horizontally

### Performance
- Manage particle count to avoid slowdown
- Target 20 FPS (50ms sleep between frames)
- Efficient rendering with lipgloss styling

### Terminal Compatibility
- Handle small terminal sizes gracefully
- Implement `Resize(width, height int)` if needed
- Ensure skull scales or wraps appropriately

### File Organization
- Follow existing pattern: one file per effect
- Config struct for complex initialization
- Palette function in centralized palettes.go
- CLI runners in main.go

## Success Criteria

- [ ] Skull drips smoothly over 5-6 seconds
- [ ] Accent illumination sequence feels natural and follows drip flow
- [ ] Ash creates atmospheric depth without overwhelming
- [ ] Text slides in with satisfying bounce effect (skull-text)
- [ ] All themes have appropriate color contrast
- [ ] Effect loops seamlessly (ash never appears to reset)
- [ ] Compatible with both CLI and TUI
- [ ] Performance maintains 20 FPS target
