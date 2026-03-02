package animations

import (
	"math/rand"
	"strings"
)

// BurnPhase represents the current state of the burn animation
type BurnPhase int

const (
	PhaseIgnition BurnPhase = iota // Sequential ignition
	PhaseBurning                   // Full burning
	PhaseSettle                    // Disappearance
	PhasePause                     // Brief rest before reset
)

// BurnChar represents a character being burned
type BurnChar struct {
	char          rune
	x, y          int
	ignitionFrame int  // When fire reached this char (-1 = not yet)
	burnStage     int  // Current stage in burn sequence
	colorIndex    int  // Current color in gradient (0-8)
	stageFrame    int  // Frames on current stage
	colorFrame    int  // Frames on current color
	active        bool // Still burning vs disappeared
}

// TreeEdge represents an edge in the spanning tree
type TreeEdge struct {
	from, to int     // Character indices
	weight   float64 // For Prim's algorithm
}

// Edge for priority queue (internal use)
type Edge struct {
	from, to int
	weight   float64
}

// SmokeParticle represents rising smoke
type SmokeParticle struct {
	x, y     float64
	char     rune
	velocity float64 // Rise speed (0.5-0.8)
	opacity  float64 // 1.0 to 0.0
	lifetime int     // Frames alive
	maxLife  int     // Frames until full fade
}

// BurnAnimation implements the burn effect
type BurnAnimation struct {
	width, height int
	palette       []string // 12 colors (9 burn + 2 smoke + 1 bg)
	theme         string
	withText      bool
	textContent   string

	// Character management
	chars []BurnChar

	// Spanning tree
	spanningTree []TreeEdge
	burnOrder    []int // Ordered char indices
	nextIgnition int   // Index in burnOrder

	// Smoke
	smokeParticles []SmokeParticle
	smokeSymbols   []rune

	// Phase tracking
	phase      BurnPhase
	frameCount int
	cycleFrame int

	// Sequences
	burnSequence  []rune // For burn-text
	flameSequence []rune // For burn
}

// NewBurnEffect creates a new burn background effect
func NewBurnEffect(width, height int, palette []string, theme string) *BurnAnimation {
	return newBurnAnimation(width, height, palette, theme, false, "")
}

// NewBurnTextEffect creates a burn effect with text destruction
func NewBurnTextEffect(width, height int, palette []string, theme string, text string) *BurnAnimation {
	return newBurnAnimation(width, height, palette, theme, true, text)
}

// newBurnAnimation is the internal constructor
func newBurnAnimation(width, height int, palette []string, theme string, withText bool, text string) *BurnAnimation {
	b := &BurnAnimation{
		width:        width,
		height:       height,
		palette:      palette,
		theme:        theme,
		withText:     withText,
		textContent:  text,
		phase:        PhaseIgnition,
		frameCount:   0,
		cycleFrame:   0,
		nextIgnition: 0,
	}

	// Set sequences
	b.burnSequence = []rune{'\'', '.', '▖', '▙', '█', '▜', '▀', '▝', '.'} // 9 stages
	b.flameSequence = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}      // 8 stages
	b.smokeSymbols = b.getSmokeSymbols()

	// Initialize characters
	b.initializeCharacters()

	// Generate spanning tree
	b.generateSpanningTree()

	return b
}

// getSmokeSymbols returns theme-specific smoke characters
func (b *BurnAnimation) getSmokeSymbols() []rune {
	switch strings.ToLower(b.theme) {
	case "monochrome", "dark":
		return []rune{'▒', '░', '▓'}
	case "dracula", "catppuccin", "rama":
		return []rune{'·', '˙', '°'}
	case "nord", "gruvbox", "tokyo-night", "tokyonight":
		return []rune{'.', '\'', ',', '·'}
	case "material", "solarized":
		return []rune{'░', '·', '.'}
	default:
		return []rune{'·', '.', '░'}
	}
}

// initializeCharacters creates the character array
func (b *BurnAnimation) initializeCharacters() {
	b.chars = []BurnChar{}

	if b.withText {
		b.initializeTextChars()
	} else {
		b.initializeRandomChars()
	}
}

// initializeTextChars parses and positions user text
func (b *BurnAnimation) initializeTextChars() {
	if b.textContent == "" {
		return
	}

	lines := strings.Split(b.textContent, "\n")
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	// Center text
	textHeight := len(lines)
	offsetY := (b.height - textHeight) / 2
	offsetX := (b.width - maxWidth) / 2

	for y, line := range lines {
		for x, ch := range line {
			if ch != ' ' && ch != '\n' && ch != '\r' {
				b.chars = append(b.chars, BurnChar{
					char:          ch,
					x:             offsetX + x,
					y:             offsetY + y,
					ignitionFrame: -1, // Not yet ignited
					burnStage:     0,
					colorIndex:    0,
					stageFrame:    0,
					colorFrame:    0,
					active:        false,
				})
			}
		}
	}
}

// initializeRandomChars fills canvas with random ASCII
func (b *BurnAnimation) initializeRandomChars() {
	// Printable ASCII characters
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-=_+[]{}|;:,.<>?/"

	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			// ~30% fill density for background effect
			if rand.Float64() < 0.3 {
				ch := rune(chars[rand.Intn(len(chars))])
				b.chars = append(b.chars, BurnChar{
					char:          ch,
					x:             x,
					y:             y,
					ignitionFrame: -1,
					burnStage:     0,
					colorIndex:    0,
					stageFrame:    0,
					colorFrame:    0,
					active:        false,
				})
			}
		}
	}
}

