# Model surface shading

## Goal

Make the rotating mesh effects read with more shape and surface detail by
improving their normals and lighting. Keep the theme palette and current sample
counts.

## Current behavior

`cmd/bake-model` reads GLB positions and triangle indices, then discards the
source `NORMAL` attribute. `internal/mesh` places deterministic area-weighted
samples and stores one flat face normal per triangle. `animations/meshspin.go`
rotates those normals, applies one directional light, and maps the result to a
16-color theme ramp. The richer theme shader in `animations/logospin.go` already
provides shadow, diffuse, fill, rim, and gloss terms.

## Design

Extend the GLB bake path to read source vertex normals, apply the inverse
transpose of each node's linear transform and the Y-up to screen conversion,
then interpolate and normalize normals using the same barycentric weights as
the sampled positions. When a primitive has no usable normal attribute, keep
the current geometric face-normal behavior.

Use the resulting rotated normals with the existing theme lighting values and
the `logoShade` lighting terms. This adds broader fill and controlled edge and
highlight response while keeping the selected theme colors. Keep `SampleBytes`
at 12 bytes: the existing normal fields carry the improved data. Do not raise
sample counts in this change.

First inspect fixed-angle renders at the same terminal dimensions for a statue,
a hard-surface model, and an existing mesh effect. Check whether the lighting
still flattens any surfaces before deciding whether to sample source texture or
occlusion maps in a later change. Texture and material parsing are out of scope.

## Alternatives considered

- Change only the runtime light. This is small, but source face normals keep
  broad curves faceted.
- Preserve source normals and improve the runtime lighting. This addresses the
  bake and renderer together without increasing asset size, so it is selected.
- Bake source texture or occlusion detail into each sample. Some models contain
  useful maps, but this requires UV and material support and changes the baked
  record format. Revisit after the first visual comparison.

## Validation

- Test normal interpolation, normalization, and face-normal fallback in
  `internal/mesh`.
- Test GLB normal accessor decoding and normal transforms, including a
  non-uniform node scale and Y conversion, in `cmd/bake-model`.
- Test the mesh shader for theme-colored shade separation and deterministic
  output. Check fixed-angle model renders at standard and large terminal sizes.
- Keep the existing 30 ms frame budget for 96,000-sample models.
