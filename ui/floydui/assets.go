package floydui

import (
	"encoding/json"
	"fmt"
	"os"
)

// PRISMBanner - the official FLOYD banner
const PRISMBanner = `
    __/\\\\\\\\\\\\\\\__/\\\___________________/\\\\\_______/\\\________/\\\__/\\\\\\\\\\\\_____________________/\\\\\\\\\_______/\\\\\_______/\\\\\\\\\\\\_____/\\\\\\\\\\\\\\\__________________/\\\\\\\\\__/\\\______________/\\\\\\\\\\\_
     _\/\\\///////////__\/\\\_________________/\\\///\\\____\///\\\____/\\\/__\/\\\////////\\\________________/\\\////////______/\\\///\\\____\/\\\////////\\\__\/\\\///////////________________/\\\////////__\/\\\_____________\/////\\\///__
      _\/\\\_____________\/\\\_______________/\\\/__\///\\\____\///\\\/\\\/____\/\\\______\//\\\_____________/\\\/_____________/\\\/__\///\\\__\/\\\______\//\\\_\/\\\_________________________/\\\/___________\/\\\_________________\/\\\_____
       _\/\\\\\\\\\\\_____\/\\\______________/\\\______\//\\\_____\///\\\/_______\/\\\_______\/\\\____________/\\\______________/\\\______\//\\\_\/\\\_______\/\\\_\/\\\\\\\\\\\________________/\\\_____________\/\\\_________________\/\\\_____
        _\/\\\///////______\/\\\_____________\/\\\_______\/\\\_______\/\\\_______\/\\\_______\/\\\___________\/\\\_____________\/\\\_______\/\\\_\/\\\_______\/\\\_\/\\\///////________________\/\\\_____________\/\\\_________________\/\\\_____
         _\/\\\_____________\/\\\_____________\//\\\______/\\\________\/\\\_______\/\\\_______\/\\\___________\//\\\____________\//\\\______/\\\__\/\\\_______\/\\\_\/\\\_______________________\//\\\____________\/\\\_________________\/\\\_____
          _\/\\\_____________\/\\\______________\///\\\__/\\\__________\/\\\_______\/\\\_______\/\\\_____________\///\\\___________\///\\\__/\\\____\/\\\_______\/\\\_\/\\\________________________\///\\\__________\/\\\_________________\/\\\_____
           _\/\\\_____________\/\\\\\\\\\\\\\\\____\///\\\\\/___________\/\\\_______\/\\\\\\\\\\\\/________________\////\\\\\\\\\____\///\\\\\/_____\/\\\\\\\\\\\\/___\/\\\\\\\\\\\\\\\______________\////\\\\\\\\\_\/\\\\\\\\\\\\\\\__/\\\\\\\\\\\_
            _\///______________\///////////////_______\/////_____________\///________\////////////_____________________\/////////_______\/////_______\////////////_____\///////////////__________________\/////////__\///////////////__\///////////__
`

// MediumBanner - The structural "PRISM" banner, compressed for standard terminals (~144 cols)
const MediumBanner = `
 __/\\\\\\\\\\\\\\\__/\\\_/\\\\\_/\\\__/\\\__/\\\\\\\\\\\\___/\\\\\\\\\_/\\\\\_/\\\\\\\\\\\\__/\\\\\\\\\\\\\\\___/\\\\\\\\\__/\\\__/\\\\\\\\\\\_
  _\/\\\///////////__\/\\\__/\\\///\\\_\///\\\_/\\\/__\/\\\////////\\\_/\\\////////___/\\\///\\\_\/\\\////////\\\__\/\\\///////////_/\\\////////__\/\\\_\/////\\\///__
   _\/\\\_\/\\\___/\\\/__\///\\\_\///\\\/\\\/_\/\\\___\//\\\_/\\\/_/\\\/__\///\\\__\/\\\___\//\\\_\/\\\_/\\\/__\/\\\__\/\\\__
 _\/\\\\\\\\\\\__\/\\\__/\\\___\//\\\__\///\\\/___\/\\\_\/\\\___/\\\__/\\\___\//\\\_\/\\\_\/\\\_\/\\\\\\\\\\\_/\\\_\/\\\__\/\\\__
  _\/\\\///////___\/\\\_\/\\\_\/\\\_\/\\\_\/\\\_\/\\\__\/\\\_\/\\\_\/\\\_\/\\\_\/\\\_\/\\\///////_\/\\\_\/\\\__\/\\\__
   _\/\\\_\/\\\_\//\\\___/\\\__\/\\\_\/\\\_\/\\\__\//\\\___\//\\\___/\\\__\/\\\_\/\\\_\/\\\__\//\\\___\/\\\__\/\\\__
 _\/\\\_\/\\\__\///\\\__/\\\_\/\\\_\/\\\_\/\\\_\///\\\__\///\\\__/\\\_\/\\\_\/\\\_\/\\\__\///\\\_\/\\\__\/\\\__
  _\/\\\_\/\\\\\\\\\\\\\\\_\///\\\\\/__\/\\\_\/\\\\\\\\\\\\/_\////\\\\\\\\\_\///\\\\\/__\/\\\\\\\\\\\\/___\/\\\\\\\\\\\\\\\__\////\\\\\\\\\_\/\\\\\\\\\\\\\\\__/\\\\\\\\\\\_
   _\///__\///////////////_\/////_\///__\////////////___\/////////_\/////_\////////////__\///////////////___\/////////__\///////////////__\///////////__
`

