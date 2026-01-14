# FLOYD CLI UI/UX Roadmap

**Created:** 2026-01-12
**Status:** Planning

---

## Overview

This document outlines 14 high-value UI/UX improvements for the FLOYD CLI, split into:
- **7 Functionality Features** - Core user experience improvements
- **7 Design Elements** - Visual polish and aesthetics

All features follow the principle of **progressive disclosure** - available when needed but not cluttering the interface.

---

## Phase 1: Functionality Features

### 1. ✅ Command History Navigation
**Status:** Partially Implemented
**Priority:** P0 (Critical)

Navigate previous commands with Up/Down arrows. Subtle visual indicator when navigating history.

**Implementation:**
- [x] Basic history stored in Model.History
- [ ] Up/Down arrow navigation in input field
- [ ] Visual indicator (e.g., "↑ History [3/15]")
- [ ] Persist history to session file

**Files:** `ui/floydui/update.go`, `ui/floydui/session.go`

---

### 2. ✅ Contextual Help System
**Status:** Partially Implemented
**Priority:** P0 (Critical)

Display relevant help information in a dedicated panel.

**Implementation:**
- [x] `/help` command toggles help
- [ ] Context-sensitive help based on current state
- [ ] Slide-in panel from right (overlay)
- [ ] Keyboard shortcut hints in help

**Files:** `ui/floydui/view.go`, `ui/floydui/help.go`

---

### 3. ☐ Keyboard Shortcuts Display
**Status:** Not Started
**Priority:** P1 (High)

Toggleable shortcuts panel showing key commands.

**Implementation:**
- [ ] Create shortcuts data structure
- [ ] `/shortcuts` or `?` to toggle panel
- [ ] Context-sensitive shortcuts in status bar
- [ ] Group by category (navigation, editing, commands)

**Shortcuts to Display:**
```
Navigation:       Editing:           Commands:
↑/↓   History    Ctrl+A  Select All  /help     Help
PgUp  Scroll Up  Ctrl+C  Copy        /clear    Clear
PgDn  Scroll Dn  Ctrl+V  Paste       /compact  Compact
Home  Start      Ctrl+U  Clear Line  /exit     Exit
End   End        
```

---

### 4. ☐ Progress Indicators
**Status:** Partially Implemented
**Priority:** P1 (High)

Visual feedback for long-running operations.

**Implementation:**
- [x] Basic thinking animation (phrases + dots)
- [ ] Progress bar in status bar for agent calls
- [ ] Step indicator for multi-step processes ("Step 2/4")
- [ ] Tool execution progress
- [ ] Elapsed time display

**Design:**
```
[████████░░░░░░░░░░░░] 42% │ Executing bash... │ 12s
```

---

### 5. ✅ Multi-line Input Support
**Status:** Partially Implemented
**Priority:** P1 (High)

Compose multi-line messages naturally.

**Implementation:**
- [x] Input field exists
- [ ] Shift+Enter creates new line
- [ ] Enter on empty new line submits
- [ ] Dynamic height expansion of input area
- [ ] Line count indicator

**Behavior:**
- `Enter` on empty line = Submit
- `Shift+Enter` = New line
- Input area grows up to 10 lines max

---

### 6. ✅ Configurable Themes
**Status:** Implemented
**Priority:** P2 (Medium)

Theme presets maintaining glassmorphic design.

**Implementation:**
- [x] Theme struct with all colors
- [x] 5 themes defined (classic, dark, highcontrast, darkside, midnight)
- [x] `/theme <name>` command
- [x] Theme persists in session
- [ ] Custom theme loading from ~/.floyd/themes/

---

### 7. ✅ Session Persistence
**Status:** Implemented
**Priority:** P0 (Critical)

Save and restore session state.

**Implementation:**
- [x] Auto-save on exit
- [x] Load previous session on startup
- [x] Save messages and history
- [ ] Session restore indicator in status bar
- [ ] Multiple session slots

---

## Phase 2: Design Elements

