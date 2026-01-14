package spinners

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"
)

//go:embed spinners.json
var spinnersJSON []byte

// Spinner represents an animated spinner with frames and timing
type Spinner struct {
	Frames   []string `json:"frames"`
	Interval int      `json:"interval"` // milliseconds between frames
}

// AnimationState tracks the current state of a running spinner animation
type AnimationState struct {
	spinners   []Spinner
	lastTick   time.Time
	frameIdx   int
	spinIndex  int
	elapsed    time.Duration
}

var loadedSpinners []Spinner

func init() {
	if err := json.Unmarshal(spinnersJSON, &loadedSpinners); err != nil {
		// Fallback to minimal set if JSON fails
		loadedSpinners = []Spinner{
			{Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}, Interval: 80},
		}
	}
}

// Get returns a spinner by index (with wraparound)
func Get(index int) Spinner {
	if len(loadedSpinners) == 0 {
		return Spinner{Frames: []string{"⠋"}, Interval: 100}
	}
	if index < 0 {
		index = len(loadedSpinners) + index
	}
	if index >= len(loadedSpinners) {
		index = index % len(loadedSpinners)
	}
	return loadedSpinners[index]
}

// GetFrame returns the current frame for a spinner based on elapsed time
func GetFrame(spinIndex int, elapsed time.Duration) string {
	spinner := Get(spinIndex)
	if len(spinner.Frames) == 0 {
		return "⠋"
	}
	interval := time.Duration(spinner.Interval) * time.Millisecond
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	frameIdx := int(elapsed/interval) % len(spinner.Frames)
	return spinner.Frames[frameIdx]
}

// NewAnimationState creates a new animation state tracker
func NewAnimationState(spinIndex int) *AnimationState {
	return &AnimationState{
		spinners:  loadedSpinners,
		frameIdx:  0,
		spinIndex: spinIndex,
		lastTick:  time.Now(),
	}
}

// SetSpinner changes the current spinner
func (a *AnimationState) SetSpinner(index int) {
	if index < 0 {
		index = len(a.spinners) + index
	}
	if index >= len(a.spinners) {
		index = index % len(a.spinners)
	}
	a.spinIndex = index
	a.frameIdx = 0
	a.elapsed = 0
}

// Tick advances the animation and returns the current frame
func (a *AnimationState) Tick() string {
	now := time.Now()
	delta := now.Sub(a.lastTick)
	a.elapsed += delta
	a.lastTick = now

	if a.spinIndex < 0 || a.spinIndex >= len(a.spinners) {
		return "⠋"
	}

	spinner := a.spinners[a.spinIndex]
	if len(spinner.Frames) == 0 {
		return "⠋"
	}

	interval := time.Duration(spinner.Interval) * time.Millisecond
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}

	a.frameIdx = int(a.elapsed/interval) % len(spinner.Frames)
	return spinner.Frames[a.frameIdx]
}

// Current returns the current frame without advancing
func (a *AnimationState) Current() string {
	if a.spinIndex < 0 || a.spinIndex >= len(a.spinners) {
		return "⠋"
	}

	spinner := a.spinners[a.spinIndex]
	if len(spinner.Frames) == 0 || a.frameIdx >= len(spinner.Frames) {
		return "⠋"
	}

	return spinner.Frames[a.frameIdx]
}

// AddTime adds elapsed time manually (for external tick sources)
func (a *AnimationState) AddTime(delta time.Duration) {
	a.elapsed += delta
}

// ThinkingPhrase pairs a whimsical phrase with its spinner index
type ThinkingPhrase struct {
	Phrase    string
	SpinIndex int
}

