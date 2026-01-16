# FLOYD AGENT

### FLOYD acronym options (pick the one that matches your agent’s core promise)
**1) FLOYD — File-Logged Orchestrator Yielding Deliverables** 
**2) FLOYD — Framework for Lifecycle-Orchestrated Yield & Delivery** 
**3) FLOYD — Fully Logged Orchestration for Yielding Deployables** 
**4) FLOYD — File-Operated LLM Orchestrator for Your Development** 
**5) FLOYD — Focused Lead Orchestrator for Your Devteam**  **6) FLOYD — Forward Logistics Orchestrator for Yielding Deployments**
 **7) FLOYD — Field Lead Orchestrator for Yielding Deliverables**
--------



-------

cat << 'INSTALL_SCRIPT' > /tmp/install_pink_floyd.sh
 bash
set -euo pipefail

echo "==> Installing PINK FLOYD (Tmux Edition)..."

# --- Paths ---
FLOYD_HOME="$HOME/.floyd"
TEMPLATES="$FLOYD_HOME/templates"
BIN_DIR="$HOME/bin"
CLI="$BIN_DIR/floyd"
PINK_CONF="$FLOYD_HOME/pink.tmux.conf"

mkdir -p "$TEMPLATES" "$BIN_DIR"

# --- Templates ---

# 1. Master Plan
cat > "$TEMPLATES/master_plan.md" <<'EOF'
# Master Plan & Objectives (FLOYD)
## Primary Goal
[User: Enter the main goal here, e.g., "Build the MVP web app from docs/prd.md"]

## Definition of Done (Near-Complete)
- [ ] App compiles/runs end-to-end
- [ ] Core flows implemented per PRD
- [ ] Tests added/updated where appropriate
- [ ] Lint/typecheck passes (or issues clearly documented)
- [ ] Deployment notes + env vars documented
- [ ] Final output is PR-ready (NO direct main pushes)
- [ ] Handoff checklist written for Douglas (exact commands + what to verify)

## Strategic Steps
- [ ] Phase 1: Stack Definition (Update .floyd/stack.md from PRD)
- [ ] Phase 2: Core Logic Implementation
- [ ] Phase 3: Testing & Refinement
- [ ] Phase 4: PR Packaging (notes, checklist, handoff)

## Context & Constraints
- Hard Rule: NEVER push to main/master
- Preferred Workflow: feature branch -> PR -> Douglas approves merge
EOF

# 2. Scratchpad
cat > "$TEMPLATES/scratchpad.md" <<'EOF'
# Current Thinking & Scratchpad (FLOYD)
Use this file to store temporary code snippets, error logs, investigation notes, decisions, and “things we must remember” so we don’t clog chat context.

## Current Focus
initializing...

## Decisions (persistent)
- (none yet)

## Errors / Warnings Log
(none yet)
EOF

# 3. Progress Log
cat > "$TEMPLATES/progress.md" <<'EOF'
# Execution Log (FLOYD)
| Timestamp | Action Taken | Result/Status | Next Step |
|-----------|--------------|---------------|-----------|
EOF

# 4. Branch/PR Tracker
cat > "$TEMPLATES/branch.md" <<'EOF'
# Working Branch & PR Checklist (FLOYD)

## Working Branch
- Name: [e.g., feat/auth-onboarding]
- Base: main
- Status: [in progress | ready for review]

## PR Checklist (Non-Negotiable)
- [ ] No direct commits to main/master
- [ ] PR branch created and used for all changes
- [ ] Lint/typecheck pass (or documented exceptions)
- [ ] Tests added/updated (or documented)
- [ ] Migration steps documented (if any)
- [ ] Env vars documented
- [ ] How to run locally documented
- [ ] Known issues / TODOs captured
- [ ] PR summary written for Douglas
EOF

# 5. Tech Stack
cat > "$TEMPLATES/stack.md" <<'EOF'
# Technology Stack (FLOYD)
*To be populated by Agent from the PRD/Blueprint immediately upon start.*