### 8. ☐ Subtle Animated Transitions
**Status:** Not Started
**Priority:** P2 (Medium)

Smooth transitions between states for premium feel.

**Implementation:**
- [ ] Fade-in/fade-out for panels
- [ ] Smooth scrolling in viewport
- [ ] Easing functions for animations
- [ ] Error panel slide-in animation

**Technical:** Use Bubble Tea's tick mechanism with interpolation

---

### 9. ☐ Glassmorphism Depth Layers
**Status:** Partially Implemented
**Priority:** P2 (Medium)

Visual depth through layered effects.

**Implementation:**
- [x] Different background colors for panels
- [ ] Varying opacity levels (header 90%, viewport 85%, status 95%)
- [ ] Subtle glow around active elements
- [ ] Border gradients

---

### 10. ✅ Themed Loading Animations
**Status:** Partially Implemented
**Priority:** P3 (Low)

Custom loading animations matching theme.

**Implementation:**
- [x] Thinking phrases rotation
- [x] Dot animation (. .. ...)
- [ ] Pink pulsing prism effect
- [ ] Spinning FLOYD logo ASCII art
- [ ] Different animations per operation type

---

### 11. ☐ Typography Hierarchy
**Status:** Not Started  
**Priority:** P2 (Medium)

Clear typographic scale for visual order.

**Implementation:**
- [ ] Header text styling (bold, larger effect via spacing)
- [ ] Message differentiation (user vs assistant)
- [ ] Code block styling (monospace, highlighted)
- [ ] Status text styling (muted, smaller)
- [ ] Timestamp styling

---

### 12. ☐ Micro-interactions
**Status:** Not Started
**Priority:** P3 (Low)

Small interactive details that feel alive.

**Implementation:**
- [ ] Cursor pulse in input field
- [ ] "Message sent" subtle flash
- [ ] Tool execution indicator pulse
- [ ] Scroll position feedback

---

### 13. ☐ Themed Sound Effects (Optional)
**Status:** Not Started
**Priority:** P4 (Optional)

Audio feedback matching theme.

**Implementation:**
- [ ] Optional - disabled by default
- [ ] Notification chime
- [ ] Message sent sound
- [ ] Error sound
- [ ] Volume control in settings

**Note:** This is terminal-dependent and may not work everywhere.

---

### 14. ✅ Color Psychology Implementation
**Status:** Implemented
**Priority:** P2 (Medium)

Strategic use of color for emotion and clarity.

**Implementation:**
- [x] Pink for active/accent elements
- [x] Softer tones for backgrounds
- [x] Error red for errors
- [x] Green for success
- [x] Yellow for warnings
- [ ] Time-of-day color shift (optional)

---

## Implementation Priority Order

### Sprint 1: Foundation (Current)
1. ✅ Configurable Themes
2. ✅ Session Persistence  
3. ✅ Contextual Help System (basic)
4. ✅ Color Psychology

### Sprint 2: Core UX
5. Command History Navigation (complete)
6. Multi-line Input Support
7. Progress Indicators (enhanced)
8. Keyboard Shortcuts Display

### Sprint 3: Visual Polish
9. Typography Hierarchy
10. Glassmorphism Depth Layers
11. Subtle Animated Transitions
12. Themed Loading Animations

### Sprint 4: Delight
13. Micro-interactions
14. Themed Sound Effects (optional)

---

## Technical Notes

### Animation Framework
Use Bubble Tea's tick system with state interpolation:
```go
type animationState struct {
    progress float64
    target   float64
    easing   func(float64) float64
}
```

### Multi-line Input
Modify textinput handling to intercept Shift+Enter:
```go
case tea.KeyShiftEnter:
    m.Input.InsertRune('\n')
    return m, nil
```

### Progress Bar
Integrate with tool loop events:
```go
case loop.EventTypeProgress:
    m.ProgressPercent = event.Progress
```

---

## References

- [awesome-cli-apps](https://github.com/agarrharr/awesome-cli-apps)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)

---

*"The details are not the details. They make the design." - Charles Eames*
