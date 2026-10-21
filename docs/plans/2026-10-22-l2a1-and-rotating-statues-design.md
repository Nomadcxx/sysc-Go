# L2A1 detail refinement and rotating statue effects

## Goal

Remove the two unwanted ammunition groups from the L2A1 effect, improve the
rifle's visible detail, then add six attributed model effects that rotate
continuously around their upright vertical axis. The six new effects use a
fixed camera angle and no particle transition, pitch, or roll.

## Existing path

`l2a1-sar` and `ak-47` use an offline GLB baker, quantized surface samples,
`renderMeshSpin`, the braille canvas, and an effect registry used by sysc-Go's
CLI and TUI. sysc-terminal constructs those effects from the sysc-Go module.
The sysc-shell Terminal Art panel refreshes its cards from the installed
`sysc-terminal --list` catalog, so it does not need a second model registry.
sysc-greet has explicit background menu and update/render wiring.

The two L2A1 ammunition node groups are `308 casing_30` and
`308 cartridge_31`; each owns child nodes with the actual mesh primitives.
Omit both complete subtrees in the bake and preserve the source GLB.

The six supplied GLBs contain 346,386 to 672,600 triangles and range from
17.3 MB to 57.4 MB. Four models carry CC BY-NC 4.0; retain that restriction in
the notices and visible credits.

## Design

- Keep the L2A1 at 24,000 samples. Filter its two named node subtrees in the
  offline bake, then refine directional shading. At 80x24, the current 1.12
  scale already spans columns 1–78; 1.16 reaches the outer columns during the
  12-second cycle. Keep 1.12 rather than clipping the rifle further. Keep AK
  rendering and its cycle unchanged.
- Generalize the existing offline baker for the pinned GLBs rather than add a
  second parser. Raise bounded input limits to 64 MiB and 1,000,000 triangles.
  Keep the source exports out of the repository; commit only bounded sampled
  data, source hashes, and attribution.
- Start each statue at 48,000 samples (576,000 bytes). Let the shared codec and
  baker encode up to 96,000 samples; raise an individual model's count only if
  the denser bake adds visible detail at its chosen on-screen size. Keep the
  established 24,000 default for AK and L2A1. Qualify higher counts against
  sysc-terminal's 491x90 budget of 30 ms per frame / 25% of one CPU core.
- Use one public model-effect constructor keyed by six registry IDs. Each
  effect stays upright and turns at constant speed through a full horizontal
  yaw loop. Keep camera elevation fixed and avoid top or underside views unless
  the model needs them for identification. Choose each model's fit to use the
  available screen area while preserving a margin and the credit row. Use
  normal-based theme shading. No pitch, roll, particle cloud, swirl, or
  animation transitions.
- Register and expose the IDs in sysc-Go's CLI and TUI. Update sysc-terminal's
  constructor and sysc-greet's background menu/render lifecycle. Verify the
  existing shell catalog refresh and card rendering with the six reported IDs.
- Record each model's exact title, creator, source URL, and license. The
  no-derivative-data code license does not override the embedded model terms.

## Model attribution

| Effect ID | Source title | Creator | License |
|---|---|---|---|
| `jesus-christ-statue` | Jesus Christ Statue | amcgi | CC BY 4.0 |
| `jason` | Jason | noe-3d.at | CC BY-NC 4.0 |
| `nymphe-der-flora` | Nymphe der Flora | noe-3d.at | CC BY-NC 4.0 |
| `lowe` | Löwe | noe-3d.at | CC BY-NC 4.0 |
| `roman-gladiator` | Sculpture of a Roman gladiator | Virtual Museums of Małopolska | CC BY 4.0 |
| `grabfigur` | Grabfigur | noe-3d.at | CC BY-NC 4.0 |

## Acceptance

- The rebuilt L2A1 frame contains neither removed object group, has measurably
  more useful tonal separation, remains inside the grid through its full
  existing cycle, and retains its supplied credit.
- The six models render deterministically, stay upright, complete a smooth
  continuous yaw loop, keep all visible points inside the frame, and show the
  correct credit and license.
- AK output and both AK timelines remain unchanged.
- All six IDs construct in sysc-Go, sysc-terminal, and sysc-greet. The shell
  Terminal Art panel displays every ID from the refreshed `--list` catalog.
- Focused checks pass in each touched repository. Public PRs are manually
  reviewed and merged only when LGTM.
