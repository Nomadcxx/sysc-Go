# Fighter Jet Rotation Effects Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add four credited fighter-aircraft effects that loop through level horizontal rotation and appear in sysc-Go, sysc-terminal, and sysc-greet.

**Architecture:** Reuse sysc-Go's offline GLB baker, sample codec, rotating-model catalog, and effect registry. Extend the existing consumers after sysc-Go merges; sysc-shell reads the updated effect list from sysc-terminal and needs no separate registry.

**Tech Stack:** Go 1.26, GLB, Git worktrees, GitHub CLI.

---

## Task 1: Add failing sysc-Go coverage for the aircraft

**Files:**
- Modify: `animations/rotating_model_test.go`
- Modify: `animations/logospin_test.go`

**Steps:**

1. Add `f-14-tomcat`, `f-15e-strike-eagle`, `f-22-raptor`, and `fa-18e-super-hornet` to `rotatingModelTestIDs`.
2. Add an exact-credit test table for the four models:
   - `F-14 Tomcat - Fighter Jet - Free / bohmerang / CC BY-NC-SA 4.0`
   - `F-15E Strike Eagle - Fighter Jet - Free / bohmerang / CC BY-NC-SA 4.0`
   - `F-22 Raptor - Fighter Jet - Free / bohmerang / CC BY-NC-SA 4.0`
   - `F/A-18E Super Hornet (USA) / 42manako / CC BY-NC 4.0`
3. Set the library-version assertion to `1.0.10`.
4. Run `go test ./animations -run 'TestRotatingModel|TestLibraryVersion' -count=1`.
5. Confirm the new model cases fail because the catalog has no aircraft entries yet.
6. Commit the tests with `test: cover rotating fighter model effects`.

## Task 2: Bake the four pinned GLBs

**Files:**
- Create: `animations/models/f-14-tomcat.bin`
- Create: `animations/models/f-15e-strike-eagle.bin`
- Create: `animations/models/f-22-raptor.bin`
- Create: `animations/models/fa-18e-super-hornet.bin`

Run these commands from the sysc-Go worktree:

```sh
go run ./cmd/bake-model -sha256 3b2754fafc2cea9fbc62352575337b0d0315fe805c08fa018caca35321df16e7 -samples 96000 /home/nomadx/Downloads/sysc-assets/f-14_tomcat_-_fighter_jet_-_free.glb animations/models/f-14-tomcat.bin
go run ./cmd/bake-model -sha256 d8fbe2d534d8bb4f120e9f25a63e6b6ef67bb59ca9c9383dc1268ec15b3fce31 -samples 96000 /home/nomadx/Downloads/sysc-assets/f-15e_strike_eagle_-_fighter_jet_-_free.glb animations/models/f-15e-strike-eagle.bin
go run ./cmd/bake-model -sha256 fac05850b4b0025a78b9679fc3b6bfd71f38082eb35f0d6420c9a5794d3cf3d9 -samples 96000 /home/nomadx/Downloads/sysc-assets/f-22_raptor_-_fighter_jet_-_free.glb animations/models/f-22-raptor.bin
go run ./cmd/bake-model -sha256 0910707eeb04d85f35c55103b129df6aff961e32cfe3e4d358f75d73b3783273 -samples 96000 /home/nomadx/Downloads/sysc-assets/fa-18e_super_hornet_usa.glb animations/models/fa-18e-super-hornet.bin
```

Each command must report 96,000 samples and create a 1,152,000-byte file. The baker checks the source digest before converting it.

## Task 3: Register the models in sysc-Go

**Files:**
- Modify: `animations/rotating_model.go`
- Modify: `animations/registry.go`
- Modify: `animations/logospin_test.go`

**Steps:**

1. Add `go:embed` variables for the four baked files.
2. Add catalog entries with the exact credits from Task 1 and provisional start angles and display scales for each plane.
3. Add a per-model start angle to `rotatingModelSpec` and apply it to the shared yaw phase. Keep the zero value for existing catalog entries.
4. Add four scene entries to `EffectRegistry`, mark them as version `1.0.10`, and describe their horizontal rotation and licenses.
5. Set `LibraryVersion` to `1.0.10`.
7. Run `go test ./animations -run 'TestRotatingModel|TestLibraryVersion' -count=1`.
8. Confirm exact credits, deterministic rendering, shade separation, full-loop continuity, sample-data sizes, and the 30 ms render budget.
9. Commit with `feat: add rotating fighter jet effects`.

## Task 4: Fit and document the four models

**Files:**
- Modify: `animations/rotating_model.go`
- Modify: `animations/models/README.md`
- Modify: `README.md`

**Steps:**

1. Run the full-yaw frame-bounds test at `80x24` and `491x90`; tune each model's start angle and `displayScale` until the aircraft keeps a margin and the credit row remains clear.
2. Check the generated frame at the start and quarter turns for recognizable aircraft profiles and readable theme shading.
3. Add a model-data table with the exact title, creator, Sketchfab URL, source digest, triangle count, and license from the design document.
4. Add reproduction commands for the four bakes to `animations/models/README.md`.
5. Add the four effect names and continuous-rotation examples to the root `README.md`.
6. Run `go test ./animations -run 'TestRotatingModel|TestLibraryVersion' -count=1` and `go test ./tui ./cmd/syscgo`.
7. Commit with `docs: document fighter jet effects`.