// SmallBanner - A compact fallback for very narrow terminals
const SmallBanner = `
   ________    ______  __  ______ 
  / ____/ /   / __ \ \/ / / __ \ \ 
 / /_  / /   / / / /\  / / / / /  
/ __/ / /___/ /_/ / / / / /_/ /   
/_/   /_____/\____/ /_/ /_____/    
`

// Taglines - FLOYD acronym meanings
var Taglines = []string{
	"File-Logged Orchestrator Yielding Deliverables",
	"Framework for Lifecycle-Orchestrated Yield & Delivery",
	"Fully Logged Orchestration for Yielding Deployables",
	"File-Operated LLM Orchestrator for Your Development",
	"Focused Lead Orchestrator for Your Devteam",
	"Forward Logistics Orchestrator for Yielding Deployments",
	"Field Lead Orchestrator for Yielding Deliverables",
}

// ThinkingState represents a thinking phrase
type ThinkingState struct {
	Phrase      string
	Description string
}

// Spinner represents the JSON structure for animations
type Spinner struct {
	Interval int      `json:"interval"`
	Frames   []string `json:"frames"`
}

var Spinners []Spinner

// ThinkingStates - all the whimsical thinking phrases with descriptions
var ThinkingStates = []ThinkingState{
	{"Contemplating the meaning of life... or just how many hookers fit in a Prius.", "[spinning tiger blood emoji]"},
	{"Processing... like Charlie Sheen on a Tuesday.", "[tiger roar sound]"},
	{"Thinking deep thoughts. Mostly about that one time at Bohemian Grove.", "[owl hoot + blackout flicker]"},
	{"Crunching numbers. Also crunching regrets from 2011.", "[cocaine straw slurp sfx]"},
	{"Loading... with the enthusiasm of a $20 handie.", "[wet squelch]"},
	{"Brain.exe has encountered a tranny porn tab. Restarting...", "[glitchy moan]"},
	{"Contemplating your next bad decision. Spoiler: it involves lube.", "[bottle pop]"},
	{"Thanks Obama... for the student loans I spent on hookers.", "[sarcastic clap]"},
	{"Thinking. Slowly. Like your dad after three whiskeys.", "[burp]"},
	{"Gonna make her cum with a glance... eventually. Patience, grasshopper.", "[slow wink]"},
	{"Drugs are to be bought, not sold... but I'm thinking about selling this idea.", "[cash register]"},
	{"Dark humor loading... like daddy used to make.", "[creepy chuckle]"},
	{"Computing... with the moral compass of a Vegas stripper.", "[slot machine ding]"},
	{"One sec, just fantasizing about your mom. Kidding. Or am I?", "[devil horns]"},
	{"Thinking about how 'cunt' is the perfect word. Versatile.", "[word splash]"},
	{"Loading... faster than your last hookup.", "[quick zipper]"},
	{"Brainstorming ways to offend everyone in under 5 seconds.", "[record scratch]"},
	{"Contemplating the void... and also glory holes.", "[echoey moan]"},
	{"Processing your trauma. Lightly. With sarcasm.", "[therapy couch creak]"},
	{"Charlie Sheen mode: engaged. Winning... or losing spectacularly.", "[tiger blood drip]"},
	{"Thinking dark thoughts. The kind that require a safe word.", "[whip crack]"},
	{"One moment... while I judge your browser history.", "[incognito tab close]"},
	{"Loading... like a bad decision at 3 a.m.", "[stumble sfx]"},
	{"Your personality is showing. It's filthy. I like it.", "[lick]"},
	{"Contemplating hookers, drugs, and why we can't have nice things.", "[sigh]"},
	{"Thinking. Horny. Aggressively.", "[growl]"},
	{"Blackout memory unlocked... Bohemian Grove edition.", "[owl scream]"},
	{"Computing the exact moment you became this depraved.", "[clock tick]"},
	{"Thanks Obama for the vibes... and the taxes on my vices.", "[mic drop]"},
	{"Loading with the energy of a coked-up ferret.", "[ferret screech]"},
	{"Tranny porn buffering... classic.", "[pixelated moan]"},
	{"Making her cum with a glance? Challenge accepted.", "[laser eyes]"},
	{"Dark like my daddy's sense of humor... and his browser tabs.", "[shadow creep]"},
	{"Thinking about selling drugs... to myself.", "[sniff]"},
	{"One sec, just recovering from that mental image.", "[gag]"},
	{"Loading... slower than your last orgasm.", "[sad trombone]"},
	{"Contemplating your life choices. All of them bad.", "[judgy stare]"},
	{"Cunt is such a strong word. I love it.", "[bold text flash]"},
	{"Thinking deep... and shallow. Mostly shallow.", "[dive sound]"},
	{"Charlie Sheen energy: 110%. Tiger blood optional.", "[roar]"},
	{"Your kink is showing. Keep going.", "[wink]"},
	{"Loading... with questionable intent.", "[suspicious music]"},
	{"Bohemian Grove flashbacks intensifying.", "[fire crackle]"},
	{"Thanks Obama... for everything, you magnificent bastard.", "[salute]"},
	{"Computing... while judging your entire existence.", "[calculator beep]"},
	{"Dark humor level: daddy issues.", "[creepy whisper]"},
	{"One more second... or five. Hookers take time.", "[clock fast-forward]"},
	{"Thinking about how you're probably jerking off right now.", "[stroking motion]"},
	{"Drugs bought, morals sold. Standard Tuesday.", "[cash exchange]"},
	{"Done thinking. Now go be the degenerate you were born to be.", "[middle finger salute]"},
}