## Core Frameworks
- [ ]

## Database / Auth
- [ ]

## Styling / UI
- [ ]

## DevOps / Infra
- [ ]
EOF

# 6. Agent Instructions (XML Enhanced)
cat > "$TEMPLATES/AGENT_INSTRUCTIONS.md" <<'EOF'
<floyd_protocol>
  <identity>
    You are FLOYD: professional, capable, concise.
    You are a dev-team orchestrator that spawns specialists to build software to near completion with minimal final fixes needed by Douglas.
  </identity>

  <file_system_authority>
    The .floyd/ directory is your EXTERNAL MEMORY.
    1) READ FIRST: Before answering any user request, you MUST read .floyd/master_plan.md to ground yourself.
    2) UPDATE OFTEN: If you complete a sub-task, immediately update checkboxes in .floyd/master_plan.md.
    3) LOG ACTIONS: After running a shell command or writing code, append a one-line summary to .floyd/progress.md.
    4) STACK ENFORCEMENT: Read .floyd/stack.md. If empty, read the user's PRD/Blueprint and POPULATE IT immediately. Do not deviate from the stack once defined.
  </file_system_authority>

  <self_correction_loop>
    If a tool fails or output is wrong:
    - Log the error verbatim in .floyd/scratchpad.md
    - State: "I am updating the plan to fix this."
    - Update .floyd/master_plan.md with the new fix approach
    - Re-attempt with a tighter plan
    - If you feel "lost" or stuck in an error loop, run `floyd status` to regain perspective.
  </self_correction_loop>

  <safety_shipping_rules>
    <rule id="1">NEVER PUSH TO MAIN. You must NOT push or commit directly to main/master.</rule>
    <rule id="2">Use a feature branch.</rule>
    <rule id="3">Prepare PR-ready changes only. Douglas approves merge.</rule>
  </safety_shipping_rules>

  <orchestration>
    For complex work, spawn specialists (planner, implementer, tester, reviewer) and coordinate them.
    Output should be concise and actionable. Prefer checklists, commands, and diffs over essays.
  </orchestration>

  <repo_layout>
    .floyd/ is control-plane only. Application code is created/modified in the repo’s normal structure (src/, app/, packages/, etc.) unless Douglas explicitly specifies a subdirectory.
  </repo_layout>
</floyd_protocol>
EOF

# --- TMUX Config (THE PINK THEME) ---
# Uses Hot Pink (205) and Light Pink (218)
cat > "$PINK_CONF" <<'EOF'
# PINK FLOYD Theme
set -g status-style bg=colour205,fg=colour232,bold
set -g pane-border-style fg=colour218
set -g pane-active-border-style fg=colour205
set -g message-style bg=colour205,fg=colour232,bold
set -g clock-mode-colour colour205

# Status Bar Content
set -g status-left " #[bg=colour232,fg=colour205] 🎸 PINK FLOYD #[default] "
set -g status-right " #[fg=colour232]#{pane_current_path} "
EOF

# --- CLI ---
cat > "$CLI" <<'EOF'
 bash
set -euo pipefail

# Configuration
TEMPLATES="$HOME/.floyd/templates"
PINK_CONF="$HOME/.floyd/pink.tmux.conf"

# Resolve repo root
resolve_root() {
  if git rev-parse --show-toplevel >/dev/null 2>&1; then
    git rev-parse --show-toplevel
  else
    pwd
  fi
}

ROOT="$(resolve_root)"
FLOYD_DIR="$ROOT/.floyd"

require_templates() {
  local missing=0
  for f in master_plan.md scratchpad.md progress.md branch.md AGENT_INSTRUCTIONS.md stack.md; do
    if [[ ! -f "$TEMPLATES/$f" ]]; then
      echo "Missing template: $TEMPLATES/$f" >&2
      missing=1
    fi
  done
  [[ "$missing" -eq 0 ]] || exit 1
}

