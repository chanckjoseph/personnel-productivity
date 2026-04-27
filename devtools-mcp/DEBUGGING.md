# Scientific Debugging Guide

## Overview

The devtools-mcp debugging tools implement the **scientific method** for systematic, reproducible bug fixing. This framework helps you debug complex issues methodically instead of guessing.

### The 6-Step Scientific Debugging Process

```
1. Observation & Questioning  →  Define what's broken
                                  ↓
2. Hypothesis Formulation      →  Propose a testable explanation
                                  ↓
3. Prediction                  →  State expected outcome if hypothesis is true
                                  ↓
4. Experimentation             →  Test the prediction with controlled steps
                                  ↓
5. Data Analysis               →  Evaluate if results support hypothesis
                                  ↓
6. Fix & Iteration             →  Apply fix or repeat with new hypothesis
```

---

## Using the Tools

### Three Layers of Tools

**Layer A: Interactive Orchestrator (simplest)**
- Single tool: `debug_workflow`
- Guides you step-by-step through all 6 phases
- Best for: Agents, automated workflows, guided debugging

**Layer B: Modular Tools (most flexible)**
- `formulate_hypothesis` - Validate falsifiable hypothesis
- `design_experiment` - Structure test design
- `analyze_results` - Evaluate experiment outcome
- `track_iteration` - View progress history

**Layer C: Session Management (foundation)**
- `start_debug_session` - Create new session
- `get_session_state` - Query session metadata
- `update_session_context` - Store arbitrary context

---

## Quick Start Example

### Using the Interactive Workflow (Recommended for beginners)

```json
1. Start a session:
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "start",
    "bug_description": "API returns 500 error when creating users with special characters"
  }
}

Response: {"session_id": "session_1714286400000000000", ...}

2. Formulate hypothesis:
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "hypothesis",
    "session_id": "session_1714286400000000000",
    "hypothesis": "The API validation regex does not properly escape special characters, causing a regex error"
  }
}

3. Make prediction:
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "predict",
    "session_id": "session_1714286400000000000",
    "prediction": "If I bypass the regex validation, the user creation will succeed"
  }
}

4. Design experiment:
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "experiment",
    "session_id": "session_1714286400000000000",
    "steps": [
      "Enable debug logging in validation",
      "Create user with special chars: !@#$%",
      "Check logs for regex error"
    ],
    "independent_vars": ["user input string"],
    "dependent_vars": ["validation error", "regex match result"]
  }
}

5. Analyze results:
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "analyze",
    "session_id": "session_1714286400000000000",
    "observations": "Logs show: regex.MatchString() fails on '!@#$%' input with error about special character escaping",
    "conclusion": "supported"
  }
}

6. Apply fix:
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "fix",
    "session_id": "session_1714286400000000000",
    "fix_description": "Updated regex pattern to use regexp.QuoteMeta() for user input escaping"
  }
}
```

---

## Tool Reference

### Layer A: debug_workflow

**Single tool orchestrates entire 6-step workflow**

**Parameters:**
- `step` (required): One of `start`, `hypothesis`, `predict`, `experiment`, `analyze`, `fix`, `iterate`
- `session_id` (required except for `start`): Session ID from previous `start` call
- **For `start`:**
  - `bug_description` - Description of the bug

- **For `hypothesis`:**
  - `hypothesis` - Testable hypothesis statement (must be falsifiable)

- **For `predict`:**
  - `prediction` - Expected outcome if hypothesis is true

- **For `experiment`:**
  - `steps` - Array of experiment steps
  - `independent_vars` - Variables to manipulate (optional)
  - `dependent_vars` - Variables to measure (optional)
  - `control_vars` - Variables to keep constant (optional)

- **For `analyze`:**
  - `observations` - What was actually observed
  - `conclusion` - One of: `supported`, `refuted`, `inconclusive`

- **For `fix`:**
  - `fix_description` - Description of the bug fix

- **For `iterate`:**
  - No additional parameters

**Example: Start a session**
```json
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "start",
    "bug_description": "Database connection pool exhausts under load"
  }
}
```

---

### Layer B: Modular Tools

#### formulate_hypothesis

Validates and stores a hypothesis. **Rejects if not falsifiable.**

**Parameters:**
- `session_id` - Session ID
- `hypothesis_text` - Hypothesis statement (must be falsifiable)
- `expected_outcome` - Expected outcome (optional)

