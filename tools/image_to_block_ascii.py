#!/usr/bin/env python3
"""
Convert images to block ASCII art using Unicode block characters.
Supports: ▀ ▄ █ ▌ ▐ ▖ ▗ ▘ ▙ ▚ ▛ ▜ ▝ ▞ ▟
"""

from PIL import Image
import sys

# Unicode block characters for 2x2 subdivision
BLOCK_CHARS = {
    (0, 0, 0, 0): ' ',
    (0, 0, 0, 1): '▗',
    (0, 0, 1, 0): '▖',
    (0, 0, 1, 1): '▄',
    (0, 1, 0, 0): '▝',
    (0, 1, 0, 1): '▐',
    (0, 1, 1, 0): '▞',
    (0, 1, 1, 1): '▟',
    (1, 0, 0, 0): '▘',
    (1, 0, 0, 1): '▚',
    (1, 0, 1, 0): '▌',
    (1, 0, 1, 1): '▙',
    (1, 1, 0, 0): '▀',
    (1, 1, 0, 1): '▜',
    (1, 1, 1, 0): '▛',
    (1, 1, 1, 1): '█',
}

def image_to_block_ascii(image_path, width=60, threshold=128):
    """Convert image to block ASCII using 2x2 pixel blocks."""
    img = Image.open(image_path).convert('L')  # Grayscale

    # Calculate height maintaining aspect ratio
    aspect = img.height / img.width
    height = int(width * aspect * 0.5)  # 0.5 because each char is 2 pixels tall

    img = img.resize((width * 2, height * 2), Image.Resampling.LANCZOS)

    result = []
    for y in range(0, height * 2, 2):
        line = ''
        for x in range(0, width * 2, 2):
            # Get 2x2 pixel block
            p1 = img.getpixel((x, y)) > threshold  # Top-left
            p2 = img.getpixel((x + 1, y)) > threshold  # Top-right
            p3 = img.getpixel((x, y + 1)) > threshold  # Bottom-left
            p4 = img.getpixel((x + 1, y + 1)) > threshold  # Bottom-right

            # Map to block character
            char = BLOCK_CHARS.get((p1, p2, p3, p4), '█')
            line += char
        result.append(line)

    return '\n'.join(result)

def image_to_simple_ascii(image_path, width=60):
    """Simple ASCII using different characters for density."""
    chars = ' .:-=+*#%@'
    img = Image.open(image_path).convert('L')

    aspect = img.height / img.width
    height = int(width * aspect * 0.5)

    img = img.resize((width, height), Image.Resampling.LANCZOS)

    result = []
    for y in range(height):
        line = ''
        for x in range(width):
            pixel = img.getpixel((x, y))
            char_idx = int(pixel / 255 * (len(chars) - 1))
            line += chars[char_idx]
        result.append(line)

    return '\n'.join(result)

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage: python image_to_block_ascii.py <image_path> [width] [mode]")
        print("  mode: 'block' (default) or 'simple'")
        sys.exit(1)

    image_path = sys.argv[1]
    width = int(sys.argv[2]) if len(sys.argv) > 2 else 60
    mode = sys.argv[3] if len(sys.argv) > 3 else 'block'

    try:
        if mode == 'simple':
            print(image_to_simple_ascii(image_path, width))
        else:
            print(image_to_block_ascii(image_path, width))
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)
