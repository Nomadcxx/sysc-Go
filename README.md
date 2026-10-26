<img src="assets/header.webp" alt="sysc-Go">

Terminal animation library for Go. Pure Go animations ready to use in your TUI applications.

<img src="assets/ak-47-text.webp" alt="AK-47 Text Effect" loading="lazy">

## Effects

Effect names and text requirements match the registry in [`animations/registry.go`](animations/registry.go).

### Animations
Standalone effects that do not require text input.

- **Fire** (`fire`): DOOM PSX-style fire
- **Matrix Rain** (`matrix`): classic digital rain
- **Rain** (`rain`): falling ASCII characters
- **Fireworks** (`fireworks`): particle-based fireworks
- **Beams** (`beams`): full-screen light beams
- **Aquarium** (`aquarium`): underwater scene with fish and sea life
- **Skull** (`skull`): skull with a Tron grid, sonar pulses, and scanline sweep
- **Sonar** (`sonar`): pulses radiating from the centre to the screen edges
- **Cracktro** (`cracktro`): demoscene animation with stars, raster bars, and wireframe
- **SYSC Logo** (`sysc-logo`): spinning 3D braille wordmark
- **Cross Logo** (`cross-logo`): spinning 3D braille Latin cross
- **Logo Morph** (`logo-morph`): rotating SYSC wordmark and cross morph
- **Higgs Field** (`higgs-field`): particle field with travelling waves and vortex motion
- **AK-47** (`ak-47`): Lokeig's braille model with a particle swirl and reconstruction
- **Australian L2A1 SAR** (`l2a1-sar`): notcplkerry's braille model with a particle swirl and reconstruction
- **Jesus Christ Statue** (`jesus-christ-statue`): amcgi's braille model with continuous horizontal rotation (CC BY 4.0)
- **Jason** (`jason`): noe-3d.at's braille model with continuous horizontal rotation (CC BY-NC 4.0)
- **Nymphe der Flora** (`nymphe-der-flora`): noe-3d.at's braille model with continuous horizontal rotation (CC BY-NC 4.0)
- **Löwe** (`lowe`): noe-3d.at's braille model with continuous horizontal rotation (CC BY-NC 4.0)
- **Sculpture of a Roman gladiator** (`roman-gladiator`): Virtual Museums of Małopolska's braille model with continuous horizontal rotation (CC BY 4.0)
- **Grabfigur** (`grabfigur`): noe-3d.at's braille model with continuous horizontal rotation (CC BY-NC 4.0)
- **Laocoon and His Sons** (`laocoon-and-his-sons`): Rigsters' rotating braille model (CC BY 4.0)
- **Napoleon Bust** (`napoleon-bust`): Virtual Museums of Małopolska's bust dissolves into a particle field and reforms (CC BY 4.0)
- **Brandenburg Gate** (`brandenburg-gate`): Resinnnn's rotating braille model (CC BY 4.0)
- **Louis XIV** (`louis-xiv`): HoangHiepVu's rotating braille model (CC BY 4.0)
- **Engel** (`engel`): noe-3d.at's rotating braille model (CC BY-NC 4.0)
- **St George** (`st-george`): VIMUNE's rotating braille model (CC BY-SA 4.0)
- **Santa Muerte** (`santa-muerte`): Mateusz Woliński's rotating braille model (CC BY-SA 4.0)
- **Aspasia** (`aspasia`): noe-3d.at's rotating braille model (CC BY-NC 4.0)
- **St Nepomuk** (`st-nepomuk`): noe-3d.at's rotating braille model (CC BY-NC 4.0)
- **Lion Statue** (`lion-statue`): Thomas Flynn's rotating braille model (CC0 1.0)
- **F-14 Tomcat** (`f-14-tomcat`): bohmerang's braille aircraft banks, climbs and turns (CC BY-NC-SA 4.0)
- **F-15E Strike Eagle** (`f-15e-strike-eagle`): bohmerang's braille aircraft banks, climbs and turns (CC BY-NC-SA 4.0)
- **F-22 Raptor** (`f-22-raptor`): bohmerang's braille aircraft banks, climbs and turns (CC BY-NC-SA 4.0)
- **F/A-18E Super Hornet** (`fa-18e-super-hornet`): 42manako's braille aircraft banks, climbs and turns (CC BY-NC 4.0)
- **Great Wave** (`great-wave`): Hokusai's wave with animated ASCII foam and water
- **Wasteland** (`wasteland`): animated ASCII scene after Valenberg

### Text Effects
Text effects animate ASCII input. The CLI uses `SYSC.txt` by default; pass `-file` to choose another file. Blackhole can also run in particle mode without text.

- **Fire Text** (`fire-text`): ASCII art consumed by rising flames
- **Matrix Art** (`matrix-art`): ASCII art revealed by Matrix-style streams
- **Rain Art** (`rain-art`): ASCII art revealed by falling rain
- **Beam Text** (`beam-text`): text with animated light beams
- **Ring Text** (`ring-text`): rotating coloured rings around ASCII art
- **Blackhole** (`blackhole`): text pulled into a swirling blackhole
- **Pour** (`pour`): characters fall into position from different directions
- **Print** (`print`): typewriter-style text rendering
- **Decrypt** (`decrypt`): scrambled text resolves into the supplied text
- **AK-47 Text** (`ak-47-text`): the braille model forms the selected ASCII art

