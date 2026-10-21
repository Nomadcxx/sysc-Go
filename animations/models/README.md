# AK-47 model data

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
go run ./cmd/bake-l2a1 /path/to/low-poly_australian_l2a1_sar.glb animations/models/l2a1-sar.bin
```

The converter checks the pinned source digest before conversion. The binary
layout matches `ak47.bin`: 24,000 records of six little-endian int16 values
(position XYZ, then normal XYZ), divided by 32767 when read; total size is
288,000 bytes.
