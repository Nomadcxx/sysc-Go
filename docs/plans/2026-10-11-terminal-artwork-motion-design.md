# Animated character scenes

The owner approved rebuilding the three scenes after rejecting the deployed density-frame conversion. The girl and wasteland use block ASCII. Great Wave uses braille contours and shaded water. The references guide composition; each effect draws and animates its own geometry.

## Composition and motion

Cyberpunk girl fills the output height. Keep facial proportions using measured cell aspect; extend the dark surroundings across wide outputs. Draw the jaw, cheek planes, nose, lips and visor with block elements. Animate subtle head sway, breathing and mouth movement. Blood gathers beneath the nose, extends into a thread, separates into a falling drop and repeats. Shift the highlight on the cheeks. Animate visor scan bands and moving glitch fragments independently of the face. Preserve the original face and injury rather than replacing them with a generic portrait.

Wasteland fills the output. Separate the bright sun, clouds, distant ruins, small figure, reflection and angular foreground wreckage with distinct tonal levels and outlines. Move clouds, dust and water reflections at different speeds. Add a small distant moving silhouette with a visible scale difference from the foreground figure. Keep the original sunset composition and the cyan detail on the figure.

Great Wave retains Hokusai's curled crest, foam fingers, boats, sea bands and distant Fuji. Draw curved water geometry and ink contours at braille dot resolution. Animate the crest rising, curling and breaking, with flow along its body. Extend and retract foam fingers; launch spray that follows falling arcs. Boats bob with the underlying water; Fuji stays fixed. The wave changes shape through its complete cycle.

## Ownership and rendering

Keep the existing three constructors and standalone registry entries in sysc-Go. Replace baked playback with bounded procedural scene drawing in animations. Reuse the logo's braille bit order and the engine's native block raster. Use a small shared canvas for filled polygons, lines and curves; do not import a graphics framework. Quantize theme colours into a fixed, small palette so ANSI output and raster tiles remain bounded. Give empty cells stable styles.

For block scenes, draw two by two samples per character and compose quadrant glyphs with foreground/background colours. Draw the wave with two by four braille samples. Work at the requested grid, capped at 512 by 256 cells. Preserve geometry through resize using the measured cell ratio. Place readable provenance in the last row without reserving a blank band through the scene.

Advance scene time only in Update. Add an elapsed-duration update seam for these effects, retaining nominal 50 ms Update for existing hosts. The wallpaper owner supplies elapsed active time through its existing worker request; pause/resume clears the timing baseline so paused time never advances the scene. Render stays deterministic and does not advance time. Keep frame callbacks and buffer releases as the sole scheduler.

## Acceptance and release

Show complete loops at 491 by 90 and 365 by 90 cells, corresponding to the owner's outputs, plus 80 by 24. Include Kanagawa and Dracula. Review the wave prototype first, then the girl's facial details and wasteland depth. Visual acceptance requires readable forms and sustained motion; a changed-frame count alone proves neither.

Retain the approved CPU mean/p95 caps: 12.5 ms at blur 0 and 30 ms at blur 64. RSS stays below 150 MiB; ten-minute growth stays below 10 MiB; pause does zero effect work. Use focused named checks for bounds, block/braille composition, repeated rendering, animated regions, resize/reset and elapsed-time equivalence. Reuse the existing budget pipeline and installed font. Keep source GIFs local and preserve Valenberg.com credit. Credit Hokusai and The Met for the public-domain wave reference.

After complete-loop review and resource qualification, merge the library before pinning the engine. Install only the reviewed merged engine. Keep the owner's live assignments intact during headless development and use owned sockets/pids for any approved live qualification.
