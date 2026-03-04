package animations

import (
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
)

// MatrixArtEffect implements Matrix rain that crystallizes into ASCII art
type MatrixArtEffect struct {
	width   int
	height  int
	palette []string
	chars   []rune

	// Matrix streaks
	streaks []MatrixStreak
	frame   int

	// ASCII art formation
	text         string
	artPositions map[int]map[int]rune // [y][x] = character
	frozenChars  map[int]map[int]*FrozenMatrixChar
	centerX      int
	centerY      int
	artWidth     int
	artHeight    int
	rng          *rand.Rand
	freezeChance float64 // Probability a character freezes

	// Reusable render buffers — allocated once, cleared per frame
	canvas  [][]rune
	colors  [][]string
	builder strings.Builder
}

// FrozenMatrixChar represents a matrix character that has frozen to form the art
type FrozenMatrixChar struct {
	char  rune
	color string
}

// NewMatrixArtEffect creates a new matrix-art effect
func NewMatrixArtEffect(width, height int, palette []string, text string) *MatrixArtEffect {
	text = normalizeMultilineText(text)

	m := &MatrixArtEffect{
		width:   width,
		height:  height,
		palette: palette,
		chars: []rune{
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'α', 'β', 'γ', 'δ', 'ε', 'ζ', 'η', 'θ', 'ι', 'κ', 'λ', 'μ',
			'ν', 'ξ', 'ο', 'π', 'ρ', 'σ', 'τ', 'υ', 'φ', 'χ', 'ψ', 'ω',
			'А', 'Б', 'В', 'Г', 'Д', 'Е', 'Ж', 'З', 'И', 'Й', 'К', 'Л', 'М',
			'Н', 'О', 'П', 'Р', 'С', 'Т', 'У', 'Ф', 'Х', 'Ц', 'Ч', 'Ш', 'Щ',
			'░', '▒', '▓', '█', '▀', '▄', '▌', '▐', '■', '□', '▪', '▫',
		},
		streaks:      make([]MatrixStreak, 0, 100),
		frame:        0,
		text:         text,
		artPositions: make(map[int]map[int]rune),
		frozenChars:  make(map[int]map[int]*FrozenMatrixChar),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
		freezeChance: 0.99, // 99% chance to freeze when passing through art position (extremely fast crystallization)
	}

	m.parseArt()
	m.init()
	return m
}

// parseArt extracts ASCII art character positions
func (m *MatrixArtEffect) parseArt() {
	lines := strings.Split(m.text, "\n")
	m.artHeight = len(lines)

	// Find max line width
	m.artWidth = 0
	for _, line := range lines {
		if len([]rune(line)) > m.artWidth {
			m.artWidth = len([]rune(line))
		}
	}

	// Center the art
	m.centerX = (m.width - m.artWidth) / 2
	m.centerY = (m.height - m.artHeight) / 2

	// Parse character positions
	for lineIdx, line := range lines {
		lineRunes := []rune(line)
		for charIdx, char := range lineRunes {
			if char != ' ' && char != '\n' {
				x := m.centerX + charIdx
				y := m.centerY + lineIdx

				// Only store if within bounds
				if x >= 0 && x < m.width && y >= 0 && y < m.height {
					if m.artPositions[y] == nil {
						m.artPositions[y] = make(map[int]rune)
					}
					m.artPositions[y][x] = char
				}
			}
		}
	}
}

// init initializes matrix streaks
func (m *MatrixArtEffect) init() {
	m.initBuffers()

	// Create fixed pool of recycling streams (fewer needed with recycling + high freeze rate)
	for i := 0; i < m.width*2; i++ {
		streak := MatrixStreak{
			X:       m.rng.Intn(m.width),
			Y:       -m.rng.Intn(m.height),
			Length:  m.rng.Intn(15) + 5,
			Speed:   m.rng.Intn(3) + 1,
			Counter: 0,
			Active:  true,
		}
		m.streaks = append(m.streaks, streak)
	}
}

func (m *MatrixArtEffect) initBuffers() {
	m.canvas = make([][]rune, m.height)
	m.colors = make([][]string, m.height)
	for i := range m.canvas {
		m.canvas[i] = make([]rune, m.width)
		m.colors[i] = make([]string, m.width)
	}
	m.builder.Grow(m.width * m.height * 4)
}

func (m *MatrixArtEffect) clearBuffers() {
	for i := range m.canvas {
		for j := range m.canvas[i] {
			m.canvas[i][j] = ' '
			m.colors[i][j] = ""
		}
	}
}

// getRandomColor returns a random color from palette
func (m *MatrixArtEffect) getRandomColor() string {
	if len(m.palette) == 0 {
		return "#00ff00"
	}
	return m.palette[m.rng.Intn(len(m.palette))]
}

