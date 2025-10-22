# Demo GIF Generation

VHS tape files for generating effect showcase GIFs.

## Install VHS

```bash
go install github.com/charmbracelet/vhs@latest
```

## Generate All GIFs

```bash
# Make sure syscgo binary is in project root
cd ..
go build -o syscgo ./cmd/syscgo/

# Generate all showcase GIFs
vhs demos/fire.tape
vhs demos/matrix.tape
vhs demos/fireworks.tape
vhs demos/rain.tape
```

GIFs will be created in `assets/` directory.

## Individual Generation

```bash
vhs demos/fire.tape       # Generate fire.gif
vhs demos/matrix.tape     # Generate matrix.gif
vhs demos/fireworks.tape  # Generate fireworks.gif
vhs demos/rain.tape       # Generate rain.gif
```

## Requirements

- VHS installed
- syscgo binary built in project root
- ffmpeg (VHS dependency)