## Installation

### CLI Tools

**Via one-line installer (fastest):**
```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-Go/master/install.sh | sudo bash
```

**Via interactive installer:**
```bash
# Clone and run the TUI installer (requires sudo for system-wide installation).
# Installs both syscgo and syscgo-tui binaries.
# Resolve Go in the current shell, then invoke that toolchain by absolute path.
git clone https://github.com/Nomadcxx/sysc-Go.git
cd sysc-Go
sudo "$(go env GOROOT)/bin/go" run ./cmd/installer/
```

**Via AUR (Arch Linux):**
```bash
yay -S syscgo
```

**Via Go:**
```bash
go install github.com/Nomadcxx/sysc-Go/cmd/syscgo@latest
go install github.com/Nomadcxx/sysc-Go/cmd/syscgo-tui@latest
```

> **Note:** The TUI binary includes the BIT fonts. `go install` does not install the sample text files; use the installer or AUR package for them.

### As Library

```bash
go get github.com/Nomadcxx/sysc-Go
```

## Quick Start

### Interactive TUI

The easiest way to browse and select animations is through the interactive TUI:

```bash
syscgo-tui
```

**Features:**
- Visual selector for all animations, themes, and durations
- Built-in ASCII art editor (BIT) with live preview
- 174 block-style fonts for creating ASCII art
- Real-time animation preview
- Export ASCII art to file (Ctrl+S)
- Navigate with arrow keys or vim keybindings (h/j/k/l)
- Instant theme switching

### Command Line

Run any animation directly from command line:

The default duration is 10 seconds. Set `-duration 0` to run until interrupted.

**Animations** (no text input required):
```bash
# Fire effect with Dracula theme (infinite loop)
syscgo -effect fire -theme dracula -duration 0

# Matrix rain with Nord theme for 30 seconds
syscgo -effect matrix -theme nord -duration 30

# Rain effect with Tokyo Night theme
syscgo -effect rain -theme tokyo-night

# Fireworks display
syscgo -effect fireworks -theme catppuccin -duration 20

# Beams effect (full-screen background)
syscgo -effect beams -theme nord -duration 0

# Aquarium effect (infinite)
syscgo -effect aquarium -theme dracula -duration 0

# AK-47: repeating 12-second cycle
syscgo -effect ak-47 -theme kanagawa -duration 0

# Australian L2A1 SAR: repeating 12-second cycle
syscgo -effect l2a1-sar -theme kanagawa -duration 0

# Higgs particle field: repeating 20-second cycle
syscgo -effect higgs-field -theme tokyo-night -duration 0

# Napoleon: dissolve into a particle field and reform over a 20-second cycle
syscgo -effect napoleon-bust -theme nord -duration 0

# Statue effects: continuous horizontal rotation
syscgo -effect jesus-christ-statue -theme kanagawa -duration 0
syscgo -effect jason -theme kanagawa -duration 0
syscgo -effect nymphe-der-flora -theme kanagawa -duration 0
syscgo -effect lowe -theme kanagawa -duration 0
syscgo -effect roman-gladiator -theme kanagawa -duration 0
syscgo -effect grabfigur -theme kanagawa -duration 0

# Fighter aircraft: banking, climbing and turning
syscgo -effect f-14-tomcat -theme kanagawa -duration 0
syscgo -effect f-15e-strike-eagle -theme kanagawa -duration 0
syscgo -effect f-22-raptor -theme kanagawa -duration 0
syscgo -effect fa-18e-super-hornet -theme kanagawa -duration 0

# Use selected artwork as the braille-particle target
syscgo -effect ak-47-text -file /path/to/art.txt -theme kanagawa -duration 0
```

**Text Effects** (require `-file` flag with text/ASCII art):
```bash
# Fire text effect with ASCII art
syscgo -effect fire-text -file logo.txt -theme rama -duration 0

# Matrix art effect
syscgo -effect matrix-art -file art.txt -theme eldritch -duration 20

# Rain art effect
syscgo -effect rain-art -file banner.txt -theme tokyo-night -duration 15

# Pour effect with text file
syscgo -effect pour -file message.txt -theme gruvbox -duration 10

# Print effect (typewriter style)
syscgo -effect print -file banner.txt -theme catppuccin -duration 15

# Beam text effect with auto-sizing and display mode
syscgo -effect beam-text -file header.txt -auto -display -theme nord -duration 5

# Ring text effect
syscgo -effect ring-text -file title.txt -theme eldritch -duration 10

# Blackhole effect with text
syscgo -effect blackhole -file text.txt -theme dark -duration 15
```

**Available themes:** dracula, gruvbox, nord, tokyo-night, catppuccin, material, solarized, monochrome, transishardjob, rama, eldritch, dark