// getHeadColor returns the bright color for streak heads
func (m *MatrixArtEffect) getHeadColor() string {
	if len(m.palette) == 0 {
		return "#ffffff"
	}
	if len(m.palette) > 0 {
		return m.palette[len(m.palette)-1]
	}
	return m.palette[0]
}

// getTrailColor returns dimmer colors for trail positions
func (m *MatrixArtEffect) getTrailColor(position, length int) string {
	if len(m.palette) == 0 {
		return "#00aa00"
	}

	fadeFactor := float64(position) / float64(length)

	if fadeFactor < 0.2 {
		if len(m.palette) > 0 {
			return m.palette[len(m.palette)-1]
		}
		return m.palette[0]
	} else if fadeFactor < 0.5 {
		if len(m.palette) > 2 {
			return m.palette[len(m.palette)-2]
		}
		return m.palette[0]
	} else {
		return m.palette[0]
	}
}

// Update advances the simulation by one frame
func (m *MatrixArtEffect) Update() {
	m.frame++

	// Update existing streaks
	activeStreaks := m.streaks[:0]
	for _, streak := range m.streaks {
		if !streak.Active {
			continue
		}

		streak.Counter++
		if streak.Counter >= streak.Speed {
			streak.Counter = 0
			streak.Y++

			// Check if head position should freeze
			if streak.Y >= 0 && streak.Y < m.height {
				if _, yExists := m.artPositions[streak.Y]; yExists {
					if artChar, xExists := m.artPositions[streak.Y][streak.X]; xExists {
						// This position is part of the art
						if m.frozenChars[streak.Y] == nil || m.frozenChars[streak.Y][streak.X] == nil {
							// Position not yet frozen, maybe freeze it
							if m.rng.Float64() < m.freezeChance {
								// Freeze this character
								if m.frozenChars[streak.Y] == nil {
									m.frozenChars[streak.Y] = make(map[int]*FrozenMatrixChar)
								}
								m.frozenChars[streak.Y][streak.X] = &FrozenMatrixChar{
									char:  artChar,
									color: m.getHeadColor(), // Use bright color for frozen chars
								}
							}
						}
					}
				}
			}
		}

		// Recycle streak when it goes off screen (like rain-art)
		if streak.Y-streak.Length > m.height {
			// Reset streak to top with new random properties
			streak.Y = -m.rng.Intn(m.height)
			streak.X = m.rng.Intn(m.width)
			streak.Length = m.rng.Intn(15) + 5
			streak.Speed = m.rng.Intn(3) + 1
			streak.Counter = 0
		}

		activeStreaks = append(activeStreaks, streak)
	}
	m.streaks = activeStreaks
}

// Render converts the matrix and frozen art to colored output
func (m *MatrixArtEffect) Render() string {
	m.clearBuffers()

	// Render matrix streaks
	for _, streak := range m.streaks {
		if !streak.Active {
			continue
		}

		// Draw streak from head backwards
		for i := 0; i < streak.Length; i++ {
			y := streak.Y - i
			if y >= 0 && y < m.height && streak.X >= 0 && streak.X < m.width {
				char := m.chars[m.rng.Intn(len(m.chars))]
				var color string

				if i == 0 {
					// Head - brightest
					color = m.getHeadColor()
				} else {
					// Trail - dimmer based on position
					color = m.getTrailColor(i, streak.Length)
				}

				m.canvas[y][streak.X] = char
				m.colors[y][streak.X] = color
			}
		}
	}

	// Place frozen characters on top (they override matrix)
	for y, row := range m.frozenChars {
		for x, frozen := range row {
			if y >= 0 && y < m.height && x >= 0 && x < m.width {
				m.canvas[y][x] = frozen.char
				m.colors[y][x] = frozen.color
			}
		}
	}

	// Convert to colored string
	var lines []string
	for y := 0; y < m.height; y++ {
		m.builder.Reset()
		for x := 0; x < m.width; x++ {
			char := m.canvas[y][x]
			if char != ' ' && m.colors[y][x] != "" {
				styled := lipgloss.NewStyle().
					Foreground(lipgloss.Color(m.colors[y][x])).
					Render(string(char))
				m.builder.WriteString(styled)
			} else {
				m.builder.WriteRune(char)
			}
		}
		lines = append(lines, m.builder.String())
	}

	return strings.Join(lines, "\n")
}

// Reset clears frozen characters to restart the formation
func (m *MatrixArtEffect) Reset() {
	m.frozenChars = make(map[int]map[int]*FrozenMatrixChar)
}

// SetText updates the displayed text and restarts the crystallization.
func (m *MatrixArtEffect) SetText(text string) {
	m.text = normalizeMultilineText(text)
	m.artPositions = make(map[int]map[int]rune)
	m.frozenChars = make(map[int]map[int]*FrozenMatrixChar)
	m.parseArt()
}
