# L2A1 SAR Effect Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a selectable, attributed L2A1 SAR braille effect with the AK-47's two-rotation and particle-reconstruction cycle.

**Architecture:** Share the AK/L2A1 sampled-mesh format and surface sampling in `internal/mesh`; parse and bake the pinned GLB offline with the Go standard library. Reuse AK's runtime depth-buffer rendering while preserving its artwork-specific behavior, then wire the new effect through the registry, CLI, and TUI.

**Tech Stack:** Go standard library, existing `LogoSpinAnimation` braille canvas, `go:embed`; no runtime model parser or added dependencies.

---

### Task 1: Extract reusable sampled-mesh data and surface sampling

**Files:**
- Create: `internal/mesh/mesh.go`
- Create: `internal/mesh/mesh_test.go`
- Modify: `cmd/bake-ak47/main.go`

**Steps:**

1. Move the existing 24,000-sample triangle area sampler and signed-int16 sample encoding/decoding into `internal/mesh`, retaining the AK algorithm's operation order and low-discrepancy constants.
2. Replace the AK baker's local sampler/encoder with the shared package.
3. Run `GOMAXPROCS=2 go test -p 2 ./internal/mesh ./cmd/bake-ak47`.
4. Confirm the existing AK animation tests still pass with `GOMAXPROCS=2 go test -p 2 ./animations -run TestAK47`.

### Task 2: Parse and bake the pinned L2A1 GLB

**Files:**
- Create: `cmd/bake-l2a1/main.go`
- Create: `cmd/bake-l2a1/main_test.go`
- Create: `animations/models/l2a1-sar.bin`

**Steps:**

1. Add bounded GLB v2 header/chunk parsing and JSON structs for scenes, nodes, meshes, accessors, and buffer views. Reject malformed bounds, unsupported primitive modes/types, non-finite values, out-of-range indices, cycles, and geometry over the documented caps.
2. Traverse only the selected scene roots, compose parent and local transforms, apply world transforms to triangle positions, and use the shared surface sampler.
3. Add one small generated-GLB test covering a transformed indexed triangle and one malformed-bounds case.
4. Run `GOMAXPROCS=2 go test -p 2 ./cmd/bake-l2a1`.
5. Bake `/tmp/l2a1-sar-source.glb` to `animations/models/l2a1-sar.bin`; confirm 8,831 source triangles, 24,000 samples, and 288,000 output bytes.

### Task 3: Share the braille mesh renderer and add the L2A1 animation

**Files:**
- Create: `animations/meshspin.go`
- Create: `animations/l2a1_sar.go`
- Modify: `animations/ak47.go`
- Test: `animations/ak47_test.go`
- Create: `animations/l2a1_sar_test.go`

**Steps:**

1. Move common mesh-point decoding, deterministic cloud assignment, 12-second phase calculation, palette shading ramp, projection/depth rendering, and footer placement into private helpers.
2. Keep AK's art parsing, target sampling, and 14-second cycle in `AK47Animation`; preserve its exported constructors and methods.
3. Add `NewL2A1SAREffect` implementing `Animation`, `Resize`, and `SetCellAspect`, with the 12-second cycle and on-screen attribution.
4. Keep the 24,000-point bake and apply an L2A1-only 1.12 display scale after comparing higher sample counts; the scale must retain margin through the rotation and particle phase.
5. Test cycle endpoints, deterministic render, increased occupied braille detail without edge clipping, loop/reset behavior, resize behavior, attribution, and AK regression invariants.
6. Run `GOMAXPROCS=2 go test -p 2 ./animations`.

### Task 4: Register and expose the effect

**Files:**
- Modify: `animations/registry.go`
- Modify: `cmd/syscgo/main.go`
- Test: `cmd/syscgo/cli_effects_test.go`
- Modify: `tui/animfactory.go`
- Modify: `tui/animfactory_test.go`

**Steps:**

1. Add the standalone `l2a1-sar` registry metadata and CLI dispatch case.
2. Add a TUI factory case; the picker already derives effect names from the registry.
3. Verify registry metadata, CLI availability, TUI construction, and motion with focused tests.
4. Run `GOMAXPROCS=2 go test -p 2 ./cmd/syscgo ./tui`.

### Task 5: Record attribution and verify the integrated effect

**Files:**
- Modify: `animations/models/README.md`
- Modify: `README.md`

**Steps:**

1. Record the exact supplied attribution, CC BY 4.0 license, source archive/export hash, derived-data changes, and baker reproduction command.
2. Add a concise CLI example for `l2a1-sar`.
3. Run `GOMAXPROCS=2 go test -p 2 ./...` and `go vet -p 2 ./...`.
4. Run the effect briefly in a PTY and inspect the braille silhouette and credit footer.
