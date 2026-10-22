# Model Surface Shading Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Improve surface shape and detail across GLB-backed mesh effects while keeping their theme palettes, sample counts, and 12-byte sample records.

**Architecture:** Carry source vertex normals from GLB accessors through node transforms and area sampling. Interpolate and normalize each sample normal, then shade it with the existing theme lighting values and `logoShade` terms. Re-bake the checked-in GLB sample files from their pinned source assets.

**Tech Stack:** Go, standard-library GLB parsing and math, `internal/mesh`, existing mesh renderer and embedded GLB-derived samples.

---

### Task 1: Interpolate vertex normals in the mesh sampler

**Files:**
- Modify: `internal/mesh/mesh.go`
- Modify: `cmd/bake-ak47/main.go`
- Test: `internal/mesh/mesh_test.go`
- Test: `cmd/bake-ak47/main_test.go`

**Step 1: Add a failing GLB-to-sample check**

Extend the GLB fixture with a `NORMAL` accessor. Parse and sample a triangle with distinct vertex normals; assert that its first sample has the expected interpolated, normalized normal. Also check that a source with no `NORMAL` accessor keeps its geometric face normal.

Run: `go test ./cmd/bake-model -run '^TestParseGLBInterpolatesAuthoredNormals$' -count=1`

Expected: the test fails because `parseGLB` currently drops `NORMAL`.

**Step 2: Add per-vertex normals to triangle input**

Extend the triangle input with three positions and three optional float32 normals per vertex. Add an `internal/mesh` test for interpolation and unit-length output. Update the FBX baker to initialize the new triangle shape while leaving its face-normal fallback intact. Use the same barycentric weights for position and normal interpolation. Normalize the interpolated normal; use the computed face normal if the interpolation has no usable direction.

**Step 3: Verify sampler behavior**

Run: `go test ./internal/mesh -count=1`

Expected: all sampler and codec tests pass. `SampleBytes` remains 12.

**Step 4: Commit**

```sh
git add internal/mesh/mesh.go internal/mesh/mesh_test.go
git commit -m 'feat: interpolate sampled mesh normals'
```

### Task 2: Read and transform GLB normal accessors

**Files:**
- Modify: `cmd/bake-model/main.go`
- Test: `cmd/bake-model/main_test.go`

**Step 1: Add failing GLB normal tests**

Extend the fixture builder to include a `NORMAL` accessor. Test FLOAT VEC3 normals, normalized signed SHORT VEC3 normals, missing-normal fallback, and a non-uniform node scale with the Y conversion. Check that the transformed normal agrees with the transformed triangle winding, including a reflected node scale.

Run: `go test ./cmd/bake-model -run 'TestParseGLB.*Normal' -count=1`

Expected: the new authored-normal tests fail because the parser ignores `NORMAL`.

**Step 2: Decode supported normal accessors**

Read glTF FLOAT VEC3 normals and normalized signed BYTE or SHORT VEC3 normals. Validate accessor type, count, bounds, alignment, finite values, and nonzero directions. Keep missing accessors on the face-normal path.

**Step 3: Transform normals correctly**

Apply the inverse transpose of the accumulated node transform, normalize, then negate Y for the screen-space conversion. When a node transform has no invertible normal transform, leave that primitive on the face-normal fallback path.

**Step 4: Pass vertex normals to the sampler**

Store each primitive's transformed normals alongside its triangle positions. Continue applying the current position and winding conversion.

**Step 5: Verify GLB parsing**

Run: `go test ./cmd/bake-model -count=1`

Expected: normal encoding, transform, reflection, missing-normal, source digest, and existing GLB validation tests pass.

**Step 6: Commit**

```sh
git add cmd/bake-model/main.go cmd/bake-model/main_test.go
git commit -m 'feat: preserve GLB surface normals'
```

### Task 3: Apply theme-aware mesh lighting

**Files:**
- Modify: `animations/meshspin.go`
- Modify: `animations/l2a1_sar.go`
- Modify: `animations/rotating_model.go`
- Modify: `animations/rotating_model_test.go`

**Step 1: Add a focused shading assertion**

Check that a fixed mesh frame produces theme-colored front, fill, and shadow values and keeps front and rear faces distinct.

Run: `go test ./animations -run 'TestMeshSpin.*Lighting' -count=1`

Expected: the new test fails against the current single-direction ramp.

**Step 2: Use the existing lighting terms**

Shade rotated normals through `logoShade` with the mesh sample depth and current theme lighting values. Remove the redundant per-instance 16-color ramp and lighting flag once the tests confirm the new path covers AK-47, L2A1, and rotating models.

**Step 3: Check model frames and runtime cost**

Run: `go test ./animations -run 'TestRotatingModelsShadeTheirSurfaces|TestMeshSpin.*Lighting|TestRotatingModelPerformanceBudget' -count=1`