**Text Effect Flags:**
- `-auto` - Auto-size canvas to fit text (beam-text only)
- `-display` - Complete once and hold at final state (beam-text only)
- `-file` - Path to text file for text-based effects

## Asset Directories

When using text effects with `-file`, syscgo searches for files in the following order:

1. **Absolute path** - If you provide a full path (e.g., `/home/user/art.txt`)
2. **Current directory** - Files in your working directory
3. **User assets** - `~/.local/share/syscgo/assets/` (where new art is saved)
4. **System assets** - `/usr/local/share/syscgo/assets/` or `/usr/share/syscgo/assets/`

**Where your ASCII art is saved:**

When you create ASCII art using the TUI (syscgo-tui) and export it, files are saved to:

```
~/.local/share/syscgo/assets/
```

Set `XDG_DATA_HOME` to save under a different data root. If you already keep a
`~/sysc-Go/assets/` directory from a source checkout, exports keep going there.

**Example usage:**
```bash
# Create art in TUI
syscgo-tui
# Use BIT editor, export as "hyprland.txt"

# Use your exported art
syscgo -effect fire-text -file hyprland.txt -theme rama -duration 0
```

For sysc-walls integration, art exports to `~/.config/sysc-walls/ascii/` instead.

## Effect Showcase

The AK-47 Text preview leads the gallery. The remaining effects appear newest to oldest.

<details>
<summary>Show the effect gallery</summary>

#### Great Wave
<img src="assets/great-wave.webp" alt="Great Wave Effect" loading="lazy">

#### Wasteland
<img src="assets/wasteland.webp" alt="Wasteland Effect" loading="lazy">

#### Logo Morph
<img src="assets/logo-morph.webp" alt="Logo Morph Effect" loading="lazy">

#### Cross Logo
<img src="assets/cross-logo.webp" alt="Cross Logo Effect" loading="lazy">

#### SYSC Logo
<img src="assets/sysc-logo.webp" alt="SYSC Logo Effect" loading="lazy">

#### Cracktro
<img src="assets/cracktro.webp" alt="Cracktro Effect" loading="lazy">

#### Sonar
<img src="assets/sonar.webp" alt="Sonar Effect" loading="lazy">

#### Skull
<img src="assets/skull.webp" alt="Skull Effect" loading="lazy">

#### Fire Text
<img src="assets/fire-text.webp" alt="Fire Text Effect" loading="lazy">

#### Blackhole
<img src="assets/blackhole.webp" alt="Blackhole Effect" loading="lazy">

#### Ring Text
<img src="assets/ring-text.webp" alt="Ring Text Effect" loading="lazy">

#### Rain Art
<img src="assets/rain-art.webp" alt="Rain Art Effect" loading="lazy">

#### Matrix Art
<img src="assets/matrix-art.webp" alt="Matrix Art Effect" loading="lazy">

#### Aquarium
<img src="assets/aquarium.webp" alt="Aquarium Effect" loading="lazy">

#### Beams
<img src="assets/beams.webp" alt="Beams Effect" loading="lazy">

#### Beam Text
<img src="assets/beam-text.webp" alt="Beam Text Effect" loading="lazy">

#### Pour
<img src="assets/pour.webp" alt="Pour Effect" loading="lazy">

#### Print
<img src="assets/print.webp" alt="Print Effect" loading="lazy">

#### Fireworks
<img src="assets/fireworks.webp" alt="Fireworks" loading="lazy">

#### Rain
<img src="assets/rain.webp" alt="ASCII Rain" loading="lazy">

#### Fire
<img src="assets/fire.webp" alt="Fire Effect" loading="lazy">

#### Matrix Rain
<img src="assets/matrix.webp" alt="Matrix Rain Effect" loading="lazy">

</details>

<details>
<summary>Show the TUI preview</summary>

<img src="assets/showcase_tui.webp" alt="TUI Showcase" loading="lazy">
</details>

## Live Preview

Run an effect in a terminal. Press `Ctrl+C` to stop it.

```bash
syscgo -effect great-wave -theme nord -duration 0
```

## Documentation

See [GUIDE.md](GUIDE.md) for detailed CLI usage.

## Acknowledgements

- [terminaltexteffects](https://github.com/ChrisBuilds/terminaltexteffects) - Inspiration for terminal visual effects
- [bit](https://github.com/superstarryeyes/bit) - ASCII art editor and font library integration

## License

Code: MIT. The bundled AK-47 model data is by Lokeig and retains
[CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/), including its
noncommercial restriction. See [model credits and conversion](animations/models/README.md).

---

<a href="https://github.com/Nomadcxx"><img src="https://raw.githubusercontent.com/Nomadcxx/Nomadcxx/main/assets/rama-mark.svg" height="22" alt="RAMA"></a> — terminal-native tooling for the linux desktop.
[More projects →](https://github.com/Nomadcxx) · [Sponsor](https://github.com/sponsors/Nomadcxx) ❤️