**Validation Rules:**
- ❌ Rejected patterns: "might", "could", "maybe", "probably", "I think"
- ✅ Required patterns: "will", "causes", "results in", "prevents", "leads to"
- ❌ Rejected words: "always true", "good", "bad"

**Example:**
```json
{
  "tool": "formulate_hypothesis",
  "arguments": {
    "session_id": "session_12345",
    "hypothesis_text": "Connection pool exhaustion causes 500 errors because pool.acquire() blocks indefinitely when limit is reached",
    "expected_outcome": "If I increase pool size to 100, errors will stop under load"
  }
}
```

#### design_experiment

Structures experiment design with controlled variables.

**Parameters:**
- `session_id` - Session ID
- `hypothesis_id` - ID of hypothesis to test
- `steps` - Array of steps to execute (required)
- `independent_vars` - Variables you manipulate (optional)
- `dependent_vars` - Variables you measure (optional)
- `control_vars` - Variables you keep constant (optional)

**Example:**
```json
{
  "tool": "design_experiment",
  "arguments": {
    "session_id": "session_12345",
    "hypothesis_id": "hyp_1234567",
    "steps": [
      "Set pool maxSize to 20",
      "Run 100 concurrent requests",
      "Record response times and errors",
      "Check pool metrics"
    ],
    "independent_vars": ["pool maxSize"],
    "dependent_vars": ["response time", "error count", "active connections"],
    "control_vars": ["request timeout", "database availability"]
  }
}
```

#### analyze_results

Analyzes experiment results to validate hypothesis.

**Parameters:**
- `session_id` - Session ID
- `experiment_id` - ID of experiment
- `observations` - What was actually observed
- `conclusion` - One of: `supported`, `refuted`, `inconclusive`

**Example:**
```json
{
  "tool": "analyze_results",
  "arguments": {
    "session_id": "session_12345",
    "experiment_id": "exp_1234567",
    "observations": "With maxSize=20: 8% errors. With maxSize=100: 0% errors. Pool metrics show connection limit was hit at 20 connections.",
    "conclusion": "supported"
  }
}
```

#### track_iteration

Displays debugging progress and iteration history.

**Parameters:**
- `session_id` - Session ID

**Example:**
```json
{
  "tool": "track_iteration",
  "arguments": {
    "session_id": "session_12345"
  }
}
```

---

### Layer C: Session Management

#### start_debug_session

Creates a new debugging session.

**Parameters:**
- `bug_description` - Description of the bug being debugged (optional)

**Example:**
```json
{
  "tool": "start_debug_session",
  "arguments": {
    "bug_description": "User authentication fails with Okta SSO"
  }
}
```

#### get_session_state

Retrieves current session state and progress.

**Parameters:**
- `session_id` - Session ID to retrieve

**Example:**
```json
{
  "tool": "get_session_state",
  "arguments": {
    "session_id": "session_12345"
  }
}
```

#### update_session_context

Stores arbitrary metadata in a session.

**Parameters:**
- `session_id` - Session ID
- `key` - Context key
- `value` - Value (any type)

**Example:**
```json
{
  "tool": "update_session_context",
  "arguments": {
    "session_id": "session_12345",
    "key": "affected_users",
    "value": 500
  }
}
```

---

## Best Practices

### ✅ DO

1. **Make hypotheses falsifiable**
   - "System will respond in < 100ms" ✅
   - "System might be slow" ❌

2. **Control variables in experiments**
   - Keep unchanged: database, network, load
   - Vary one thing: pool size, timeout, retry count

3. **Record observations precisely**
   - "Error rate 15%" ✅
   - "Pretty slow" ❌

4. **Use git_commit/git_push to save fixes**
   ```json
   {
     "tool": "git_commit",
     "arguments": {
       "message": "Fix: Increase connection pool size from 20 to 100\n\nSession: session_12345\nConclusion: Pool exhaustion was causing 500 errors"
     }
   }
   ```

### ❌ DON'T

1. **Vague hypotheses**
   - ❌ "Something is wrong with caching"
   - ✅ "Redis cache misses cause N+1 queries"

2. **Multiple variables at once**
   - ❌ "Change pool size AND increase timeout AND add retries"
   - ✅ "Change pool size from 20 to 100 (keep other vars constant)"

3. **Skip prediction**
   - The prediction forces you to think critically
   - Surprise results teach you something new

4. **Ignore inconclusive results**
   - Inconclusive → refine experiment design or hypothesis
   - Iterate until you get clarity

---

## Session Storage

