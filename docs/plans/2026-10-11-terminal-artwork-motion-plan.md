# Animated character scenes implementation plan

**Goal:** Deliver readable block ASCII girl/wasteland scenes and a braille wave with authored continuous motion.

**Architecture:** Procedural geometry and a bounded character canvas in sysc-Go; elapsed active time flows through the existing wallpaper worker. The authoritative constructors and shell catalog remain the integration boundary.

**Tech stack:** Go standard library, existing sysc-Go palettes, native block and font raster.

Use the executing-plans workflow in the isolated artwork-motion worktrees. Tracking lives in sysc-terminal-8 and sysc-1136/1137. The installed bd export skips unchanged records while replacing the entire JSONL file; preserve database snapshots and do not run a full export without guarding record preservation.

## 1. Canvas and wave prototype

Files: animations/artwork.go, animations/artwork_wave.go, animations/artwork_test.go.

1. Run the existing named artwork check as a baseline.
2. Write TestArtworkMotion before implementation: require character geometry, deterministic Render, reset, bounds and changes in the wave crest/body regions over a complete cycle. Run it and record the expected failure.
3. Replace the wave bake with filled curved water forms, travelling ink contours, articulated foam and ballistic spray. Reuse braille bit order. Keep public constructors compatible.
4. Run the named check; generate a complete cycle through the actual engine parser/raster at both monitor grids and 80 by 24.
5. Inspect the full loop and show the owner the prototype before release.

## 2. Block ASCII portrait and wasteland

Files: animations/artwork.go, animations/artwork_girl.go, animations/artwork_wasteland.go, animations/artwork_test.go, animations/artwork/README.md.

1. Extend the same focused check for quadrant composition, animated face/blood regions and distant wasteland motion. Run the failing checks.
2. Draw the girl's facial planes and visor in block ASCII. Add head/mouth movement, blood accumulation and falling drops, visor scans and light changes.
3. Draw wasteland's tonal layers, sun, ruins, foreground debris and figure. Add cloud drift, dust, reflection motion and a distant moving object.
4. Remove the obsolete embedded cell loops and bake-only checks after all constructors use procedural drawing. Retain reference provenance.
5. Run named animation and selector checks. Render complete cycles in two themes at the acceptance grids. Inspect critical details and show the complete loops.

## 3. Active-time integration

Files: internal/effect/effect.go, internal/effect/effect_test.go, internal/wayland/owner.go, internal/wayland/worker.go, internal/wayland/worker_test.go.

1. Write a focused named check that compares 50 ms and 120 ms advancement at equal elapsed active time and proves pause excludes elapsed time. Run the expected failure.
2. Add UpdateElapsed(time.Duration) to the artwork type and route elapsed tick duration through the existing worker. Keep ordinary effects' Update behavior. Clear the owner's tick baseline at pause/resume and clamp exceptional gaps.
3. Use a temporary development modfile pointing at the isolated library worktree. Run named effect and worker checks; build the preview helper against the same renderer.

## 4. Qualification and release

Files: existing internal/pipeline/budget_test.go, go.mod/go.sum, qualification report in docs/plans.

1. Measure complete cycles at actual wallpaper dimensions for all scenes, blur 0 and 64; use the existing named budget test and inspect CPU mean/p95.
2. Review the local diff and run focused package checks and builds. Screen commit messages with the installed hook; keep author/committer Nomadcxx <noovie@gmail.com>.
3. Obtain complete-loop visual acceptance before merging or replacing the installed build. Merge library first, pin its remote commit in the engine, repeat the focused integration and CPU gates on the final pin.
4. Merge through a local human-identity commit and normal push. Install the engine via documented CGO-disabled Go install from the remote main branch. Deploy shell only if shell code changes.
5. Re-query live output assignments, process ownership and installed versions. Qualify through owned sockets/pids and preserve current owner selections. Record RSS soak and pause evidence; close only satisfied acceptance gates.

Focused command form:

```sh
timeout 90s env CGO_ENABLED=0 GOMAXPROCS=2 go test -count=1 ./animations -run TestArtwork
timeout 90s env CGO_ENABLED=0 GOMAXPROCS=2 go test -count=1 ./tui -run 'TestArtworkSelector|TestAnimationNeedsFile'
timeout 90s env CGO_ENABLED=0 GOMAXPROCS=2 go test -count=1 ./internal/effect -run TestArtwork
timeout 90s env CGO_ENABLED=0 GOMAXPROCS=2 go test -count=1 ./internal/wayland -run TestArtworkElapsed
```
