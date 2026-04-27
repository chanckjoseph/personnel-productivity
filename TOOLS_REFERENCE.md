# MCP Tools Reference - Organized by Category

All 12 tools with clear categorical prefixes for agents to understand the workflow.

## Tools by Category

### 🏗️ Project Tools (Inspection & Analysis)
| Tool | Purpose | Usage |
|------|---------|-------|
| `project_structure` | Get directory tree and project organization | Understand codebase layout before debugging |

### 📚 Git Tools (Version Control)
| Tool | Purpose | Usage |
|------|---------|-------|
| `git_status` | Get repo URL, current branch, modified files | Check what's changed and where |
| `git_commit` | Stage + commit all changes with message | Save debugging progress with context |
| `git_push` | Push commits to GitHub with auth | Deploy fixes and persist changes |

### 🔬 Debug Tools - Session Management
| Tool | Purpose | Usage |
|------|---------|-------|
| `debug_start_session` | Create new debugging session | Begin systematic debugging workflow |
| `debug_session_state` | Get current session state & metadata | Check progress, review hypotheses |
| `debug_update_context` | Store arbitrary context in session | Save environment details, error logs |

### 🧪 Debug Tools - Scientific Workflow (Individual)
| Tool | Purpose | Usage | Step |
|------|---------|-------|------|
| `debug_formulate_hypothesis` | Record testable hypothesis | State what you think is wrong | 2 |
| `debug_design_experiment` | Design controlled experiment | Plan how to test the hypothesis | 4 |
| `debug_analyze_results` | Analyze & conclude hypothesis validity | Compare prediction vs observation | 5 |
| `debug_session_history` | Show iteration history | Review debugging progress & lessons | ∞ |

### 🚀 Debug Tools - Orchestrator
| Tool | Purpose | Usage |
|------|---------|-------|
| `debug_workflow` | Interactive 6-step orchestrator | Guided debugging from start to fix |

---

## Recommended Usage Patterns

### Pattern 1: Fully Guided (Simplest for Agents)
Use `debug_workflow` orchestrator - it handles all 6 steps:

```json
{"tool": "debug_workflow", "step": "start", "bug_description": "..."}
{"tool": "debug_workflow", "step": "hypothesis", "session_id": "...", "hypothesis": "..."}
{"tool": "debug_workflow", "step": "predict", "session_id": "...", "prediction": "..."}
{"tool": "debug_workflow", "step": "experiment", "session_id": "...", "steps": [...]}
{"tool": "debug_workflow", "step": "analyze", "session_id": "...", "observations": "..."}
{"tool": "debug_workflow", "step": "fix", "session_id": "...", "fix_description": "..."}
```

### Pattern 2: Modular (Power Users)
Use individual tools for fine control:

```json
{"tool": "debug_start_session", "bug_description": "..."}
→ returns session_id

{"tool": "debug_formulate_hypothesis", "session_id": "...", "hypothesis_text": "..."}
→ records hypothesis, returns hypothesis_id

{"tool": "debug_design_experiment", "session_id": "...", "hypothesis_id": "...", "steps": [...]}
→ designs experiment, returns experiment_id

{"tool": "debug_analyze_results", "session_id": "...", "experiment_id": "...", "observations": "...", "conclusion": "supported"}
→ analyzes and validates hypothesis

{"tool": "debug_session_history", "session_id": "..."}
→ shows full iteration history
```

### Pattern 3: Mixed (Most Flexible)
Start with session, then use individual tools:

```json
{"tool": "debug_start_session", "bug_description": "..."}
→ use individual tools for steps
{"tool": "debug_formulate_hypothesis", ...}
{"tool": "debug_design_experiment", ...}
→ then use workflow to orchestrate remaining steps
{"tool": "debug_workflow", "step": "analyze", ...}
```

### Pattern 4: Git Integration
After debugging, commit and push:

```json
{"tool": "debug_workflow", "step": "fix", "session_id": "...", "fix_description": "..."}
→ describes the fix

{"tool": "git_status"}
→ verify what changed

{"tool": "git_commit", "message": "Fix: [description with session_id reference]"}
→ commit with debugging context

{"tool": "git_push"}
→ push to repo
```

---

## Naming Convention Explanation

### Prefix Strategy

**`project_*`** — Tools for understanding project structure
- `project_structure` — Inspect codebase organization

**`git_*`** — Tools for version control operations
- `git_status` — Check repo state
- `git_commit` — Commit changes
- `git_push` — Push to remote

**`debug_*`** — Tools for scientific debugging workflow
- `debug_start_session` — Create debug session (S step 1)
- `debug_session_state` — Query session (helper)
- `debug_update_context` — Store session metadata (helper)
- `debug_formulate_hypothesis` — Record hypothesis (step 2)
- `debug_design_experiment` — Design experiment (step 4)
- `debug_analyze_results` — Analyze results (step 5)
- `debug_session_history` — Show history (helper)
- `debug_workflow` — Orchestrate all 6 steps (master tool)

### Why This Naming?

1. **Clear categories** — Agent instantly knows tool purpose by prefix
2. **Consistent ordering** — Tools appear grouped in alphabetical lists
3. **Self-documenting** — Name explains what tool does and where it fits
4. **No ambiguity** — No confusion between orchestrator and individual tools
5. **Scalable** — Easy to add more tools later (e.g., `debug_deploy`, `debug_rollback`)

---

## Complete Tool Flow

```
AGENT STARTS
    ↓
[Use project_structure to understand codebase]
    ↓
[Identify bug: "money disappears"]
    ↓
debug_start_session (create debugging session)
    ↓
[Either Path A or B:]

PATH A: GUIDED (debug_workflow)
    debug_workflow step=start → debug_workflow step=hypothesis
    → debug_workflow step=predict → debug_workflow step=experiment
    → debug_workflow step=analyze → debug_workflow step=fix

PATH B: MODULAR (individual tools)
    debug_formulate_hypothesis → debug_design_experiment
    → debug_analyze_results → debug_session_history
    
[BOTH PATHS converge here:]
    ↓
git_status (check what changed)
    ↓
git_commit (commit with debugging context)
    ↓
git_push (deploy fix)
    ↓
AGENT COMPLETE
```

---

## Quick Copy-Paste Reference

### Start a Debug Session
```json
{"tool": "debug_start_session", "bug_description": "Description of what's wrong"}
```

### Formulate Hypothesis
```json
{"tool": "debug_formulate_hypothesis", "session_id": "...", "hypothesis_text": "If X then Y because Z"}
```

### Design Experiment
```json
{"tool": "debug_design_experiment", "session_id": "...", "hypothesis_id": "...", "steps": ["step 1", "step 2", "..."], "independent_vars": ["var1"], "dependent_vars": ["var2"], "control_vars": ["var3"]}
```

### Analyze Results
```json
{"tool": "debug_analyze_results", "session_id": "...", "experiment_id": "...", "observations": "What we saw...", "conclusion": "supported"}
```

### Show History
```json
{"tool": "debug_session_history", "session_id": "..."}
```

### Commit Fix
```json
{"tool": "git_commit", "message": "Fix: [description]\n\nDebugging Session: [session_id]\nRoot cause: [what caused it]\nSolution: [what fixed it]"}
```

---

## See Also

- [DEBUGGING_DEMO.md](DEBUGGING_DEMO.md) — Real-world example with bank account race condition
- [devtools-mcp/DEBUGGING.md](devtools-mcp/DEBUGGING.md) — Complete tool documentation
- [devtools-mcp/README.md](devtools-mcp/README.md) — MCP server setup instructions
