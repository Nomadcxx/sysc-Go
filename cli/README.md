# FLOYD CLI - REPL Interface

## Goal
Simple terminal chat interface like `claude` - not a TUI, just stdin/stdout streaming REPL.

## Design

```
┌─────────────────────────────────────────────────────────────┐
│ $ floyd                                                       │
│ FLOYD v1.0 | GLM-4.7 | Type 'exit' to quit                   │
│ ─────────────────────────────────────────────────────────── │
│ You: hello                                                    │
│ FLOYD: Hello! I'm FLOYD, your coding assistant.              │
│                                                               │
│ You: write a go http server                                   │
│ FLOYD: Here's a simple HTTP server in Go:                    │
│        [streams response line by line]                        │
│                                                               │
│ You: exit                                                     │
│ $                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Architecture

```
cli/
├── main.go           # Entry point, REPL loop
├── repl.go           # REPL logic (read, eval, print)
├── formatter.go      # Output formatting (colors, markdown)
└── config.go         # Config handling
```

## Features

- [ ] Simple REPL loop (read input → send to API → stream response)
- [ ] Streaming output (tokens appear as they arrive)
- [ ] Basic formatting (code blocks, bold, colors)
- [ ] History (up/down arrows)
- [ ] Multi-line input (ctrl+d to submit)
- [ ] Commands: /exit, /clear, /help, /mode

## Commands

| Command | Action |
|---------|--------|
| `exit` or `ctrl+d` | Quit |
| `ctrl+c` | Cancel current stream |
| `clear` | Clear screen |
| `help` | Show commands |
| `mode <name>` | Switch model (planner, coder, tester) |
