# Australian L2A1 SAR Braille Effect Design

## Goal

Add a standalone `l2a1-sar` animation that renders the credited Australian L2A1 SAR model in braille, using the same two-turn, particle-swirl, and reconstruction cycle as the AK-47 effect.

## Existing behavior and constraints

`AK47Animation` owns a 12-second timeline: two eased rotations run from 0 to 8 seconds, then the mesh disperses into a deterministic particle cloud and reforms by 12 seconds. It projects fixed, quantized surface samples into the depth and colour buffers owned by `LogoSpinAnimation`, then uses that canvas's braille compositor. Its artwork variant also uses the same renderer and has a 14-second cycle. The CLI has explicit effect dispatch, the TUI has a separate effect factory, and both obtain names from the animation registry.

The source export is Sketchfab's GLB for 52 triangle primitives in 94 scene nodes. Mesh nodes include local and parent transforms, so the baker must traverse the selected scene and apply the complete node transforms before sampling. The model's longest axis runs along camera depth at source rotation zero, so the effect starts with a fixed quarter-turn to present a side view before its two animated turns. The pinned GLB is 942,956 bytes with SHA-256 `88748cab888526b440066e32feec49a46498171101f26b86ccc92e979dcc258d`.

## Design

- Add a small internal mesh package that owns triangle surface sampling and the existing 12-byte sample format (24,000 signed 16-bit position/normal records). Reuse it from the AK and L2A1 offline bakers.
- Add a standard-library-only GLB baker. Validate the pinned source digest, GLB chunk boundaries, selected scene graph, accessors, indices, primitive modes, finite transforms, and geometry limits. Apply glTF column-major node matrices (or composed translation/rotation/scale) to positions before passing triangles to the shared sampler. Generate face normals after world transforms, matching the AK sample convention.
- Extract the common rotation, projection, z-buffer, shading, braille composition, and credit-row behavior into animation helpers. Keep AK artwork selection and its 14-second artwork timing intact; the L2A1 effect uses the standard 12-second mesh cycle.
- Add the `l2a1-sar` registry entry, CLI dispatch, TUI factory constructor, and focused cycle/render/registration tests.
- Document the supplied attribution and CC BY 4.0 license beside the derived model data. Display `L2A1 SAR / notcplkerry / CC BY 4.0` in the same footer position and style as the AK effect.

## Detail refinement

The L2A1 retains the 24,000-point surface bake. Trials at 36,000 and 48,000
samples added at most 4.2% occupied braille dots on 80x24, 120x40, and 160x48
grids; the 48,000-point render took roughly 73–86% longer in the local
measurement. Increasing the sample count therefore does little for visible
detail at common terminal sizes.

Instead, the shared renderer accepts a fixed per-model display scale. The
L2A1 uses 1.12 to occupy more braille positions while keeping the AK at 1.0
and applying the same scale to its mesh and particle cloud. This produced
about 20–23% more occupied dots on the measured grids without increasing
sample or per-point work. A 250ms sweep of the full 12-second L2A1 cycle at
80x24 keeps the model inside the outer screen cells.

## Compatibility

The existing AK effect, artwork constructors, and 12/14-second phase behavior remain externally unchanged. No runtime GLB, Blender, graphics, or third-party parser dependency is added; the GLB is only needed to reproduce the committed baked samples.
