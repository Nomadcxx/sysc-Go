# # FLOYD-S SUPERCACHE: TECHNICAL IMPLEMENTATION BLUEPRINT
# 1\. DIRECTORY STRUCTURE
text
.floyd/
├── .cache/                    # Supercache root (created on first init)
│   ├── reasoning/            # Tier 1: Ephemeral reasoning frames
│   │   ├── active/
│   │   │   └── frame.json    # Current active reasoning frame (constantly updated)
│   │   └── archive/
│   │       └── frame_<timestamp>_<task_hash>.json  # Archived frames
│   │
│   ├── project/              # Tier 2: Project chronicle cache
│   │   ├── state_snapshot.json
│   │   ├── phase_summaries/
│   │   │   ├── phase_1_complete.json
│   │   │   ├── phase_2_in_progress.json
│   │   │   └── ...
│   │   └── context/
│   │       ├── open_files.json
│   │       ├── active_processes.json
│   │       └── port_bindings.json
│   │
│   └── vault/                # Tier 3: Solution pattern vault
│       ├── patterns/
│       │   ├── nextauth_with_middleware_v1.json
│       │   ├── dashboard_data_viz_card_v2.json
│       │   └── ...
│       ├── index/
│       │   ├── by_signature.json      # Quick lookup table
│       │   └── vector_index.bin       # Optional: for semantic search
│       └── metadata.json              # Vault stats & version
│
├── templates/                # Original FLOYD templates (unchanged)
├── master_plan.md
├── scratchpad.md
├── progress.md
├── branch.md
└── AGENT_INSTRUCTIONS.md    # Now includes Supercache protocol
# 2\. JSON SCHEMAS FOR CACHE FILES
# 2.1 Reasoning Frame (reasoning/active/frame.json)
json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ReasoningFrame",
  "type": "object",
  "required": ["frame_id", "task_id", "start_time", "cog_steps"],
  "properties": {
    "frame_id": {
      "type": "string",
      "format": "uuid",
      "description": "Unique identifier for this reasoning session"
    },
    "task_id": {
      "type": "string",
      "description": "Correlates to master_plan.md task or user query"
    },
    "start_time": {
      "type": "string",
      "format": "date-time"
    },
    "last_updated": {
      "type": "string",
      "format": "date-time"
    },
    "cog_steps": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["timestamp", "step_type", "content"],
        "properties": {
          "timestamp": {"type": "string", "format": "date-time"},
          "step_type": {
            "type": "string",
          "enum": ["FRAME_START", "COG_STEP", "FRAME_APPEND", "TOOL_CALL", "DECISION"]
          },
          "content": {"type": "string"},
          "tool_used": {"type": "string"},
          "tool_result_hash": {"type": "string"}  // MD5 of tool output
        }
      }
    },
    "current_focus": {
      "type": "string",
      "description": "What the agent is currently working on"
    },
    "validation_hash": {
      "type": "string",
      "description": "MD5 of frame content for integrity checking"
    }
  }
}
# 2.2 Project State Snapshot (project/state_snapshot.json)
json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ProjectStateSnapshot",
  "type": "object",
  "required": ["snapshot_time", "project_root", "active_branch"],
  "properties": {
    "snapshot_time": {"type": "string", "format": "date-time"},
    "project_root": {"type": "string"},
    "active_branch": {"type": "string"},
    "last_commit": {"type": "string"},
    "master_plan_phase": {"type": "string"},
    "recent_commands": {
      "type": "array",
      "items": {"type": "string"},
      "maxItems": 20
    },
    "active_ports": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "port": {"type": "integer"},
          "service": {"type": "string"},
          "pid": {"type": "integer"}
        }
      }
    },
    "recent_errors": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "timestamp": {"type": "string", "format": "date-time"},
          "tool": {"type": "string"},
          "error": {"type": "string"},
          "resolved": {"type": "boolean"}
        }
      }
    }
  }
}
# 2.3 Solution Pattern (vault/patterns/*.json)
json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "SolutionPattern",
  "type": "object",
  "required": ["signature", "trigger_terms", "implementation"],
  "properties": {
    "signature": {
      "type": "string",
      "pattern": "^[a-z][a-z0-9_]+(_v[0-9]+)?$",
      "description": "Unique machine-readable identifier"
    },
    "human_name": {"type": "string"},
    "trigger_terms": {
      "type": "array",
      "items": {"type": "string"},
      "minItems": 3,
      "description": "Natural language terms that should retrieve this pattern"
    },
    "category": {
      "type": "string",
      "enum": ["ui_component", "api_endpoint", "auth_flow", "database", "devops", "utility"]
    },
    "implementation": {
      "type": "string",
      "description": "The actual code/template (Markdown with code blocks)"
    },
    "validation_tests": {"type": "string"},
    "dependencies": {
      "type": "array",
      "items": {"type": "string"}
    },
    "created": {"type": "string", "format": "date-time"},
    "last_used": {"type": "string", "format": "date-time"},
    "success_count": {"type": "integer"},
    "failure_count": {"type": "integer"},
    "complexity_score": {
      "type": "number",
      "minimum": 0,
      "maximum": 1
    },
    "embedding_vector": {
      "type": "array",
      "items": {"type": "number"},
      "description": "Optional: 768-dim vector for semantic search"
    }
  }
}
# 3\. TOOL DEPENDENCIES & EXTERNAL SERVICES
# 3.1 External Service Dependencies


