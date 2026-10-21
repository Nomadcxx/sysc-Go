# L2A1 refinement and rotating statue implementation plan

## Phase 1: L2A1 follow-up PR

1. Add a focused baker test proving complete omission of the `308 casing_30`
   and `308 cartridge_31` subtrees, including child mesh nodes.
2. Add the two source-specific exclusions, regenerate `l2a1-sar.bin`, and
   update the source-change record. Keep the original GLB intact.
3. Compare scale and directional-light candidates. At 80x24, 1.12 already
   spans columns 1–78, and 1.16 touches the outer columns, so retain 1.12.
   Use low-ambient directional lighting to separate front and rear surfaces.
   Add regression checks for shade separation, edge margin, footer, and AK
   behavior.
4. Run focused baker, mesh, and animation checks. Manually review the diff,
   open a follow-up PR against master, review it locally and through CI, then
   merge if LGTM.

## Phase 2: six rotation-only models in sysc-Go

1. Generalize the pinned offline GLB baker and sample codec for bounded
   variable sample counts up to 96,000, keeping 24,000 as the AK/L2A1 default
   and starting each statue at 48,000. Preserve source SHA checks, selected-scene traversal,
   transform/index validation, and strict resource limits. Add tests for the
   larger valid model, invalid limits, node filtering, and variable sample
   lengths.
2. Bake the six pinned source GLBs sequentially, starting at 48,000 points and
   raising only models that gain visible detail at their chosen on-screen size
   while staying inside the render budget. Record hashes, source URLs, credits,
   CC BY/CC BY-NC terms, and reproduction commands. Do not commit source GLBs.
3. Add the model catalog and one `RotatingModelAnimation` constructor. Implement
   constant-speed horizontal yaw with zero pitch, zero roll, zero blend, fixed
   camera elevation, model-specific upright fit and screen margin, normal
   shading, and the same credit-row style. Avoid top or underside views unless
   needed to identify a model.
4. Add registry metadata and CLI/TUI dispatch. Test construction, continuous
   phase wrap, full-turn edge bounds, deterministic rendering, footer text,
   and normal/shade behavior for every model. Keep AK/L2A1 timeline tests.
5. Run focused sysc-Go checks and the 491x90 render budget check. Open the
   effects PR, review and merge when LGTM.

## Phase 3: consumers

1. From clean worktrees based on each repository's current default branch,
   update sysc-terminal's sysc-Go pin and constructor, and update sysc-greet's
   sysc-Go pin, background menu, lifecycle, resize/theme behavior, and notices.
2. Verify sysc-shell's existing `--list` refresh and Terminal Art card path
   exposes all six IDs; do not add a duplicate static shell registry.
3. Run repository-specific focused tests and builds, manually review each
   consumer diff, open PRs, and merge when LGTM. The live testbench is
   unavailable, so do not claim a deployed or live Niri result.

## Verification commands

- sysc-Go: `GOMAXPROCS=2 go test -p 2 ./internal/mesh ./cmd/bake-model ./animations ./tui ./cmd/syscgo`
- sysc-terminal: `timeout 90s env GOMAXPROCS=2 go test -count=1 ./internal/effect -run TestRegistryTickOnce`
- sysc-greet: focused `cmd/sysc-greet` background selection/render tests and a build.
- sysc-shell: focused Terminal Art catalog refresh/card tests from a clean
  `origin/main` worktree.
