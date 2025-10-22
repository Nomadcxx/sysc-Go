# sysc-Go

Terminal animation library for Go. Pure Go animations ready to use in your TUI applications.

## Features

- **Fire Effect** - DOOM PSX-style fire animation
- **Matrix Rain** - Classic Matrix digital rain
- **Fireworks** - Particle-based fireworks display
- **ASCII Rain** - Character-based rain effect
- **Ticker** - Scrolling text with colors
- **Print Effects** - Typewriter-style text rendering

## Installation

```bash
go get github.com/Nomadcxx/sysc-Go
```

## Quick Start

```go
package main

import (
    "github.com/Nomadcxx/sysc-Go/animations"
)

func main() {
    fire := animations.NewFire(80, 24)
    fire.Render()
}
```

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
