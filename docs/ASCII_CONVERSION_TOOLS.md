# Image to ASCII Conversion Tools

## Quick Reference for Skull Art Generation

### 1. ascii-image-converter (Recommended)

Already installed on your system.

```bash
# Basic ASCII (characters like .:-=+*#%@)
ascii-image-converter skull.png -d 50,30

# Braille mode (⠀⠁⠂⠄⠈⠐⠠⡀⢀⣀) - higher resolution
ascii-image-converter skull.png -b -d 40,25

# Braille with dithering (smoother gradients)
ascii-image-converter skull.png -b --dither -d 40,25

# Save to file
ascii-image-converter skull.png -d 50,30 > assets/skull_new.txt
```

### 2. img2txt (libcaca)

Already installed, uses color ASCII.

```bash
# Basic usage
img2txt skull.png -W 50 -H 30

# Dithering for smoother output
img2txt skull.png -W 50 -H 30 -d

# Save to file
img2txt skull.png -W 50 -H 30 > assets/skull_new.txt
```

### 3. Custom Python Script (Block ASCII)

Created: `tools/image_to_block_ascii.py`

```bash
# Block ASCII using Unicode characters (▀▄█▌▐ etc.)
python3 tools/image_to_block_ascii.py skull.png 50 block

# Simple character density ASCII
python3 tools/image_to_block_ascii.py skull.png 50 simple

# Save to file
python3 tools/image_to_block_ascii.py skull.png 50 block > assets/skull_new.txt
```

## Character Sets Explained

### 1. Simple ASCII (Density-based)
```
Chars: ' .:-=+*#%@'
```
Uses different characters based on pixel brightness.

### 2. Block Characters (2x2 subdivision)
```
▀ ▄ █ ▌ ▐ ▖ ▗ ▘ ▙ ▚ ▛ ▜ ▝ ▞ ▟
```
Divides each character into 4 quadrants for higher resolution.

### 3. Braille Patterns
```
⠀⠁⠂⠃⠄⠅⠆⠇⠈⠉⠊⠋⠌⠍⠎⠏
⠐⠑⠒⠓⠔⠕⠖⠗⠘⠙⠚⠛⠜⠝⠞⠟
⠠⠡⠢⠣⠤⠥⠦⠧⠨⠩⠪⠫⠬⠭⠮⠯
... (256 combinations)
```
Uses Braille dot patterns for very high resolution.

## Recommended Workflow for Skull

1. **Find a reference skull image** (Punisher style, simple silhouette)
2. **Convert using Braille** (highest detail):
   ```bash
   ascii-image-converter punisher_skull.png -b --dither -d 50,30 > assets/skull_braille.txt
   ```
3. **Or use block ASCII** (cleaner terminal output):
   ```bash
   python3 tools/image_to_block_ascii.py punisher_skull.png 50 block > assets/skull_block.txt
   ```
4. **Test in effect**:
   ```bash
   go build -o syscgo ./cmd/syscgo/
   ./syscgo -effect skull -theme rama -duration 10
   ```

## Online Tools (No Install)

- https://ascii-generator.site/ - Web-based converter
- https://www.text-image.com/convert/ - Many output formats
- https://www.asciiart.eu/image-to-ascii - Simple and fast

## Installation (if needed)

```bash
# ascii-image-converter
sudo pacman -S ascii-image-converter  # Arch
# or download from: https://github.com/TheZoraiz/ascii-image-converter

# img2txt (libcaca)
sudo pacman -S libcaca  # Arch
apt-get install caca-utils  # Debian/Ubuntu

# Python Pillow
pip install Pillow
```

## Tips for Best Results

1. **Use high contrast images** - Black/white silhouettes work best
2. **Simple shapes work better** - Too much detail gets lost
3. **Test width** - 40-60 chars wide fits most terminals
4. **Try Braille for curves** - Round shapes look better in Braille
5. **Try Block for sharp edges** - Angles look better with block chars