## Task 5: Review and merge sysc-Go

**Files:**
- No additional files.

**Steps:**

1. Run `go test ./animations -run 'TestRotatingModel|TestLibraryVersion' -count=1`, `go test ./tui ./cmd/syscgo`, and `git diff --check refs/remotes/origin/master...HEAD`.
2. Review the complete diff for source hashes, exact credit strings, display bounds, and unchanged AK-47/L2A1 behavior.
3. Push `feat/fighter-rotation-effects` and open a PR against `master`.
4. Review the PR diff and CI results manually; merge only when LGTM.
5. Fetch `origin/master` and record the merged sysc-Go commit for consumer dependency updates.

## Task 6: Add the effects to sysc-terminal

**Files:**
- Modify: `internal/effect/effect.go`
- Modify: `internal/effect/effect_test.go`
- Modify: `README.md`
- Modify: `go.mod`
- Modify: `go.sum`

**Steps:**

1. Run `mkdir -p /home/nomadx/.config/superpowers/worktrees/sysc-terminal`.
2. Create `/home/nomadx/.config/superpowers/worktrees/sysc-terminal/fighter-rotation-effects` from `refs/remotes/origin/main` on branch `feat/fighter-rotation-effects` with `git -C /home/nomadx/sysc-terminal worktree add -b feat/fighter-rotation-effects /home/nomadx/.config/superpowers/worktrees/sysc-terminal/fighter-rotation-effects refs/remotes/origin/main`.
3. Update the sysc-Go dependency with `go get github.com/Nomadcxx/sysc-Go@<merged-sysc-Go-commit>` and run `go mod tidy`.
4. Add the four IDs to `isRotatingModel` and extend `TestRotatingStatuesListConstructAndKeepCredit` with the four exact credit strings.
5. Add the four models and license terms to the rotating-model section and attribution table in `README.md`.
6. Run `go test ./internal/effect -run '^TestRotatingStatuesListConstructAndKeepCredit$' -count=1`.
7. Run `go run ./cmd/sysc-terminal --list` and confirm each ID appears as `effect <id> 0`.
8. Run `go test ./internal/effect` and commit with `feat: expose fighter jet effects`.
9. Push the branch, open a PR against `main`, review manually, and merge when LGTM.

## Task 7: Add the backgrounds to sysc-greet

**Files:**
- Modify: `cmd/sysc-greet/backgrounds.go`
- Modify: `cmd/sysc-greet/rotating_statue_background_test.go`
- Modify: `THIRD_PARTY_NOTICES.md`
- Modify: `go.mod`
- Modify: `go.sum`

**Steps:**

1. Run `mkdir -p /home/nomadx/.config/superpowers/worktrees/sysc-greet`.
2. Create `/home/nomadx/.config/superpowers/worktrees/sysc-greet/fighter-rotation-effects` from `refs/remotes/origin/master` on branch `feat/fighter-rotation-effects` with `git -C /home/nomadx/sysc-greet worktree add -b feat/fighter-rotation-effects /home/nomadx/.config/superpowers/worktrees/sysc-greet/fighter-rotation-effects refs/remotes/origin/master`.
3. Update the sysc-Go dependency with `go get github.com/Nomadcxx/sysc-Go@<merged-sysc-Go-commit>` and run `go mod tidy`.
4. Add the four model IDs and user-facing names to `rotatingModelBackgrounds`.
5. Extend `TestRotatingStatueBackgroundsMenuAndLifecycle` with the IDs, labels, and exact credits.
6. Add source links and license terms to `THIRD_PARTY_NOTICES.md`; identify the three CC BY-NC-SA bakes as share-alike data.
7. Run `go test ./cmd/sysc-greet -run '^TestRotatingStatueBackgroundsMenuAndLifecycle$' -count=1`.
8. Run `go test ./cmd/sysc-greet` and commit with `feat: add fighter jet backgrounds`.
9. Push the branch, open a PR against `master`, review manually, and merge when LGTM.

## Task 8: Verify sysc-shell discovers the IDs

**Files:**
- No source changes expected.

**Steps:**

1. Run `mkdir -p /home/nomadx/.config/superpowers/worktrees/sysc-shell`.
2. Create `/home/nomadx/.config/superpowers/worktrees/sysc-shell/fighter-rotation-effects` from `refs/remotes/origin/main` on branch `feat/fighter-rotation-effects` with `git -C /home/nomadx/sysc-shell worktree add -b feat/fighter-rotation-effects /home/nomadx/.config/superpowers/worktrees/sysc-shell/fighter-rotation-effects refs/remotes/origin/main`.
3. From the sysc-terminal worktree, run:

```sh
go run ./cmd/sysc-terminal --list | rg '^effect (f-14-tomcat|f-15e-strike-eagle|f-22-raptor|fa-18e-super-hornet) 0$'
```

4. Run `go test ./internal/wallpaper -run '^TestEngineRefreshTerminalCatalogReadsCurrentList$' -count=1` from the clean sysc-shell worktree.
5. Confirm the Terminal Art panel lists each ID from the runtime catalog.