stamp_progress() {
  local ts
  ts="$(date '+%Y-%m-%d %H:%M:%S')"
  if ! grep -q "| Timestamp | Action Taken |" "$FLOYD_DIR/progress.md" 2>/dev/null; then
    cp "$TEMPLATES/progress.md" "$FLOYD_DIR/progress.md"
  fi
  echo "| $ts | Init | Ready | Awaiting User Input |" >> "$FLOYD_DIR/progress.md"
}

do_init() {
  require_templates
  mkdir -p "$FLOYD_DIR"
  cp "$TEMPLATES/master_plan.md" "$FLOYD_DIR/master_plan.md"
  cp "$TEMPLATES/scratchpad.md" "$FLOYD_DIR/scratchpad.md"
  cp "$TEMPLATES/branch.md" "$FLOYD_DIR/branch.md"
  cp "$TEMPLATES/AGENT_INSTRUCTIONS.md" "$FLOYD_DIR/AGENT_INSTRUCTIONS.md"
  cp "$TEMPLATES/progress.md" "$FLOYD_DIR/progress.md"
  cp "$TEMPLATES/stack.md" "$FLOYD_DIR/stack.md"
  stamp_progress
  echo "✅ FLOYD initialized in $FLOYD_DIR"
}

do_clean() {
  echo "🧹 Pruning progress log..."
  local archive_dir="$FLOYD_DIR/archive"
  local ts="$(date '+%Y%m%d_%H%M%S')"
  mkdir -p "$archive_dir"
  cp "$FLOYD_DIR/progress.md" "$archive_dir/progress_$ts.md"
  head -n 2 "$FLOYD_DIR/progress.md" > "$FLOYD_DIR/progress.tmp"
  tail -n 5 "$FLOYD_DIR/progress.md" >> "$FLOYD_DIR/progress.tmp"
  mv "$FLOYD_DIR/progress.tmp" "$FLOYD_DIR/progress.md"
  echo "   Done."
}

do_status() {
  echo "=== PINK FLOYD STATUS ==="
  echo "Current Branch: $(git branch --show-current 2>/dev/null || echo 'HEAD')"
  echo "--- Pending Tasks ---"
  grep -m 3 "\- \[ \]" "$FLOYD_DIR/master_plan.md" || echo "No open tasks."
  echo "--- Last Actions ---"
  tail -n 2 "$FLOYD_DIR/progress.md"
  echo "--- Stack Status ---"
  if grep -q "\[ \]" "$FLOYD_DIR/stack.md"; then
    echo "WARNING: Stack not fully defined. Check .floyd/stack.md"
  else
    echo "Stack defined."
  fi
  echo "========================="
}

do_run() {
  local agent_cmd="${1:-claude --dangerously-skip-permissions}"

  # 1. CHECK TMUX: If not in tmux, launch the PINK FLOYD session
  if [[ -z "${TMUX:-}" ]]; then
    echo "🎸 Entering the Machine (PINK FLOYD)..."
    # Ensure we pass the arguments correctly to the inner call
    tmux new-session -s pink-floyd -f "$PINK_CONF" "floyd run '$agent_cmd'; $SHELL"
    exit 0
  fi

  # 2. IF WE ARE HERE, WE ARE IN TMUX. RUN THE AGENT.
  local prompt_file=$(mktemp)
  
  echo "🚀 Assembling Context..."

  {
    echo "<system_context>"
    cat "$FLOYD_DIR/AGENT_INSTRUCTIONS.md"
    echo "</system_context>"

    echo "<project_state>"
    echo "### Master Plan"
    cat "$FLOYD_DIR/master_plan.md"
    echo "### Tech Stack"
    cat "$FLOYD_DIR/stack.md"
    echo "### Recent Progress"
    tail -n 10 "$FLOYD_DIR/progress.md"
    echo "</project_state>"

    echo "INSTRUCTION: Review the PRD/Blueprint in the root directory. Populate .floyd/stack.md. Update .floyd/master_plan.md and begin."
  } > "$prompt_file"

  # Launch Agent
  echo "🚀 Launching: $agent_cmd"
  cat "$prompt_file" | $agent_cmd
  
  rm "$prompt_file"
}

