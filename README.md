# sysc-Go

Terminal animation library for Go. Pure Go animations ready to use in your TUI applications.

## Features

- **Rain Effect** - ASCII character rain effect
- **Matrix Rain** - Classic Matrix digital rain
- **Fireworks** - Particle-based fireworks display
- **Fire Effect** - DOOM PSX-style fire animation
- **Decrypt Effect** - Movie-style text decryption animation
- **Pour Effect** - Characters pour into position from different directions
- **Print Effects** - Typewriter-style text rendering (library use only)

## Installation

### CLI Tool

**Via AUR (Arch Linux):**
```bash
yay -S syscgo
```

**Via Go:**
```bash
go install github.com/Nomadcxx/sysc-Go/cmd/syscgo@latest
```

### As Library

```bash
go get github.com/Nomadcxx/sysc-Go
```

## Quick Start

Run any animation directly from command line:

```bash
# Rain effect with Tokyo Night theme
syscgo -effect rain -theme tokyo-night

# Matrix rain with Nord theme for 30 seconds
syscgo -effect matrix -theme nord -duration 30

# Fire effect with Dracula theme (infinite loop)
syscgo -effect fire -theme dracula -duration 0

# Decrypt effect with Catppuccin theme
syscgo -effect decrypt -theme catppuccin -file message.txt -duration 15

# Pour effect with Tokyo Night theme
syscgo -effect pour -theme tokyo-night -duration 10
```

**Available themes:** dracula, gruvbox, nord, tokyo-night, catppuccin, material, solarized, monochrome, transishardjob

## Effect Showcase

### Rain
![ASCII Rain](assets/rain.gif)

### Matrix
![Matrix Rain](assets/matrix.gif)

### Fireworks
![Fireworks](assets/fireworks.gif)

### Fire
![Fire Effect](assets/fire.gif)

### Decrypt
![Decrypt Effect](assets/decrypt.gif)

### Pour
![Pour Effect](assets/pour.gif)

### Print *(in development)*

## Demo

Run the interactive demo to see all animations:

```bash
cd examples/demo
go run .
```

## Documentation

See [GUIDE.md](GUIDE.md) for detailed usage.

## License

MIT