// selectCenterBiasedOrigin picks origin with center bias (70% center, 30% anywhere)
func (b *BurnAnimation) selectCenterBiasedOrigin() int {
	if len(b.chars) == 0 {
		return 0
	}

	// Calculate center region (middle 40% of canvas)
	centerMinX := int(float64(b.width) * 0.3)
	centerMaxX := int(float64(b.width) * 0.7)
	centerMinY := int(float64(b.height) * 0.3)
	centerMaxY := int(float64(b.height) * 0.7)

	// 70% chance to pick from center region
	if rand.Float64() < 0.7 {
		// Find chars in center region
		centerChars := []int{}
		for i, char := range b.chars {
			if char.x >= centerMinX && char.x <= centerMaxX &&
				char.y >= centerMinY && char.y <= centerMaxY {
				centerChars = append(centerChars, i)
			}
		}

		if len(centerChars) > 0 {
			return centerChars[rand.Intn(len(centerChars))]
		}
	}

	// 30% chance or no center chars: pick anywhere
	return rand.Intn(len(b.chars))
}

// getNeighbors returns indices of characters within Manhattan distance 1
func (b *BurnAnimation) getNeighbors(charIndex int) []int {
	if charIndex < 0 || charIndex >= len(b.chars) {
		return []int{}
	}

	char := b.chars[charIndex]
	neighbors := []int{}

	// Check all characters for 4-directional neighbors
	for i, other := range b.chars {
		if i == charIndex {
			continue
		}

		// Manhattan distance = 1 (up, down, left, right)
		dx := abs(other.x - char.x)
		dy := abs(other.y - char.y)

		if (dx == 1 && dy == 0) || (dx == 0 && dy == 1) {
			neighbors = append(neighbors, i)
		}
	}

	return neighbors
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// extractMin finds and removes edge with minimum weight from slice
func extractMin(queue []Edge) (Edge, []Edge) {
	if len(queue) == 0 {
		return Edge{}, queue
	}

	minIdx := 0
	minWeight := queue[0].weight

	for i, edge := range queue {
		if edge.weight < minWeight {
			minWeight = edge.weight
			minIdx = i
		}
	}

	minEdge := queue[minIdx]
	// Remove from queue
	queue = append(queue[:minIdx], queue[minIdx+1:]...)

	return minEdge, queue
}

// generateSpanningTree builds spanning tree using Prim's algorithm
func (b *BurnAnimation) generateSpanningTree() {
	if len(b.chars) == 0 {
		return
	}

	// Select center-biased origin
	origin := b.selectCenterBiasedOrigin()

	// Prim's MST
	visited := make(map[int]bool)
	edges := []TreeEdge{}
	priorityQueue := []Edge{{from: origin, to: origin, weight: 0}}

	for len(priorityQueue) > 0 && len(visited) < len(b.chars) {
		var edge Edge
		edge, priorityQueue = extractMin(priorityQueue)

		if visited[edge.to] {
			continue
		}

		visited[edge.to] = true
		edges = append(edges, TreeEdge{from: edge.from, to: edge.to})

		// Add neighbors with random weights
		neighbors := b.getNeighbors(edge.to)
		for _, neighbor := range neighbors {
			if !visited[neighbor] {
				weight := rand.Float64() // Random for varied spread
				priorityQueue = append(priorityQueue, Edge{
					from:   edge.to,
					to:     neighbor,
					weight: weight,
				})
			}
		}
	}

	b.spanningTree = edges
	b.extractBurnOrder()
}

// extractBurnOrder performs BFS on spanning tree to get ignition sequence
func (b *BurnAnimation) extractBurnOrder() {
	if len(b.spanningTree) == 0 {
		return
	}

	// Build adjacency list from spanning tree
	adj := make(map[int][]int)
	for _, edge := range b.spanningTree {
		adj[edge.from] = append(adj[edge.from], edge.to)
		adj[edge.to] = append(adj[edge.to], edge.from)
	}

	// BFS from origin (first edge's 'from')
	origin := b.spanningTree[0].from
	visited := make(map[int]bool)
	queue := []int{origin}
	order := []int{}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if visited[node] {
			continue
		}

		visited[node] = true
		order = append(order, node)

		// Add neighbors
		for _, neighbor := range adj[node] {
			if !visited[neighbor] {
				queue = append(queue, neighbor)
			}
		}
	}

	b.burnOrder = order
}

// Update advances the animation by one frame
func (b *BurnAnimation) Update() {
	b.frameCount++
	b.cycleFrame++

	// TODO: Phase-based updates
	// TODO: Update smoke
}

// Render returns the current frame as a string
func (b *BurnAnimation) Render() string {
	// TODO: Render burn and smoke
	return ""
}

// Reset restarts the animation
func (b *BurnAnimation) Reset() {
	b.phase = PhaseIgnition
	b.frameCount = 0
	b.cycleFrame = 0
	b.nextIgnition = 0
	// TODO: Regenerate characters and spanning tree
}

// SetText updates the displayed text and reinitializes the burn animation.
func (b *BurnAnimation) SetText(text string) {
	b.textContent = text
	b.withText = text != ""
	b.phase = PhaseIgnition
	b.frameCount = 0
	b.cycleFrame = 0
	b.nextIgnition = 0
	b.smokeParticles = b.smokeParticles[:0]
	if b.withText {
		b.initializeTextChars()
	} else {
		b.initializeRandomChars()
	}
	b.generateSpanningTree()
	b.extractBurnOrder()
}
