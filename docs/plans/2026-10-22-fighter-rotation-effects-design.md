# Fighter jet rotation effects

## Goal

Add four attributed fighter-aircraft effects to the existing rotating-model
family: F-14 Tomcat, F-15E Strike Eagle, F-22 Raptor, and F/A-18E Super Hornet.
Each aircraft turns around its vertical axis in a level loop that lasts 12
seconds. Set pitch and roll to zero. Show source credit in the braille footer.
Keep flight-path motion and particle transitions out of this tranche.

## Existing path

sysc-Go bakes GLB geometry into bounded position and normal samples and renders
the rotating-model catalog with theme shading. Its registry feeds the CLI and
TUI. sysc-terminal uses the shared registry. sysc-greet has a generic
rotating-model background path. The sysc-shell Terminal Art panel reads
effects from `sysc-terminal --list`; the new IDs will appear through that path.

Reuse the existing baker, sample codec, rotating-model animation, and app
integration paths. Keep source GLBs outside the repositories. Commit the baked
data, source hashes, and license notices. Use 96,000 samples per model, the
codec maximum of 1,152,000 bytes. The four models add 4,608,000 bytes to
sysc-Go's embedded data.

## Design choice

Use one shared rotating-model catalog for all four aircraft. Each effect uses
the approved continuous level yaw loop.

Use the source yaw for the first frame and set display scale per model. Keep
each full turn inside the render area with room for attribution. Keep camera
pitch and roll fixed. Check the maximum sample count against the large-frame
render budget.

## Model source and license

The source GLBs are in `/home/nomadx/Downloads/sysc-assets`. Their embedded
metadata and Sketchfab model API list the same titles, creators, and licenses.
Each source file fits the baker's 96 MiB input limit, and each mesh triangle
total fits its 2.1 million triangle limit.

| Effect ID | Source title and creator | Triangles | License | Source URL | Source SHA-256 |
|---|---|---:|---|---|---|
| `f-14-tomcat` | F-14 Tomcat - Fighter Jet - Free by bohmerang | 43,154 | CC BY-NC-SA 4.0 | https://sketchfab.com/3d-models/f-14-tomcat-fighter-jet-free-497f77c311b5411d9b567d8707a81c33 | `3b2754fafc2cea9fbc62352575337b0d0315fe805c08fa018caca35321df16e7` |
| `f-15e-strike-eagle` | F-15E Strike Eagle - Fighter Jet - Free by bohmerang | 41,808 | CC BY-NC-SA 4.0 | https://sketchfab.com/3d-models/f-15e-strike-eagle-fighter-jet-free-fff7d75490474e9b964d90cc031c8d01 | `d8fbe2d534d8bb4f120e9f25a63e6b6ef67bb59ca9c9383dc1268ec15b3fce31` |
| `f-22-raptor` | F-22 Raptor - Fighter Jet - Free by bohmerang | 41,072 | CC BY-NC-SA 4.0 | https://sketchfab.com/3d-models/f-22-raptor-fighter-jet-free-508de5c48845456bb033fb267ebe1d1e | `fac05850b4b0025a78b9679fc3b6bfd71f38082eb35f0d6420c9a5794d3cf3d9` |
| `fa-18e-super-hornet` | F/A-18E Super Hornet (USA) by 42manako | 4,270 | CC BY-NC 4.0 | https://sketchfab.com/3d-models/fa-18e-super-hornet-usa-0eb69d3fd53945488245b7d8464ffaa5 | `0910707eeb04d85f35c55103b129df6aff961e32cfe3e4d358f75d73b3783273` |

The effect footer and each repository's third-party notices will credit each
source title and creator and name its license. Publish the three CC BY-NC-SA
bakes under those same terms. Publish the F/A-18E bake under CC BY-NC 4.0. The
code license leaves each model license in force.

## Integration

- Register all four IDs in sysc-Go's effect metadata and construct them through
  the rotating-model factory.
- Expose the IDs in sysc-terminal's `--list` catalog and rendering path.
- Add the aircraft labels to sysc-greet's rotating-model background menu and
  reuse its existing update/render lifecycle.
- Confirm sysc-shell's Terminal Art panel lists the new IDs from refreshed
  `sysc-terminal --list` output.
- Update model attribution documentation and consumer third-party notices.

## Acceptance

- Each ID renders deterministically, completes an exact full yaw loop, and
  shows its attribution and license in the footer.
- Every model remains inside the frame for the full turn at standard and large
  display sizes, with no pitch or roll.
- Each 96,000-sample effect meets the existing 30 ms per-frame budget at
  491x90.
- sysc-Go, sysc-terminal, and sysc-greet expose and render all four IDs; the
  sysc-shell Terminal Art panel discovers them from `--list`.
- Focused checks pass in every changed repository. Open PRs for changed
  repositories, review them manually, and merge only when LGTM.