Expected: theme shade separation remains visible and every 96,000-sample model stays below the existing 30 ms frame budget at 491x90. If the shader exceeds the budget, simplify its per-sample math before continuing.

**Step 4: Commit**

```sh
git add animations/meshspin.go animations/l2a1_sar.go animations/rotating_model.go animations/rotating_model_test.go
git commit -m 'feat: add richer lighting to mesh effects'
```

### Task 4: Re-bake GLB-backed model samples

**Files:**
- Modify: `animations/models/l2a1-sar.bin`
- Modify: `animations/models/jesus-christ-statue.bin`
- Modify: `animations/models/jason.bin`
- Modify: `animations/models/nymphe-der-flora.bin`
- Modify: `animations/models/lowe.bin`
- Modify: `animations/models/roman-gladiator.bin`
- Modify: `animations/models/grabfigur.bin`
- Modify: `animations/models/laocoon-and-his-sons.bin`
- Modify: `animations/models/napoleon-bust.bin`
- Modify: `animations/models/brandenburg-gate.bin`
- Modify: `animations/models/louis-xiv.bin`
- Modify: `animations/models/engel.bin`
- Modify: `animations/models/st-george.bin`
- Modify: `animations/models/santa-muerte.bin`
- Modify: `animations/models/aspasia.bin`
- Modify: `animations/models/st-nepomuk.bin`
- Modify: `animations/models/lion-statue.bin`
- Modify: `animations/models/f-14-tomcat.bin`
- Modify: `animations/models/f-15e-strike-eagle.bin`
- Modify: `animations/models/f-22-raptor.bin`
- Modify: `animations/models/fa-18e-super-hornet.bin`

Use the pinned SHA-256 values and documented sample counts in `animations/models/README.md`. The matching GLBs are in `/home/nomadx/Downloads/sysc-assets`. Preserve the L2A1 exclusions for `308 casing_30` and `308 cartridge_31`. Run the bakes sequentially because St George approaches the converter's triangle limit. Leave the FBX-derived `ak47.bin` unchanged.

**Step 1: Re-bake L2A1**

Run the documented command with `/home/nomadx/Downloads/sysc-assets/low-poly_australian_l2a1_sar.glb` and the two excluded node names.

Expected: 24,000 records and 288,000 bytes.

**Step 2: Re-bake the six rotating statues**

Run the six pinned commands from the rotating statue section using the matching files in `/home/nomadx/Downloads/sysc-assets`.

Expected: 96,000 records and 1,152,000 bytes per file.

**Step 3: Re-bake the ten historical and religious models**

Run the ten pinned commands from the historical and religious section using the matching files in `/home/nomadx/Downloads/sysc-assets`.

Expected: 96,000 records and 1,152,000 bytes per file.

**Step 4: Re-bake the four fighter models**

Run the four pinned commands from the fighter aircraft section using the matching files in `/home/nomadx/Downloads/sysc-assets`.

Expected: 96,000 records and 1,152,000 bytes per file.

**Step 5: Check the baked assets**

Run: `go test ./animations -run 'TestRotatingModelBakesUseTheBoundedVariableSampleCodec|TestL2A1.*' -count=1`

Expected: each embedded file retains its documented record count and codec size.

### Task 5: Document and verify the complete change

**Files:**
- Modify: `animations/models/README.md`
- Test: `internal/mesh/mesh_test.go`
- Test: `cmd/bake-model/main_test.go`
- Test: `animations/rotating_model_test.go`
- Test: `animations/l2a1_sar_test.go`

**Step 1: Update bake documentation**

Describe transformed, interpolated source normals and face-normal fallback. Keep attribution, source hashes, sample counts, and file sizes unchanged. Leave texture and occlusion maps documented as unused.

**Step 2: Run focused checks**

Run: `go test ./internal/mesh ./cmd/bake-model ./animations -count=1`

Expected: all changed-package tests pass, including the 30 ms model frame budget.

**Step 3: Review fixed-angle frames**

Render baseline and updated 100x32 frames at the start and quarter turn for one statue and one aircraft. Compare AK-47 and L2A1 at the same size to check the shared renderer. Existing tests cover 80x24 and 491x90 fit and the 30 ms large-frame budget.

**Step 4: Run repository checks**

Run: `go test ./...` and `git diff --check`.

Expected: the complete Go suite passes and the diff has no whitespace errors.

**Step 5: Commit documentation and baked model data**

```sh
git add animations/models/README.md animations/models/*.bin
git commit -m 'data: rebake models with source normals'
```

## Completion criteria

- GLB samples use normalized interpolated source normals where available and geometric face normals otherwise.
- The renderer uses theme-aware fill, edge, and highlight shading.
- The 21 GLB-derived files keep their current sample counts and sizes; `ak47.bin` stays unchanged.
- Fixed-angle frames show clearer form and surface relief at both supported test sizes.
- All tests pass, including the 30 ms render budget.
