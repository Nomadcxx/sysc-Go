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
It accepts up to 96 MiB and 2,100,000 triangles; the sample codec caps bakes
at 96,000 records. The renderer rotates each model horizontally around its
upright axis, with a fixed camera angle and no pitch or roll.

## Historical and religious rotating model data

Each source GLB stays outside the repository. These `.bin` files contain
96,000 sampled positions and normals each. Their credits appear in the effect
footer. Source licenses remain attached to each derived model: CC BY-SA data
keeps the same share-alike terms, and CC BY-NC models remain limited to
noncommercial use.

| Data file | Source title and creator | Triangles | Source | License | Source SHA-256 |
|---|---|---:|---|---|---|
| `laocoon-and-his-sons.bin` | Laocoon and His Sons by Rigsters | 50,106 | [Sketchfab](https://sketchfab.com/3d-models/laocoon-and-his-sons-649111a9a7b74ddab3937292be5545fc) | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) | `0d15d6bfb2cffe79d4da54adc9ff2b700486be158bd3bd43933cbb63dd82ba63` |
| `napoleon-bust.bin` | The bust of Napoleon Bonaparte by Virtual Museums of Małopolska | 264,708 | [Sketchfab](https://sketchfab.com/3d-models/the-bust-of-napoleon-bonaparte-a177bf0e121641bea6cf1d58ad3efc5b) | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) | `bdfb7e76642556b6fc92f651582464282f9d4b3717c31faaec1be66bbb2babf5` |
| `brandenburg-gate.bin` | Brandenburg Gate (9 Mb) by Resinnnn | 390,218 | [Sketchfab](https://sketchfab.com/3d-models/brandenburg-gate-9-mb-012f2cb87cfa47029808063da072ad5c) | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) | `e4c1854b6c9a5d71742e2eb388ab28b51a21cf742d6232811f32842ddb1ffb7a` |
| `louis-xiv.bin` | Louis XIV de France, Louvre, Paris by HoangHiepVu | 421,965 | [Sketchfab](https://sketchfab.com/3d-models/louis-xiv-de-france-louvre-paris-a0cc0e7eee384c99838dff2857b8158c) | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) | `46c94e4a1142c545e34e1c30c10dab24f1b66797daee76796fd417d399c0748f` |
| `engel.bin` | Engel by noe-3d.at | 559,786 | [Sketchfab](https://sketchfab.com/3d-models/engel-a625483175624e0ab6f7cd67ee53cc92) | [CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/) | `79f76b2d24f7917fb5d97437eae3ace96010f2599373dc2433af2223447895ce` |
| `st-george.bin` | Hl. Georg als Drachentöter by VIMUNE | 2,000,065 | [Sketchfab](https://sketchfab.com/3d-models/hl-georg-als-drachentoter-a211b6af71b249e1a03854d00fc41cbd) | [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/) | `3b681018dc1deb1ad843309ed771dab665ac2dcb18f17537a23e1c1158099b44` |
| `santa-muerte.bin` | Santa Muerte & Requiem for the Sketchfab by Mateusz Woliński | 155,410 | [Sketchfab](https://sketchfab.com/3d-models/santa-muerte-requiem-for-the-sketchfab-3fa36f79496a454995c64ca65b5d52b0) | [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/) | `fa8ca42e6c9b20df40304979603663b17624df08977fb9e82c0b8f58b010c65e` |
| `aspasia.bin` | Aspasia by noe-3d.at | 733,174 | [Sketchfab](https://sketchfab.com/3d-models/aspasia-6f0bd1e65a244cc78c5fe082c8e0d3d1) | [CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/) | `a4da809bdc37490a41906154e8adb0d31dc1ef6fd0e91bf5667cbbe9ffea204d` |
| `st-nepomuk.bin` | Hl. Nepomuk by noe-3d.at | 572,386 | [Sketchfab](https://sketchfab.com/3d-models/hl-nepomuk-9d65daaa9e07470cad5ce7c668b03434) | [CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/) | `f32dc9210b0edc93c455047b228eea1a9db93b789fcfdec654b98a770d4819b3` |
| `lion-statue.bin` | Lion Statue - optimized by Thomas Flynn | 28,488 | [Sketchfab](https://sketchfab.com/3d-models/lion-statue-optimized-62662d27c94d4994b2479b8de3a66ca7) | [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/) | `55e6f7784afa8159235ef9928635e19bb107cc808217ccd5640ba51496e58771` |

The converter applies scene-node transforms and Y-up conversion, triangulates
the selected scene, samples its surface area, then quantizes position and
normal values to signed 16-bit values. It ignores source textures and animation
tracks. The St George source has just over two million triangles, so the
offline converter's bounded input and triangle limits cover that model. Caesar
is excluded until its source attribution and license are available.

Rebake each model from the sysc-Go root with the source GLB available locally:

```sh
go run ./cmd/bake-model -sha256 0d15d6bfb2cffe79d4da54adc9ff2b700486be158bd3bd43933cbb63dd82ba63 -samples 96000 /path/to/laocoon_and_his_sons.glb animations/models/laocoon-and-his-sons.bin
go run ./cmd/bake-model -sha256 bdfb7e76642556b6fc92f651582464282f9d4b3717c31faaec1be66bbb2babf5 -samples 96000 /path/to/the_bust_of_napoleon_bonaparte.glb animations/models/napoleon-bust.bin
go run ./cmd/bake-model -sha256 e4c1854b6c9a5d71742e2eb388ab28b51a21cf742d6232811f32842ddb1ffb7a -samples 96000 /path/to/brandenburg_gate_9_mb.glb animations/models/brandenburg-gate.bin
go run ./cmd/bake-model -sha256 46c94e4a1142c545e34e1c30c10dab24f1b66797daee76796fd417d399c0748f -samples 96000 /path/to/louis_xiv_de_france_louvre_paris.glb animations/models/louis-xiv.bin
go run ./cmd/bake-model -sha256 79f76b2d24f7917fb5d97437eae3ace96010f2599373dc2433af2223447895ce -samples 96000 /path/to/engel.glb animations/models/engel.bin
go run ./cmd/bake-model -sha256 3b681018dc1deb1ad843309ed771dab665ac2dcb18f17537a23e1c1158099b44 -samples 96000 /path/to/hl._georg_als_drachentoter.glb animations/models/st-george.bin
go run ./cmd/bake-model -sha256 fa8ca42e6c9b20df40304979603663b17624df08977fb9e82c0b8f58b010c65e -samples 96000 /path/to/santa_muerte__requiem_for_the_sketchfab.glb animations/models/santa-muerte.bin
go run ./cmd/bake-model -sha256 a4da809bdc37490a41906154e8adb0d31dc1ef6fd0e91bf5667cbbe9ffea204d -samples 96000 /path/to/aspasia.glb animations/models/aspasia.bin
go run ./cmd/bake-model -sha256 f32dc9210b0edc93c455047b228eea1a9db93b789fcfdec654b98a770d4819b3 -samples 96000 /path/to/hl._nepomuk.glb animations/models/st-nepomuk.bin
go run ./cmd/bake-model -sha256 55e6f7784afa8159235ef9928635e19bb107cc808217ccd5640ba51496e58771 -samples 96000 /path/to/lion_statue_-_optimized.glb animations/models/lion-statue.bin
```

## Fighter aircraft rotating model data

Each source GLB stays outside the repository. Each embedded `.bin` file
contains 96,000 sampled positions and normals (1,152,000 bytes). The models
rotate continuously around their vertical axis with fixed pitch and roll; the
effect footer displays the model title, creator, and license.

| Data file | Source title and creator | Triangles | Source | License | Source SHA-256 |
|---|---|---:|---|---|---|
| `f-14-tomcat.bin` | F-14 Tomcat - Fighter Jet - Free by bohmerang | 43,154 | [Sketchfab](https://sketchfab.com/3d-models/f-14-tomcat-fighter-jet-free-497f77c311b5411d9b567d8707a81c33) | [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0/) | `3b2754fafc2cea9fbc62352575337b0d0315fe805c08fa018caca35321df16e7` |
| `f-15e-strike-eagle.bin` | F-15E Strike Eagle - Fighter Jet - Free by bohmerang | 41,808 | [Sketchfab](https://sketchfab.com/3d-models/f-15e-strike-eagle-fighter-jet-free-fff7d75490474e9b964d90cc031c8d01) | [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0/) | `d8fbe2d534d8bb4f120e9f25a63e6b6ef67bb59ca9c9383dc1268ec15b3fce31` |
| `f-22-raptor.bin` | F-22 Raptor - Fighter Jet - Free by bohmerang | 41,072 | [Sketchfab](https://sketchfab.com/3d-models/f-22-raptor-fighter-jet-free-508de5c48845456bb033fb267ebe1d1e) | [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0/) | `fac05850b4b0025a78b9679fc3b6bfd71f38082eb35f0d6420c9a5794d3cf3d9` |
| `fa-18e-super-hornet.bin` | F/A-18E Super Hornet (USA) by 42manako | 4,270 | [Sketchfab](https://sketchfab.com/3d-models/fa-18e-super-hornet-usa-0eb69d3fd53945488245b7d8464ffaa5) | [CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/) | `0910707eeb04d85f35c55103b129df6aff961e32cfe3e4d358f75d73b3783273` |

The three bohmerang model bakes remain under CC BY-NC-SA 4.0, including its
noncommercial and share-alike terms. The F/A-18E bake remains under CC BY-NC
4.0. The repository's code license does not change these model licenses.

Rebake each model from the sysc-Go root with its source GLB available locally:

```sh
go run ./cmd/bake-model -sha256 3b2754fafc2cea9fbc62352575337b0d0315fe805c08fa018caca35321df16e7 -samples 96000 /path/to/f-14_tomcat_-_fighter_jet_-_free.glb animations/models/f-14-tomcat.bin
go run ./cmd/bake-model -sha256 d8fbe2d534d8bb4f120e9f25a63e6b6ef67bb59ca9c9383dc1268ec15b3fce31 -samples 96000 /path/to/f-15e_strike_eagle_-_fighter_jet_-_free.glb animations/models/f-15e-strike-eagle.bin
go run ./cmd/bake-model -sha256 fac05850b4b0025a78b9679fc3b6bfd71f38082eb35f0d6420c9a5794d3cf3d9 -samples 96000 /path/to/f-22_raptor_-_fighter_jet_-_free.glb animations/models/f-22-raptor.bin
go run ./cmd/bake-model -sha256 0910707eeb04d85f35c55103b129df6aff961e32cfe3e4d358f75d73b3783273 -samples 96000 /path/to/fa-18e_super_hornet_usa.glb animations/models/fa-18e-super-hornet.bin
```
