# Particle model motion and Higgs field

## Behaviour

Keep the approved Braille shading: theme-coloured particles, 25–85% interior
coverage and preserved silhouettes. Inspect representative frames before changing
the shading again; keep the current result unless a comparison shows more detail.

Aircraft gain pitch, roll and vertical travel over their existing twelve-second
turn. Statues stay upright. The rasterizer reserves the full travel envelope
before fitting the model, so vertical movement alone cannot change its scale.
Position and normal rotations use the same transform.

`higgs-field` is a standalone field of particles with travelling waves and vortex
motion. Fit its spread to the viewport rather than using a sculpture display
scale. Keep the prototype dissolve/reform cycle on `napoleon-bust`, with its
source attribution and a twenty-second loop. Both use the shared mesh renderer
and analytic field motion, with reusable particle buffers and no integration
drift.

## Implementation and proof

1. Extend `animations/meshspin.go` with a pose containing yaw, pitch, roll and
   reserved vertical travel. Enable the flight profile in the four aircraft
   entries in `animations/rotating_model.go`.
   Check `TestRotatingModelConstructorAndLoop`,
   `TestMeshFlightTranslationKeepsScale`,
   `TestFighterFlightStaysInsideSmallViewports` and the existing frame budget.
2. Add `animations/higgs_field.go` with
   `NewHiggsFieldEffect(width, height int, theme string) *HiggsFieldAnimation`.
   Inherit elapsed-time updates, reset, resize and cell metrics from `meshSpin`.
   Check standalone field coverage, determinism, Napoleon dissolution, exact reconstruction, viewport bounds and
   the 30 ms frame budget in `animations/higgs_field_test.go`.
3. Register `higgs-field` in `animations/registry.go`, `cmd/syscgo/main.go` and
   `tui/animfactory.go`. The TUI selector already uses the shared registry.
   Run the CLI's advertised-effect check and the TUI constructor check.
4. Publish a tagged library release after capped Go checks. Pin that release in sysc-greet, sysc-terminal and sysc-walls. Add Higgs field to their
   existing factories and selectors. Make sysc-greet model backgrounds use
   Braille rather than the filled Kitty image path.
5. Verify consumers without local module replacements. Use focused named tests
   for sysc-terminal, as its repository requires. Preserve the desktop's running
   wallpaper and greeter; do not restart them for a source update.

All changes use existing dependencies. Keep renderer, motion, new effect and
consumer updates in separate commits. Commit as Nomadcxx <noovie@gmail.com>.