| Tool | External Dependency | Notes |
|---|---|---|
| SearchCodebase | Optional: Qdrant/Weaviate/Pinecone | For semantic search across large codebases. Can fall back to grep if not available. |
| Computer (Visual AI) | Active browser session + vision model | Requires a running browser and optional vision API (e.g., GPT-4V) for full functionality. |
| InspectSite | Headless Chrome/Firefox | Requires puppeteer/playwright installed. |
| Solution Vault Search | Optional: Vector database | For semantic pattern matching. Can use simple keyword matching as fallback. |
# 3.2 Pure Filesystem Tools (No External Dependencies)
* Read/Write/Edit/MultiEdit - Direct file operations
* LS/Grep - Filesystem traversal
* Bash/BashOutput - Local shell execution
* TodoManager - In-memory or file-based state
* CacheManager - Pure filesystem operations
* Packager - Uses system package managers (local)

⠀4. IMPLEMENTATION VS. EXTERNAL SETUP
# 4.1 Fully Implementable Today (Pure Bash/Node)
bash
# These components require NO external services
.floyd/.cache/ directory structure
CacheManager tool (filesystem operations only)
Reasoning Frame persistence
Project State Snapshots
Solution Vault (basic keyword-based version)
Cache coherence protocols
Frame validation & hashing
Automatic archive rotation
CLI commands: floyd cache-status, floyd cache-clean
# 4.2 Requires External Setup
bash
# These need additional components
1\. Vector Search for Solution Vault
- Requires: Qdrant/Weaviate instance OR
- Local sentence-transformers + FAISS

2\. Semantic Codebase Search
- Requires: Code embedding pipeline
- Optional: Separate search index

3\. Visual AI (Computer tool)
- Requires: Browser automation setup
- Optional: Vision API endpoint

4\. Cross-Session Cache Sync
- Optional: Redis for multi-machine sync
# 5\. CONCRETE IMPLEMENTATION SPEC
# File 1:/usr/local/bin/floyd-cache (CLI Extension)
bash
****
# FLOYD Cache Management CLI
# Usage: floyd cache-<command>

CACHE_ROOT="$HOME/.floyd/.cache"