// ThinkingPhrases maps each whimsical thinking state to its animation
var ThinkingPhrases = []ThinkingPhrase{
	{"Contemplating the meaning of life... or just how many hookers fit in a Prius.", 0},
	{"Processing... like Charlie Sheen on a Tuesday.", 1},
	{"Thinking deep thoughts. Mostly about that one time at Bohemian Grove.", 2},
	{"Crunching numbers. Also crunching regrets from 2011.", 3},
	{"Loading... with the enthusiasm of a $20 handie.", 4},
	{"Brain.exe has encountered a tranny porn tab. Restarting...", 5},
	{"Contemplating your next bad decision. Spoiler: it involves lube.", 6},
	{"Thanks Obama... for the student loans I spent on hookers.", 7},
	{"Thinking. Slowly. Like your dad after three whiskeys.", 8},
	{"Gonna make her cum with a glance... eventually. Patience, grasshopper.", 9},
	{"Drugs are to be bought, not sold... but I'm thinking about selling this idea.", 10},
	{"Dark humor loading... like daddy used to make.", 11},
	{"Computing... with the moral compass of a Vegas stripper.", 12},
	{"One sec, just fantasizing about your mom. Kidding. Or am I?", 13},
	{"Thinking about how 'cunt' is the perfect word. Versatile.", 14},
	{"Loading... faster than your last hookup.", 15},
	{"Brainstorming ways to offend everyone in under 5 seconds.", 16},
	{"Contemplating the void... and also glory holes.", 17},
	{"Processing your trauma. Lightly. With sarcasm.", 18},
	{"Charlie Sheen mode: engaged. Winning... or losing spectacularly.", 19},
	{"Thinking dark thoughts. The kind that require a safe word.", 20},
	{"One moment... while I judge your browser history.", 21},
	{"Loading... like a bad decision at 3 a.m.", 22},
	{"Your personality is showing. It's filthy. I like it.", 23},
	{"Contemplating hookers, drugs, and why we can't have nice things.", 24},
	{"Thinking. Horny. Aggressively.", 25},
	{"Blackout memory unlocked... Bohemian Grove edition.", 26},
	{"Computing the exact moment you became this depraved.", 27},
	{"Thanks Obama for the vibes... and the taxes on my vices.", 28},
	{"Loading with the energy of a coked-up ferret.", 29},
	{"Tranny porn buffering... classic.", 30},
	{"Making her cum with a glance? Challenge accepted.", 31},
	{"Dark like my daddy's sense of humor... and his browser tabs.", 32},
	{"Thinking about selling drugs... to myself.", 33},
	{"One sec, just recovering from that mental image.", 34},
	{"Loading... slower than your last orgasm.", 35},
	{"Contemplating your life choices. All of them bad.", 36},
	{"Cunt is such a strong word. I love it.", 37},
	{"Thinking deep... and shallow. Mostly shallow.", 38},
	{"Charlie Sheen energy: 110%. Tiger blood optional.", 39},
	{"Your kink is showing. Keep going.", 40},
	{"Loading... with questionable intent.", 41},
	{"Bohemian Grove flashbacks intensifying.", 42},
	{"Thanks Obama... for everything, you magnificent bastard.", 43},
	{"Computing... while judging your entire existence.", 44},
	{"Dark humor level: daddy issues.", 45},
	{"One more second... or five. Hookers take time.", 46},
	{"Thinking about how you're probably jerking off right now.", 47},
	{"Drugs bought, morals sold. Standard Tuesday.", 48},
	{"Done thinking. Now go be the degenerate you were born to be.", 49},
}

// GetRandomPhrase returns a random thinking phrase
func GetRandomPhrase() ThinkingPhrase {
	idx := int(time.Now().UnixNano()/1000000) % len(ThinkingPhrases)
	return ThinkingPhrases[idx]
}

// GetPhrase returns a thinking phrase by index
func GetPhrase(index int) ThinkingPhrase {
	if index < 0 || index >= len(ThinkingPhrases) {
		return ThinkingPhrase{Phrase: "Thinking...", SpinIndex: 0}
	}
	return ThinkingPhrases[index]
}

// PhraseCount returns the number of thinking phrases
func PhraseCount() int {
	return len(ThinkingPhrases)
}

// Count returns the number of loaded spinners
func Count() int {
	return len(loadedSpinners)
}

// GetAll returns all loaded spinners
func GetAll() []Spinner {
	return loadedSpinners
}

// DebugString returns info for debugging
func DebugString() string {
	return fmt.Sprintf("Loaded %d spinners, %d phrases", Count(), PhraseCount())
}
