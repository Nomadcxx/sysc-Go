# Animated character scenes

The effects draw scene geometry in Go. Cyberpunk girl and wasteland compose
coloured quadrant block characters. Great Wave composes braille dots using
the logo effects' dot order. The renderer receives character cells and ANSI
colours, with no source image playback.

| Scene | Reference | Source SHA-256 |
|---|---|---|
| great-wave | Katsushika Hokusai, *Under the Wave off Kanagawa*, ca. 1830–32; [The Met, JP1847](https://www.metmuseum.org/art/collection/search/45434), public domain | `cbb9988f2f18b9180a1cc0cf5dbb0e3bbf8f2e85d17930cf6ca9ab36fac36303` |
| cyberpunk-girl | Owner-supplied `dystopian_cyberpunk_girl.gif`; visible credit: Valenberg.com | `743d4092e31acaf4d0751eaeb6b36b569249504cefd3db7fecb9adb09fe0209d` |
| wasteland | Owner-supplied `wasteland.gif`; visible credit: Valenberg.com | `23cd9100f9c57e149a5dd23060d4b04779815fd622b0e5723f2b4f8052f42d69` |
| ansi-love | Warpus (Legacy Krew), “ansi love”, December 2019; [original ANSI](https://www.ansilove.org/assets/examples/wa-ansilove.ans); visible credit: Drawn by Warpus, December 2019 | `5c5d75b205680a1098517ef36432ce36161415c2dd2f9552b9ede7cdc2330ce3` |

The Met marks its reference as public domain. The supplied GIFs retain their
creator's credit; this adaptation makes no public-domain claim for them.
The ANSI source and baked raster are embedded for runtime rendering. No artwork
licence is claimed; obtain the artist's permission before public release. Each
effect prints its source credit.

Girl: a full-height close-up with head sway, breathing, moving lips and cheek
light, blood accumulating and dripping from the nose, and scanning visor
fragments. Wasteland: layered ruins and wreckage, clouds, dust, a distant
vehicle, birds, and moving water reflections. Wave: a deforming curl and
breaking lip, articulated foam, falling spray, boats and fixed Mount Fuji.

Complete scene periods are 12, 16 and 8 seconds respectively. Update uses a
nominal 50 ms; UpdateElapsed accepts active time from a host with a different
cadence. It caps exceptional dispatch gaps at 250 ms. Hosts reset their time
baseline on pause/resume. Render does not advance time. Resize preserves phase.

Scenes clamp output to 512 by 256 cells. Their sample buffers contain at most
1 MiB of byte colour indices. A fixed 16-colour palette bounds style diversity.
All scenes use measured font cell aspect to preserve physical proportions.
The girl fills the height; landscapes extend scenery across wide outputs.
Polygon strokes clip work to the canvas even at extreme cell ratios.
Colours derive from existing theme palettes, with distinct foreground, middle
and distant tones. Empty cells keep a stable foreground.

Focused checks:

```sh
timeout 90s env CGO_ENABLED=0 GOMAXPROCS=2 go test -count=1 ./animations -run TestArtwork
timeout 90s env CGO_ENABLED=0 GOMAXPROCS=2 go test -count=1 ./tui -run 'TestArtworkSelector|TestAnimationNeedsFile'
```

## ANSI Love source and baked raster

Drawn by Warpus, December 2019

`ansi-love.ans` is the unmodified original downloaded from
<https://www.ansilove.org/assets/examples/wa-ansilove.ans> (23,351 bytes).
Its SHA-256 is `5c5d75b205680a1098517ef36432ce36161415c2dd2f9552b9ede7cdc2330ce3`.
SAUCE records `ansi love`, author `warpus`, group `Legacy Krew`, date
`20191228`, 80 columns × 157 rows, IBM VGA, flags `0x12` (8-pixel spacing,
square pixels, iCE off). Source comments and SAUCE remain intact.
This attribution does not grant or assert an artwork licence.

Regenerate from the repository root:

```sh
go run ./animations/cmd/bake-ansilove
go test ./animations/internal/ansiart ./animations/cmd/bake-ansilove
```

The bake checks the original source digest before decoding. It rasterizes the
80×157 CP437 cells with an 8×16 IBM VGA font into 640×2512 pixels. Runtime
rendering uses the embedded compressed pixel data; the font and bake command
are development tools, with no runtime font or external command dependency.
`ansi-love.raster.zlib` contains one zlib stream: `ALR1`, little-endian uint32
width and height, then one byte per palette index and one byte per region.
Palette indices use ANSI SGR order (black, red, green, yellow, blue, magenta,
cyan, white, then their bright counterparts). The current 68,474-byte raster
has SHA-256 `b7911bf6f89d593eae81abab0f1f94523adda240d1ce51c743f869a8fb4d6a95`.

We extracted the font bytes in `../cmd/bake-ansilove/ibm-vga8.bin` in order
from `src/fonts/font_pc_80x25.h` in libansilove revision
[`e22973d569c97c1a521a5246e057d300e865bd9f`](https://github.com/ansilove/libansilove/blob/e22973d569c97c1a521a5246e057d300e865bd9f/src/fonts/font_pc_80x25.h).
The header SHA-256 is
`9486a202c7a6247dbffa7d788c655051bd232e27b75412e3725823e61cd40d15`;
the 4,096 extracted bytes have SHA-256
`ff3640ffffdd774ec56ee7ecacf7bce8324a67071155ade07f4c6380fa879025`.
We retain the upstream BSD 2-Clause notice in
[`FONT-LICENSE`](../cmd/bake-ansilove/FONT-LICENSE); it applies to the font
source, not to the artwork.

Semantic labels use zero-based source rows: header 3–13, upper artwork 15–106,
logo white/grey pixels 104–126, heart red pixels 126–149, credits 151–155.
The provisional EKG label conservatively covers red pixels in rows 132–141;
some adjacent heart pixels in that band also remain static. Black pixels stay
background, including the trace's negative-space cutout. The exact EKG mask
still needs review before independent animation.

Validation used libansilove at the pinned revision, built in a temporary
directory and invoked with font CP437, 80 columns, 8 bits and iCE off. All
1,607,680 RGB pixels match the Go bake. The SHA-256 of the row-major RGB bytes
is `0de6e314dcdb1863ad7d49f91a3141c6677a4c655b2182c7dc3277776b61dcac`.
The bake test checks this digest, the font digest, source digest and byte-for-byte
regeneration of the compressed asset. The 9-pixel website PNG is a composition
preview; validation used a fresh 8-pixel render of the original ANSI file.
