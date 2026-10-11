# Built-in ASCII artwork

These files contain printable ASCII glyphs and four colour roles, not bitmap
frames. The Go baker keeps the GIF delays and composites offsets, transparency
and disposal before conversion. It uses the ASCII density ramp and 1:2 cell
convention from `tools/image_to_block_ascii.py`; it has no Python dependency.

| Scene | Reference | Source SHA-256 | Cells | Frames | Delay |
|---|---|---|---|---|---|
| great-wave | Katsushika Hokusai, *Under the Wave off Kanagawa*, ca. 1830–32; [The Met, JP1847](https://www.metmuseum.org/art/collection/search/45434), public domain | `cbb9988f2f18b9180a1cc0cf5dbb0e3bbf8f2e85d17930cf6ca9ab36fac36303` | 240×81 | 24 | 250 ms |
| cyberpunk-girl | Owner-supplied `dystopian_cyberpunk_girl.gif`; visible source credit: Valenberg.com | `743d4092e31acaf4d0751eaeb6b36b569249504cefd3db7fecb9adb09fe0209d` | 240×60 | 12 | 120 ms |
| wasteland | Owner-supplied `wasteland.gif`; visible source credit: Valenberg.com | `23cd9100f9c57e149a5dd23060d4b04779815fd622b0e5723f2b4f8052f42d69` | 240×68 | 160 | 90 ms |

The Met marks its reference as public domain. The supplied GIFs retain their
creator's credit; this adaptation makes no public-domain claim for them. The
original GIFs and raster stay local. The effects print a readable source credit.

Great Wave holds the composition fixed while foam and water glyph density varies
by one step. The supplied animations retain their complete loops. `Update`
advances a nominal 50 ms; `Render` reads the frame. Hosts with slower ticks play
slower: the wallpaper host's 120 ms blur cadence stretches loops by 2.4×.

Scenes fit within the grid and keep their source aspect ratio. Ordinary terminals
use 1:2 cells; wallpaper passes its measured font cell ratio. Shrinking averages
cell density; enlarging repeats cells. ASCII density preserves luminance while
existing theme palettes supply neutral, cool, warm and secondary colours. Warm
regions use the nearest warm hue in the theme's fireworks palette, so Kanagawa's
green primary accent does not recolour the sunset. Kanagawa, Dracula and
monochrome are covered by the focused checks.

The binary shares each decompressed scene once. Limits: 1 MiB compressed and
8 MiB inflated per scene; 160 frames; 256×128 baked cells; 512×256 output cells.
The loader rejects invalid headers, lengths, delays, glyphs and colour roles.

To reproduce a bake with a local reference:

```sh
go run ./tools/artwork-bake -input /absolute/path/wasteland.gif -output animations/artwork/wasteland.cells.gz
```

For Great Wave, use [the Met scan](https://images.metmuseum.org/CRDImages/as/original/DP130155.jpg)
and add `-wave`. Compression has no timestamps, so identical inputs produce
identical bakes. Focused checks:

```sh
timeout 90s env CGO_ENABLED=0 GOMAXPROCS=2 go test -count=1 ./animations -run TestArtwork
timeout 90s env CGO_ENABLED=0 GOMAXPROCS=2 go test -count=1 ./tools/artwork-bake -run TestArtworkGIFDisposal
timeout 90s env CGO_ENABLED=0 GOMAXPROCS=2 go test -count=1 ./tui -run 'TestArtworkSelector|TestAnimationNeedsFile'
```
