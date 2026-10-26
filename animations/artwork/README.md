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

The Met marks its reference as public domain. The supplied GIFs retain their
creator's credit; this adaptation makes no public-domain claim for them.
Each effect prints its source credit.

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