case "$1" in
  "init")
    mkdir -p "$CACHE_ROOT"/{reasoning/{active,archive},project/{phase_summaries,context},vault/{patterns,index}}
    echo '{"version": "2.0", "created": "'$(date -Iseconds)'"}' > "$CACHE_ROOT/vault/metadata.json"
    ;;
  "status")
    du -sh "$CACHE_ROOT"/*
    find "$CACHE_ROOT" -name "*.json" | wc -l | xargs echo "Total JSON files:"
    ;;
  "clean")
    # Move old reasoning frames to archive
    find "$CACHE_ROOT/reasoning/active" -mtime +7 -name "*.json" -exec mv {} "$CACHE_ROOT/reasoning/archive/" \;
    # Clean old archives
    find "$CACHE_ROOT/reasoning/archive" -mtime +30 -name "*.json" -delete
    ;;
  "index-rebuild")
    # Rebuild vault index
    jq -n '[inputs | {signature: .signature, terms: .trigger_terms}]' \
      "$CACHE_ROOT/vault/patterns"/*.json > "$CACHE_ROOT/vault/index/by_signature.json"
    ;;
esac
# File 2:~/.floyd/templates/cache_protocols.md
markdown
**# Cache Protocol Implementation Guide**

**## Automatic Frame Commitment**
Rule: Commit reasoning frame when ANY of these occur:
1\. 5+ tool calls completed
2\. Master plan phase checkbox checked
3\. User asks a new, unrelated question
4\. Any tool returns an error

**## Integrity Validation**
On session start:
1\. Compute hash of active frame
2\. Compare with stored validation_hash
3\. If mismatch: trigger coherence rebuild

**## Pattern Extraction Heuristics**
Extract to Solution Vault when:
- Code passes all tests on first attempt
- User says "perfect" or "exactly what I needed"
- Component reused across 2+ projects
# File 3: CacheManager Tool Implementation
javascript
// Example Node.js implementation for CacheManager tool
const fs = require('fs').promises;
const path = require('path');
const crypto = require('crypto');

class CacheManager {
  constructor(cacheRoot = '~/.floyd/.cache') {
    this.cacheRoot = path.expanduser(cacheRoot);
  }

  async store(tier, key, value, options = {}) {
    const filepath = this._resolvePath(tier, key);
    
    // Add metadata
    const enriched = {
      data: value,
      stored: new Date().toISOString(),
      hash: this._computeHash(value),
      ...options
    };

    await fs.mkdir(path.dirname(filepath), { recursive: true });
    await fs.writeFile(filepath, JSON.stringify(enriched, null, 2));
    
    // Update indices if needed
    if (tier === 'solution_vault') {
      await this._updateVaultIndex(key, enriched);
    }
    
    return { success: true, hash: enriched.hash };
  }

  async retrieve(tier, key, query = null) {
    if (tier === 'solution_vault' && query) {
      return this._semanticSearch(query);
    }
    
    const filepath = this._resolvePath(tier, key);
    try {
      const data = await fs.readFile(filepath, 'utf8');
      const parsed = JSON.parse(data);
      
      // Validate hash if present
      if (parsed.hash && parsed.hash !== this._computeHash(parsed.data)) {
        throw new Error('Cache integrity check failed');
      }
      
      return parsed.data;
    } catch (error) {
      return { error: 'CACHE_MISS', message: error.message };
    }
  }

  _resolvePath(tier, key) {
    const map = {
      reasoning_frame: `reasoning/active/${key}.json`,
      project_chronicle: `project/${key}.json`,
      solution_vault: `vault/patterns/${key}.json`
    };
    return path.join(this.cacheRoot, map[tier] || `${tier}/${key}.json`);
  }

  _computeHash(content) {
    return crypto
      .createHash('md5')
      .update(typeof content === 'string' ? content : JSON.stringify(content))
      .digest('hex');
  }

  async _updateVaultIndex(signature, pattern) {
    const indexFile = path.join(this.cacheRoot, 'vault/index/by_signature.json');
    let index = [];
    
    try {
      index = JSON.parse(await fs.readFile(indexFile, 'utf8'));
    } catch (e) {
      index = [];
    }
    
    // Update or add entry
    const existing = index.findIndex(item => item.signature === signature);
    if (existing >= 0) {
      index[existing] = { signature, terms: pattern.data.trigger_terms };
    } else {
      index.push({ signature, terms: pattern.data.trigger_terms });
    }
    
    await fs.writeFile(indexFile, JSON.stringify(index, null, 2));
  }
}
# File 4: Integration Hook in AGENT_INSTRUCTIONS.md
xml
<!-- Add this to the existing protocol section -->
<cache_integration_hooks>
  <!-- Called after successful tool execution -->
  <hook event="tool_success">
    <action>CacheManager.store(tier="reasoning_frame", key="last_tool_result", value=tool_output)</action>
    <condition>tool_name in ["Write", "Edit", "Bash", "SearchCodebase"]</condition>
  </hook>
  
  <!-- Called before starting new task -->
  <hook event="task_start">
    <action>CacheManager.retrieve(tier="project_chronicle", key="state_snapshot")</action>
    <action>CacheManager.retrieve(tier="solution_vault", key="search", query=task_description)</action>
  </hook>
  
  <!-- Called on error for recovery points -->
  <hook event="tool_error">
    <action>CacheManager.store(tier="project_chronicle", key="error_snapshot_$(timestamp)", value=full_context)</action>
  </hook>
</cache_integration_hooks>
# File 5: Phase Completion Automator
bash
**#!/bin/bash**
# ~/.floyd/scripts/phase_complete.sh
# Auto-runs when master_plan.md phase is marked complete

PHASE="$1"
PROJECT_ROOT="$2"

# Create phase summary
SUMMARY=$(jq -n \
  --arg phase "$PHASE" \
  --arg completed "$(date)" \
  --argjson files "$(find src -type f -name "*.ts" -o -name "*.js" | wc -l)" \
  '{
    phase: $phase,
    completed: $completed,
    files_modified: $files,
    tests_passed: $(npm test 2>/dev/null | grep -c "✓"),
    next_steps: ["Review PR checklist", "Update documentation"]
  }')

echo "$SUMMARY" > "$PROJECT_ROOT/.floyd/.cache/project/phase_summaries/phase_${PHASE}_complete.json"

# Trigger pattern extraction if novel solutions found
if grep -q "novel_pattern" "$PROJECT_ROOT/.floyd/scratchpad.md"; then
  "$PROJECT_ROOT/.floyd/scripts/extract_patterns.sh"
fi
# 6\. DEPLOYMENT SEQUENCE
bash
# Day 1: Core Filesystem Cache
1\. Install directory structure
2\. Implement CacheManager (basic filesystem version)
3\. Add cache protocols to AGENT_INSTRUCTIONS.md
4\. Test: floyd cache-init && floyd run

# Day 2: Integration & Validation
1\. Add automatic frame commitment
2\. Implement integrity validation
3\. Add cache status to floyd CLI
4\. Test: Verify frame persistence across breaks

# Day 3: Solution Vault (Basic)
1\. Implement keyword-based pattern storage
2\. Add pattern extraction triggers
3\. Test: Create pattern, then retrieve for similar task

# Week 2: Optional Enhancements
1\. Vector search integration (if Qdrant/etc available)
2\. Cross-project cache sharing
3\. Cache performance analytics
# 7\. VALIDATION TEST SUITE
bash
# Test 1: Cache Persistence
echo "Test: Frame persistence" > test_input.txt
floyd run --task "Process this test file"
# Kill agent, restart
floyd run --task "What was I just doing?"
# Should show [FRAME_RESUME: Processing test file...]

# Test 2: Pattern Reuse
floyd run --task "Create a React modal dialog"
# Mark as excellent solution
floyd run --task "Create a notification popup"
# Should retrieve and adapt modal pattern

# Test 3: Integrity Recovery
# Manually corrupt .floyd/.cache/project/state_snapshot.json
floyd run --task "Continue work"
# Should trigger [COHERENCE_REBUILT] and recover
## This blueprint provides a complete, incremental implementation path. Start with the pure filesystem components (Day 1-3), which provide 80% of the value, then optionally add vector search and external services as needed. The architecture ensures graceful degradation—if advanced features aren't available, the system falls back to simpler but functional alternatives.
