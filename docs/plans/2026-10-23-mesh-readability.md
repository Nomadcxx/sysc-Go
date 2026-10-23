# Shared mesh readability implementation plan

Goal: improve mesh readability through existing ANSI output, with no change to LogoSpin's shader or compositor. User approved the research recommendation and subsequent shared-renderer lighting/edge experiment.

Architecture: tune mesh shading before composition; compare a temporary depth/normal edge pass against lighting alone. Preserve theme colors, occupancy, depth ordering and morph behavior. Correct projection fit only when needed and apply measured cell geometry in CLI/TUI hosts using the installed Unix terminal API. Kitty image presentation remains a separately approved greeter integration, outside this shared-renderer slice.

1. Capture the baseline and add a small regression check for shading visibility and full-yaw small-viewport bounds. Run it before edits.
2. Compare lighting-only, edge-only and combined variants with the same models, angles and Braille compositor. Keep only the variant supported by captures; preserve LogoSpin byte-for-byte.
3. Limit projected model size to available margins without enlarging the calibrated display scale. Test small grids and existing morph behavior.
4. Connect existing SetCellAspect to terminal pixel measurements in CLI/TUI; retain existing aspect when unavailable. Use platform build tags and existing dependencies.
5. Run targeted animation/host tests and 30 ms budget. Save before/after captures, explain consumer adoption and remaining greeter work. Do not merge or publish.