Sessions are stored in `.devtools-mcp/sessions/` within your project:

```
my-project/
├── .devtools-mcp/
│   └── sessions/
│       ├── session_1714286400000000000.json  ← Session file
│       ├── session_1714286401000000000.json
│       └── ...
├── src/
├── tests/
└── .gitignore (add: .devtools-mcp/sessions/)
```

Each session file contains:
- Bug description
- All hypotheses tested
- All experiments run
- Observations and conclusions
- Iteration history
- Arbitrary metadata you stored

**Audit Trail:** Sessions are never deleted, providing a complete debugging history.

---

## Example Debugging Sessions

### Session 1: Connection Pool Exhaustion

```
Bug: "API returns 500 errors under load"

Hypothesis 1: Connection pool is too small
  → Experiment: Increase from 20 to 100
  → Result: SUPPORTED ✅
  → Fix: Set pool.maxSize = 100

Bug Fixed! ✅
```

### Session 2: Memory Leak Investigation

```
Bug: "Server memory increases over time"

Hypothesis 1: Event listeners aren't unregistered
  → Experiment: Add listeners.clear() on shutdown
  → Result: REFUTED (memory still increases)

Hypothesis 2: Request buffer is not flushed
  → Experiment: Add buffer.reset() after request
  → Result: SUPPORTED ✅
  → Fix: Auto-clear buffer after each response

Bug Fixed! ✅
```

### Session 3: Multi-Iteration Debugging

```
Bug: "User registration fails for 2% of users"

Iteration 1:
  Hypothesis: Email validation too strict
  → Result: REFUTED

Iteration 2:
  Hypothesis: Database unique constraint on normalized email
  → Result: SUPPORTED ✅
  → Fix: Normalize emails before insert

Bug Fixed! ✅
```

---

## Integration with Existing Tools

### Commit bug fixes with metadata

```json
{
  "tool": "git_commit",
  "arguments": {
    "message": "Fix: Connection pool exhaustion causing 500 errors under load\n\nDebugging:\n- Session: session_1714286400000000000\n- Hypothesis: Pool size too small\n- Fix: Increased max connections from 20 to 100\n- Verified: Error rate dropped from 8% to 0% under load test"
  }
}
```

### Check project structure before designing experiments

```json
{
  "tool": "get_project_structure",
  "arguments": {
    "max_depth": 3
  }
}
```

### Monitor git status while debugging

```json
{
  "tool": "git_status",
  "arguments": {}
}
```

---

## Fail Early, Fail Fast Principle

**All debugging tools enforce strict validation:**

- ❌ Hypothesis without falsifiable language → rejected
- ❌ Experiment with missing required fields → rejected
- ❌ Conclusion with invalid value → rejected
- ❌ Session not found → rejected immediately (no fallback)

**Why?** This forces you to think clearly and prevents accumulating broken state.

---

## Troubleshooting

**"Session not found"**
- Check session_id is correct
- Sessions are stored in `.devtools-mcp/sessions/`
- Run `get_session_state` with your session_id to verify

**"Hypothesis is not falsifiable"**
- Add testable outcome language: "will", "causes", "results in", "prevents"
- Remove vague words: "might", "could", "probably", "I think"
- Example: "Cache bypass WILL reduce query time by 50%" ✅

**"Experiment with 0 steps"**
- steps array must have at least one element
- Each step should be actionable: "Set X to Y", "Run Z", "Measure W"

---

## Session File Format

Example session file (JSON):

```json
{
  "id": "session_1714286400000000000",
  "start_time": "2026-04-27T12:00:00Z",
  "bug_description": "API returns 500 errors under load",
  "hypotheses": [
    {
      "id": "hyp_1714286401000000000",
      "hypothesis_text": "Connection pool is too small",
      "is_falsifiable": true
    }
  ],
  "experiments": [
    {
      "id": "exp_1714286402000000000",
      "steps": ["Increase pool maxSize to 100"],
      "hypothesis_supported": true,
      "actual_observations": "Error rate dropped from 8% to 0%"
    }
  ],
  "bug_fixed": true,
  "fix_description": "Set pool.maxSize = 100",
  "iteration_count": 1,
  "metadata": {
    "affected_users": 500,
    "priority": "critical"
  }
}
```

---

## See Also

- [devtools-mcp README](README.md) - Main documentation
- [git_commit tool](README.md#git_commit) - Commit fixes
- [git_push tool](README.md#git_push) - Push to GitHub
