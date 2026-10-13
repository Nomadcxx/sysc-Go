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