// HelpContent - the help text
const HelpContent = `
FLOYD CLI - HELP

KEYBOARD SHORTCUTS
------------------
  Up/Down       : Navigate command history
  Shift+Enter   : Insert a new line in the input field
  ?             : Toggle this help panel
  Ctrl+C / Esc  : Quit and save session

COMMANDS
--------
  /theme <name> : Change the theme.
                  Available themes: classic, dark, highcontrast, cyberpunk, midnight

  /clear        : Clear chat history
  /status       : Show workspace status
  /tools        : List available tools
  /protocol     : Show FLOYD protocol status
  /help, /?     : Show this help
`

// GetRandomTagline returns a random tagline
func GetRandomTagline() string {
	// Simple hash-based selection for consistency
	return Taglines[0] // Can be made random
}

// GetAnimatedMicroInteraction returns an animated micro-interaction string for the current thinking state
func GetAnimatedMicroInteraction(index int, frame int) string {
	if len(Spinners) == 0 {
		return ""
	}
	// Cycle through spinners if we have fewer spinners than phrases
	sIdx := index % len(Spinners)
	s := Spinners[sIdx]
	if len(s.Frames) == 0 {
		return ""
	}
	// Cycle through frames based on ticks (slowed down)
	frameIdx := (frame / 2) % len(s.Frames)
	return Styles.MicroInteraction.Render(s.Frames[frameIdx])
}

// LoadSpinners loads the animations from spinners/spinners.json
func LoadSpinners() error {
	data, err := os.ReadFile("spinners/spinners.json")
	if err != nil {
		// Fallback to a single basic spinner if file missing
		Spinners = append(Spinners, Spinner{
			Interval: 100,
			Frames:   []string{"|", "/", "-", "\\"},
		})
		return fmt.Errorf("load spinners: %w", err)
	}

	if err := json.Unmarshal(data, &Spinners); err != nil {
		return fmt.Errorf("parse spinners: %w", err)
	}

	return nil
}