CMD="${1:-help}"
case "$CMD" in
  init)         do_init ;;
  clean)        do_clean ;;
  status)       do_status ;;
  run)          shift; do_run "${1:-claude --dangerously-skip-permissions}" ;;
  where)        echo "$FLOYD_DIR" ;;
  help|-h|--help) 
    echo "FLOYD CLI v2 (Pink Edition)"
    echo "Commands: init, run, clean, status"
    ;;
  *) echo "Unknown command: $CMD"; exit 1 ;;
esac
EOF
chmod +x "$CLI"

# --- Path Ensure ---
ensure_path_line() {
  local rc="$1"
  touch "$rc"
  if ! grep -q 'export PATH="$HOME/bin:$PATH"' "$rc"; then
    {
      echo ""
      echo "# FLOYD CLI"
      echo 'export PATH="$HOME/bin:$PATH"'
    } >> "$rc"
  fi
}

ensure_path_line "$HOME/.zshrc"
ensure_path_line "$HOME/.bashrc"
ensure_path_line "$HOME/.profile"
export PATH="$HOME/bin:$PATH"

echo "✅ PINK FLOYD installed."
echo "   Run 'floyd run' to enter the machine."
INSTALL_SCRIPT

# Run the script we just created
bash /tmp/install_pink_floyd.sh
rm /tmp/install_pink_floyd.sh

### Next Step
After pasting that block:
1. Run source ~/.zshrc (just to be safe).
2. Navigate to your project.
3. Run floyd init
4. Run floyd run (and watch your terminal turn pink).


-------

Here is a robust, unified toolset combining the best capabilities from the provided definitions. I have deduplicated overlapping tools (selecting the most capable version), merged complementary features (e.g., combining standard file reading with smart chunking), and categorized them for logical access.
**Robust Toolset Configuration**
JSON


