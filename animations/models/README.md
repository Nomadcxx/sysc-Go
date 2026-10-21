# Embedded braille model data

## AK-47 model data

"AK-47" (https://skfb.ly/oyAsI) by Lokeig is licensed under
[Creative Commons Attribution-NonCommercial 4.0 International](https://creativecommons.org/licenses/by-nc/4.0/).

`ak47.bin` is derived from this model and retains CC BY-NC 4.0. The code
license does not relicense this asset. Commercial use requires separate
permission from the creator. No endorsement by Lokeig is implied.

Source archive: owner-supplied `ak-47.zip`, member `source/AKM.fbx`.
FBX SHA-256: `152f608c545ebc691a711c72fdb3ea7637304c4b83d472c4217fd9f4571639b4`.
The Sketchfab source page identifies CC BY-NC 4.0, checked 2026-10-12.

Changes: apply the source part transforms, convert Y-up to screen Y-down,
center and normalize the longest dimension, triangulate faces, sample 24,000
surface points with face normals and quantize to signed 16-bit coordinates.
The effect adds theme shading, braille projection, rotation and particle morph.
Textures and source animation tracks are unused. The source FBX stays local.

Reproduce from the sysc-Go root:

```sh
unzip -p /path/to/ak-47.zip source/AKM.fbx > /tmp/AKM.fbx
CGO_ENABLED=0 go run ./cmd/bake-ak47 /tmp/AKM.fbx animations/models/ak47.bin
timeout 90s env CGO_ENABLED=0 GOMAXPROCS=2 AK47_SOURCE=/tmp/AKM.fbx go test -count=1 ./cmd/bake-ak47 -run TestAK47Bake
```

Binary layout: 24,000 records, each containing six little-endian int16 values
(position XYZ followed by normal XYZ), divided by 32767 when read. Total:
288,000 bytes. The converter checks the source digest before conversion.

`RAMA_SYSC.txt` is the owner-supplied RAMA block artwork, embedded as the
built-in morph target. Custom artwork is sampled from occupied character cells.
Both variants add a two-second artwork hold between formation and reconstruction.

## Australian L2A1 SAR model data

"low-poly Australian L2A1 SAR" (https://skfb.ly/oErMB) by notcplkerry is
licensed under Creative Commons Attribution. The license is [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/).

`l2a1-sar.bin` is derived from this model and retains CC BY 4.0. The code
license does not relicense this asset. Commercial use is allowed by the model
license. No endorsement by notcplkerry is implied.

The supplied `low-poly-australian-l2a1-sar.zip` contains
`source/L2A1_aussie.blend`. The matching `low-poly_australian_l2a1_sar.zip`
contains `scene.gltf`, `scene.bin`, and the source license. The GLB used for
the bake is `low-poly_australian_l2a1_sar.glb`, SHA-256
`88748cab888526b440066e32feec49a46498171101f26b86ccc92e979dcc258d`.

Changes: traverse the selected glTF scene and apply all parent and node
transforms, omit the `308 casing_30` and `308 cartridge_31` node subtrees,
convert Y-up to screen Y-down, center and normalize the longest dimension,
triangulate the remaining indexed faces, compute face normals, sample 24,000
surface points, and quantize positions and normals to signed 16-bit values.
The filtered scene contains 8,439 triangles; the two omitted subtrees contain
392 triangles. The effect adds theme shading, braille projection, two
rotations, and a particle swirl/reconstruction. Materials, textures, and
animation tracks are unused. The source export is not included in the
repository.

Reproduce from the sysc-Go root with the source GLB available locally:

```sh
go run ./cmd/bake-model -sha256 88748cab888526b440066e32feec49a46498171101f26b86ccc92e979dcc258d \
  -exclude-node '308 casing_30' -exclude-node '308 cartridge_31' \
  /path/to/low-poly_australian_l2a1_sar.glb animations/models/l2a1-sar.bin
```

The converter checks the supplied source digest before conversion. The binary
layout matches `ak47.bin`: 24,000 records of six little-endian int16 values
(position XYZ, then normal XYZ), divided by 32767 when read; total size is
288,000 bytes.

## Rotating statue model data

Each statue source stays outside the repository. The `.bin` files contain
96,000 sampled positions and normals from the selected glTF scene. At 491×90,
the denser bake adds visible braille detail for all six models and meets the
30 ms per-frame budget. Each file is 1,152,000 bytes.
The source digest, triangle count, creator, and Sketchfab license identify
each bake.

| Data file | Model and creator | Triangles | Source | License | Source SHA-256 |
|---|---|---:|---|---|---|
| `jesus-christ-statue.bin` | Jesus Christ Statue by amcgi | 499,992 | [Sketchfab](https://sketchfab.com/3d-models/jesus-christ-statue-375d626ccd524c439e9dc81c1033828e) | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) | `0ef5d3ea1b588b9abb71a88d844d9a6b87eae2b559ed64d0aa9bc04d9a9c16bc` |
| `jason.bin` | Jason by noe-3d.at | 558,084 | [Sketchfab](https://sketchfab.com/3d-models/jason-e7da6b09d300488e93fbe384412b8b18) | [CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/) | `7339eed9258b748b45743a8188abaa4bbf913d0a2283f781444ba6cf9e6cb868` |
| `nymphe-der-flora.bin` | Nymphe der Flora by noe-3d.at | 505,493 | [Sketchfab](https://sketchfab.com/3d-models/nymphe-der-flora-f1f4fca883954c0e9ac2bda9e865b144) | [CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/) | `b5cd6a31d45079ae339be641d7933bd738a1af12591b308899b1a1801d78b046` |
| `lowe.bin` | Löwe by noe-3d.at | 594,880 | [Sketchfab](https://sketchfab.com/3d-models/lowe-7e78d1f7fe04481697f1388ebba6ef9f) | [CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/) | `8a111419f69654cba07d83ae245ce8ede831c29b0933534f7f065a15d00d6e81` |
| `roman-gladiator.bin` | Sculpture of a Roman gladiator by Virtual Museums of Małopolska | 346,386 | [Sketchfab](https://sketchfab.com/3d-models/sculpture-of-a-roman-gladiator-ca42e593d8aa4b2894dc2b288f595af5) | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) | `6265957fb00163ec6a1c07b8fe14da72115647598a1c1653ab4491393e974acf` |
| `grabfigur.bin` | Grabfigur by noe-3d.at | 672,600 | [Sketchfab](https://sketchfab.com/3d-models/grabfigur-0c69e48471de4467aba03c5767164237) | [CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/) | `abfcb27fd5da0b9deb7d41239d9321057213da3042ff9f86d5edc5de2792018a` |

The four CC BY-NC 4.0 models prohibit commercial use without separate
permission from their creators. The repository's code license does not change
the model licenses. Each effect credits the model title, creator, and license
in its braille frame.

Rebake each model from the sysc-Go root with its source GLB available locally:

```sh
go run ./cmd/bake-model -sha256 0ef5d3ea1b588b9abb71a88d844d9a6b87eae2b559ed64d0aa9bc04d9a9c16bc -samples 96000 /path/to/jesus_christ_statue.glb animations/models/jesus-christ-statue.bin
go run ./cmd/bake-model -sha256 7339eed9258b748b45743a8188abaa4bbf913d0a2283f781444ba6cf9e6cb868 -samples 96000 /path/to/jason.glb animations/models/jason.bin
go run ./cmd/bake-model -sha256 b5cd6a31d45079ae339be641d7933bd738a1af12591b308899b1a1801d78b046 -samples 96000 /path/to/nymphe_der_flora.glb animations/models/nymphe-der-flora.bin
go run ./cmd/bake-model -sha256 8a111419f69654cba07d83ae245ce8ede831c29b0933534f7f065a15d00d6e81 -samples 96000 /path/to/lowe.glb animations/models/lowe.bin
go run ./cmd/bake-model -sha256 6265957fb00163ec6a1c07b8fe14da72115647598a1c1653ab4491393e974acf -samples 96000 /path/to/sculpture_of_a_roman_gladiator.glb animations/models/roman-gladiator.bin
go run ./cmd/bake-model -sha256 abfcb27fd5da0b9deb7d41239d9321057213da3042ff9f86d5edc5de2792018a -samples 96000 /path/to/grabfigur.glb animations/models/grabfigur.bin
```

The baker applies selected-scene node transforms and Y-up conversion, samples
the surface by triangle area, and stores signed 16-bit positions and normals.
It accepts up to 64 MiB and 1,000,000 triangles; the sample codec caps bakes at
96,000 records. The renderer rotates each model horizontally around its upright
axis, with a fixed camera angle and no pitch or roll.
