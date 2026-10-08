# logo-spin effect family — design

Date: 2026-10-07
Status: approved by owner (continuous spin + morph variant)

## Goal

Four new registry effects that render vector shapes as a 3D braille plate
spinning on its vertical axis, inspired by the "blossom" welcome animation
popularized by a recent OpenAI TUI release:

- `sysc-logo`     — block "SYSC" wordmark, continuous spin
- `cross-logo`    — Latin cross, continuous spin
- `justice-cross` — faceted justice cross wireframe, continuous spin
- `logo-morph`    — SYSC <-> cross morph (signed distance fields lerp) while spinning

Wallpaper use: `sysc-terminal` wires the same IDs into its construct switch,
so all four become selectable live wallpapers.

## Shape model

A shape is a list of thick strokes: `LogoStroke{X1,Y1,X2,Y2,T}` in a local
unit box. Each stroke expands to one axis-quad loop perpendicular to the
segment. Loops are wound consistently; the union is filled by nonzero winding
so overlapping strokes merge without artifacts. No SVG parser.

`sysc` = S, Y, S, C laid out as a 5x7 pixel font. Each letter is emitted as
horizontal runs of lit pixels rather than individual strokes, so the wordmark
stays crisp at braille resolution. `cross` = one tall vertical bar, one shorter
horizontal bar. `justice` = a 16-vertex outline plus three interior facet
creases in a 486x486 space, which reads as a chiselled cross in wireframe
style.

## Pipeline (per instance, baked once, cached by name)

1. Shape loops -> scanline pass on a 160x160 grid over `[-1.08,1.08]^2`:
   per-row crossing events (x, winding delta), sorted, accumulated; nonzero
   inside. Bounding box of the union is scaled to a centered extent with margin.
2. Two-pass chamfer distance transform (weights 1, sqrt(2)) for inside and
   outside distances; signed field `d = inside ? din : -dout`, converted to
   normalized units (step = 2*extent/(160-1)).
3. Gradient per grid point via central differences (used for bevel normals
   and side walls).

Per frame:

4. Blend: `d = mix(dA, dB, morph)` (morph = 0 except `logo-morph`).
5. Surface: interior points (d > 0) emit front/back pillow caps with bevel
   rounding `z = hd - b + b*sqrt(1-edge^2)`, `edge = clamp(1 - d/b, 0..1)`;
   boundary points emit vertical side walls with gradient normals.
   Traveling ripple in object space: `z += sin(2.2x+0.7y)*0.075 +
   sin(2.8y-0.4x)*0.04` with matching normal perturbation.
6. Motion: frame-based phase; the loop is `logoSpinFrames` = 144 (~7.2 s at
   20 fps) and each loop runs **two eased full turns**, so the plate passes
   through facing-forward twice per loop instead of turning continuously.
   Each half-loop smoothsteps its turn over the first 82% of its time, then
   holds facing forward for the remainder. Transform `R_z(tiltZ) * R_x(tiltX)
   * R_y(rotation)`; `tiltX` oscillates once per turn, `tiltZ` sways once per
   loop. Tilt base ~0.62 rad so the plate reads as 3D from above.
7. Perspective divide `s = 4.4/(4.4 - z)`; project to a dot grid of
   `(cols*2) x (rows*4)`, uniform pixel scale limited by height (terminal
   cells are ~2:1, and a braille cell's 2x4 dots are ~square in pixels).
8. Z-buffer keeps the nearest sample per dot (larger transformed z = nearer).
   Dots pack into braille glyphs: `rune(0x2800 | bits)`, bit map
   `[1,8,2,16,4,32,64,128]` indexed `(dx%2) + (dy%4)*2`.
9. Shading: key/fill diffuse + rim from |edge| + gloss pow-18 specular +
   ambient; intensity mapped along `GetLogoPalette(theme)` slots
   (0 deep shadow, 1 shadow, 2 mid, 3 lit, 4 highlight, 5 rim accent,
   6 background) with `hexLerp`. Cell color = average of its lit dots.
10. Render emits truecolor SGR like skull/cracktro:
    `\033[38;2;r;g;bmglyph\033[0m`, plain space where no dot is lit.

## Determinism and pacing

`Update()` only advances an integer frame counter; every visual is a pure
function of `frame`. Both hosts tick at ~20 fps (CLI sleeps cliFrameInterval,
engine caps at 20 fps), so on-screen speed matches everywhere and tests can
assert exact strings.

## Public API

```go
type LogoSpinConfig struct {
    Width, Height int
    Shape         string   // "sysc" | "cross" | "justice" | "sysc-cross" | "cross-sysc" | "morph"
    Palette       []string // GetLogoPalette(theme)
    Theme         string
    Style         string   // "edge" (default: crisp 3D wireframe) | "ring"/"wire" | "solid"
}
func NewLogoSpinEffect(c LogoSpinConfig) *LogoSpinAnimation
```

Implements `Animation` (+ `Resize` so the engine's resize path reuses it).
`Style` was added once the wireframe renderer shipped: `edge` rasterizes real
3D segments instead of a shaded field. Edge mode never morphs, so a morph
shape in edge mode is drawn as its ring-style contour rather than blending.

## Registry / wiring

- 4 `EffectRegistry` entries, VersionAdded "1.0.6", Category "abstract",
  RequiresText false. `LibraryVersion` -> "1.0.6".
- `GetLogoPalette(theme)` covers all ThemeRegistry names + default.
- `cmd/syscgo`: four cases + shared `runLogoSpin`.
- `tui`: four factory cases + four picker names.
- GUIDE.md effect section.
- sysc-terminal PR: cases in construct, go.mod bump, one effect test.

## Tests (named, single package)

- SDF sign test: cross center inside positive, far corner negative; SYSC bar
  interiors inside; `justice` bakes a non-empty field; unknown shape is nil.
- Braille range + row count + truecolor presence after several Updates.
- Determinism: two fresh instances agree on strings for same frame counts.
- Reset returns frame 0 == fresh instance first frame.
- Minimum grid (21x24) renders without panic; morph mid frame differs from
  both endpoints.
- Registry test covers all four IDs and `LibraryVersion`.
- `GetLogoPalette` covers every `ThemeRegistry` name at runtime.
- Edge geometry test pins the justice wireframe edge count.
- Style test covers every style string, including the morph-in-edge downgrade.

## Out of scope (YAGNI)

SVG/text input, user shape editor, grain, 256/ANSI-16 degradation,
light-terminal monochrome mode.