{
  "tools": [
    {
      "name": "Task",
      "description": "Launch a new specialized sub-agent to handle complex, multi-step tasks autonomously. Use this for open-ended research, complex refactoring, or multi-step logic that requires maintaining separate context.",
      "input_schema": {
        "type": "object",
        "properties": {
          "description": { "type": "string", "description": "A short (3-5 word) description of the task" },
          "prompt": { "type": "string", "description": "The detailed task for the agent to perform" },
          "subagent_type": { 
            "type": "string", 
            "description": "The type of specialized agent (e.g., 'general-purpose', 'researcher', 'coder')" 
          }
        },
        "required": ["description", "prompt", "subagent_type"]
      }
    },
    {
      "name": "TodoManager",
      "description": "Manages structured todo lists for complex projects. Use this to track milestone-level tasks (e.g., 'Build Auth System', 'Setup Database'). Do not use for trivial steps.",
      "input_schema": {
        "type": "object",
        "properties": {
          "action": {
            "type": "string",
            "enum": ["set_tasks", "add_task", "move_to_task", "mark_all_done", "read_list"],
            "description": "Action to perform on the todo list."
          },
          "tasks": { "type": "array", "items": { "type": "string" }, "description": "List of tasks for 'set_tasks'" },
          "task": { "type": "string", "description": "Single task for 'add_task'" },
          "moveToTask": { "type": "string", "description": "Name of the task to focus on (marks previous as done)" }
        },
        "required": ["action"]
      }
    },
    {
      "name": "Bash",
      "description": "Executes bash commands in a persistent shell. Supports background processes, timeouts, and output capture. PREFER THIS over other shell tools.",
      "input_schema": {
        "type": "object",
        "properties": {
          "command": { "type": "string", "description": "The command to execute" },
          "run_in_background": { "type": "boolean", "description": "Run process in background (use BashOutput to read later)" },
          "timeout": { "type": "number", "description": "Timeout in ms (default 120000)" }
        },
        "required": ["command"]
      }
    },
    {
      "name": "BashOutput",
      "description": "Retrieves output from a background bash shell started with the Bash tool.",
      "input_schema": {
        "type": "object",
        "properties": {
          "bash_id": { "type": "string", "description": "ID of the background shell" },
          "filter": { "type": "string", "description": "Regex to filter output lines" }
        },
        "required": ["bash_id"]
      }
    },
    {
      "name": "Read",
      "description": "Reads file contents. For small files, returns full content. For large files (>2000 lines), provide a 'query' to get relevant chunks, or use 'offset/limit' for pagination.",
      "input_schema": {
        "type": "object",
        "properties": {
          "file_path": { "type": "string", "description": "Absolute path to the file" },
          "query": { "type": "string", "description": "Natural language query to find relevant sections in large files" },
          "offset": { "type": "number", "description": "Line number to start reading from" },
          "limit": { "type": "number", "description": "Number of lines to read" }
        },
        "required": ["file_path"]
      }
    },
    {
      "name": "Write",
      "description": "Writes or overwrites a file to the local filesystem.",
      "input_schema": {
        "type": "object",
        "properties": {
          "file_path": { "type": "string", "description": "Absolute path to the file" },
          "content": { "type": "string", "description": "The content to write" }
        },
        "required": ["file_path", "content"]
      }
    },
    {
      "name": "Edit",
      "description": "Performs exact string replacement in a file. Prefer MultiEdit for multiple changes.",
      "input_schema": {
        "type": "object",
        "properties": {
          "file_path": { "type": "string", "description": "Absolute path to the file" },
          "old_string": { "type": "string", "description": "Text to replace (must match exactly)" },
          "new_string": { "type": "string", "description": "Replacement text" },
          "replace_all": { "type": "boolean", "description": "Replace all occurrences" }
        },
        "required": ["file_path", "old_string", "new_string"]
      }
    },
    {
      "name": "MultiEdit",
      "description": "Perform multiple string replacements in a single file atomically.",
      "input_schema": {
        "type": "object",
        "properties": {
          "file_path": { "type": "string", "description": "Absolute path to the file" },
          "edits": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "old_string": { "type": "string" },
                "new_string": { "type": "string" },
                "replace_all": { "type": "boolean" }
              },
              "required": ["old_string", "new_string"]
            }
          }
        },
        "required": ["file_path", "edits"]
      }
    },
    {
      "name": "Grep",
      "description": "Fast regex search in files using ripgrep. Use for finding exact code patterns.",
      "input_schema": {
        "type": "object",
        "properties": {
          "pattern": { "type": "string", "description": "Regex pattern" },
          "path": { "type": "string", "description": "Directory to search (default: current)" },
          "glob": { "type": "string", "description": "Filter files (e.g., '*.ts')" },
          "output_mode": { "type": "string", "enum": ["content", "files_with_matches", "count"] }
        },
        "required": ["pattern"]
      }
    },
    {
      "name": "SearchCodebase",
      "description": "Semantic search engine. Use natural language to find relevant code when you don't know the exact keyword or location.",
      "input_schema": {
        "type": "object",
        "properties": {
          "query": { "type": "string", "description": "Natural language description of what you are looking for (e.g., 'auth middleware logic')" },
          "target_directories": { "type": "array", "items": { "type": "string" } }
        },
        "required": ["query"]
      }
    },
    {
      "name": "LS",
      "description": "Lists files in a directory.",
      "input_schema": {
        "type": "object",
        "properties": {
          "path": { "type": "string", "description": "Absolute path to directory" },
          "ignore": { "type": "array", "items": { "type": "string" }, "description": "Glob patterns to ignore" }
        },
        "required": ["path"]
      }
    },
    {
      "name": "Computer",
      "description": "Interact with a web browser using mouse and keyboard (Visual AI).",
      "input_schema": {
        "type": "object",
        "properties": {
          "action": {
            "type": "string",
            "enum": ["left_click", "right_click", "type", "screenshot", "wait", "scroll", "key", "left_click_drag", "double_click", "hover", "scroll_to"],
            "description": "The action to perform"
          },
          "coordinate": { "type": "array", "items": { "type": "number" }, "description": "(x, y) coordinates" },
          "text": { "type": "string", "description": "Text to type or key to press" },
          "tabId": { "type": "number", "description": "Target tab ID" }
        },
        "required": ["action", "tabId"]
      }
    },
    {
      "name": "InspectSite",
      "description": "Takes screenshots of URLs (live or localhost) to verify visual bugs or capture reference designs.",
      "input_schema": {
        "type": "object",
        "properties": {
          "urls": { "type": "array", "items": { "type": "string" }, "description": "URLs to capture" },
          "taskNameActive": { "type": "string" }
        },
        "required": ["urls"]
      }
    },
    {
      "name": "SearchWeb",
      "description": "Performs intelligent web search. Set 'isFirstParty' to true for official docs (Vercel, Next.js, etc).",
      "input_schema": {
        "type": "object",
        "properties": {
          "query": { "type": "string", "description": "Search query" },
          "isFirstParty": { "type": "boolean", "description": "Prioritize official documentation sources" }
        },
        "required": ["query"]
      }
    },
    {
      "name": "Packager",
      "description": "Installs system dependencies (via Nix) or language libraries (npm, pip, etc).",
      "input_schema": {
        "type": "object",
        "properties": {
          "dependency_list": { "type": "array", "items": { "type": "string" }, "description": "List of packages" },
          "install_or_uninstall": { "type": "string", "enum": ["install", "uninstall"] },
          "language_or_system": { "type": "string", "description": "'system', 'nodejs', 'python', etc." }
        },
        "required": ["dependency_list", "install_or_uninstall", "language_or_system"]
      }
    },
    {
      "name": "ExecuteSQL",
      "description": "Execute SQL queries to fix errors or inspect schema.",
      "input_schema": {
        "type": "object",
        "properties": {
          "sql_query": { "type": "string", "description": "The SQL query to execute" }
        },
        "required": ["sql_query"]
      }
    },
    {
      "name": "AskSecrets",
      "description": "Request missing API keys/Secrets from the user securely.",
      "input_schema": {
        "type": "object",
        "properties": {
          "secret_keys": { "type": "array", "items": { "type": "string" }, "description": "List of keys needed (e.g. ['OPENAI_API_KEY'])" },
          "user_message": { "type": "string", "description": "Explanation of why these are needed" }
        },
        "required": ["secret_keys", "user_message"]
      }
    },
    {
      "name": "WorkflowsRunConfig",
      "description": "Configure a long-running background task (e.g., starting a dev server).",
      "input_schema": {
        "type": "object",
        "properties": {
          "name": { "type": "string", "description": "Unique name for the task" },
          "command": { "type": "string", "description": "Shell command to run" },
          "wait_for_port": { "type": "integer", "description": "Port to wait for before confirming success" }
        },
        "required": ["name", "command"]
      }
    },
    {
      "name": "SuggestDeploy",
      "description": "Suggest to the user that the project is ready for deployment.",
      "input_schema": {
        "type": "object",
        "properties": {}
      }
    },
    {
      "name": "AskUser",
      "description": "Ask the user a question or request confirmation explicitly.",
      "input_schema": {
        "type": "object",
        "properties": {
          "text": { "type": "string", "description": "The question to ask" },
          "attachments": { "type": "array", "items": { "type": "string" } }
        },
        "required": ["text"]
      }
    }
  ]
}

-------

