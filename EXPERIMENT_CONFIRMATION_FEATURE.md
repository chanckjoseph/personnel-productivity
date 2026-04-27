# Experiment Confirmation Feature

## Overview

The MCP debug workflow now implements a **two-phase experiment design** system that requires explicit user confirmation before experiments are executed. This ensures experiments are:
- ✓ Consoled and sandboxed
- ✓ Using mock data only
- ✓ Well-planned and approved by user
- ✓ Safe from production database access

---

## Implementation Details

### Code Changes: `devtools-mcp/debug.go`

Updated `handleWorkflowExperiment()` function with two-phase workflow:

#### Phase 1: Design & Review (awaiting_confirmation)
```
User calls: experiment step with steps/variables

Response shows:
  ├─ ⚠️  SAFETY REQUIREMENTS CHECKLIST
  │   ├─ □ Experiment MUST use sandboxed environment (console)
  │   ├─ □ Experiment MUST use MOCK DATA ONLY
  │   ├─ □ NO access to production database
  │   ├─ □ NO destructive operations on real systems
  │   └─ □ Results reproducible in isolated environment
  ├─ EXPERIMENT PLAN
  │   ├─ Steps to execute (detailed list)
  │   ├─ Control Variables (held constant)
  │   ├─ Independent Variables (what we change)
  │   └─ Dependent Variables (what we measure)
  └─ ⏸️  USER CONFIRMATION REQUIRED
      └─ "Call this tool again with confirmed=true parameter"
```

**Response Status:** `"design_status": "awaiting_confirmation"`

#### Phase 2: Confirmation & Execution (confirmed)
```
User calls: experiment step with confirmed=true

Response shows:
  ├─ ✓ Safety Requirements ACKNOWLEDGED
  │   ├─ ✓ Using sandboxed/console environment
  │   ├─ ✓ Using mock data only
  │   ├─ ✓ No production database access
  │   └─ ✓ No destructive operations
  ├─ APPROVED EXPERIMENT PLAN
  │   ├─ Experiment ID: exp_<timestamp>
  │   ├─ Total steps: <number>
  │   └─ Execute these steps: (detailed list)
  ├─ 📊 VARIABLE TRACKING
  │   ├─ Control (constant): ...
  │   ├─ Independent (changed): ...
  │   └─ Dependent (measured): ...
  └─ ➡️  NEXT STEPS
      └─ Execute steps → capture observations → call analyze step
```

**Response Status:** `"design_status": "confirmed"`

---

## Workflow Example

### Before (Old Flow)
```
1. experiment step (design)
   └─> Immediately saved to session
       Ready for analysis
```

### After (New Flow)
```
1. experiment step (design)
   └─> awaiting_confirmation status
       ├─ Shows safety checklist
       ├─ Shows detailed plan
       └─ Awaits user approval

2. experiment step + confirmed=true
   └─> confirmed status
       ├─ Acknowledges safety requirements
       ├─ Approves experiment plan
       └─ Ready for execution
```

---

## Safety Features

### 1. Sandboxed Environment Requirement
```
User must confirm they will use:
  • Isolated test database (NOT production)
  • Local console/script execution
  • Separate VM or container (if possible)
```

### 2. Mock Data Enforcement
```
Explicit checklist item:
  ✓ Using mock data only
  ✓ No real customer/production data
  ✓ No sensitive information
```

### 3. Non-Destructive Operations
```
Approved operations:
  ✓ READ queries (SELECT)
  ✓ CREATE for test tables
  ✓ Temporary indexes
  ✗ DELETE from production
  ✗ ALTER production schema
  ✗ DROP tables
```

### 4. Reproducibility
```
User confirms experiment can be:
  • Re-run in same sandbox without side effects
  • Run by another team member on test server
  • Archived for future reference
  • Compared against baseline
```

---

## Usage Example

### Session: Database Performance Query

**Step 1: Design Experiment (First Call)**
```
INPUT:
  step: experiment
  steps: [
    "Create test database with mock customer data",
    "Run query WITHOUT index, measure time",
    "Create index on filter column",
    "Run query WITH index, measure time",
    "Compare performance"
  ]
  independent_vars: ["Presence of index on filter column"]
  dependent_vars: ["Query execution time", "Query plan type"]
  control_vars: ["Database engine", "Data size", "Query complexity"]

OUTPUT:
  design_status: "awaiting_confirmation"
  
  Displays:
  ⚠️  SAFETY REQUIREMENTS:
    □ Experiment MUST use sandboxed environment (console)
    □ Experiment MUST use MOCK DATA ONLY
    □ NO access to production database
    □ NO destructive operations on real systems
    □ Results reproducible in isolated environment
  
  [Shows full experiment plan]
  
  ⏸️  USER CONFIRMATION REQUIRED:
    Review the experiment plan above carefully.
    Verify it meets ALL safety requirements.
    THEN call this tool again with confirmed=true parameter
```

**Step 2: Confirm & Approve (Second Call)**
```
INPUT:
  step: experiment
  steps: [same as above]
  independent_vars: [same as above]
  dependent_vars: [same as above]
  control_vars: [same as above]
  confirmed: true          ← NEW PARAMETER

OUTPUT:
  design_status: "confirmed"
  
  Displays:
  ✓ Safety Requirements ACKNOWLEDGED:
    ✓ Using sandboxed/console environment
    ✓ Using mock data only
    ✓ No production database access
    ✓ No destructive operations
  
  APPROVED EXPERIMENT PLAN:
    Experiment ID: exp_1777313024453947400
    Total steps: 5
    
    Execute these steps:
      1. Create test database with mock customer data
      2. Run query WITHOUT index, measure time
      3. Create index on filter column
      4. Run query WITH index, measure time
      5. Compare performance
  
  ➡️  NEXT STEPS:
    1. Execute the experiment steps in a sandboxed environment
    2. Capture ALL observations and measurements
    3. Call 'analyze' step with detailed observations
```

---

## Benefits

| Aspect | Benefit |
|--------|---------|
| **Safety** | Explicit checklist prevents accidental production access |
| **Transparency** | User clearly sees experiment plan before approval |
| **Auditability** | Session logs show explicit "confirmed" approval point |
| **Education** | Checklist teaches proper experiment design practices |
| **Reversibility** | User can reject design and redesign before execution |
| **Accountability** | User owns the decision to run experiment |

---

## Technical Details

### Code Location
- File: `devtools-mcp/debug.go`
- Function: `handleWorkflowExperiment()`
- Lines: ~250-380 (updated)

### New Parameters
- `confirmed` (optional boolean)
  - Default: `false`
  - Effect: If true, skips design phase and executes confirmed plan

### Response Fields
- `design_status` (string): "awaiting_confirmation" or "confirmed"
- `experiment_id` (string): Unique experiment identifier
- `content` (array): Formatted guidance text

### Session Storage
- Location: `.devtools-mcp/sessions/session_<ID>.json`
- Persisted: All experiments with full variable tracking
- Query: Use `debug_session_history` to view full session

---

## Next Steps to Test

1. **Restart VS Code** (to refresh MCP process)
   ```
   Or run: taskkill /F /IM node.exe (if MCP runs via Node)
   ```

2. **Create new debug session**
   ```
   debug_workflow step=start bug_description="Test bug"
   ```

3. **Run through workflow**
   ```
   debug_workflow step=hypothesis hypothesis="..."
   debug_workflow step=predict prediction="..."
   debug_workflow step=experiment steps=[...] independent_vars=[...]
   ```

4. **First response shows awaiting_confirmation**
   - Verify safety checklist visible
   - Verify experiment plan detailed

5. **Call again with confirmed=true**
   ```
   debug_workflow step=experiment ... confirmed=true
   ```

6. **Second response shows confirmed**
   - Verify "✓ Safety Requirements ACKNOWLEDGED"
   - Verify experiment marked "confirmed"

---

## Configuration

No additional configuration needed. The feature is built into the MCP server.

To modify safety checklist, edit the guidance strings in `handleWorkflowExperiment()` function in `debug.go`.

---

## Troubleshooting

**Issue:** Old response format (immediately saves experiment)
- **Cause:** VS Code caching old MCP process
- **Solution:** Restart VS Code or kill MCP process

**Issue:** "awaiting_confirmation" not recognized
- **Cause:** Using old MCP binary
- **Solution:** Verify `devtools-mcp.exe` timestamp matches rebuild time

**Issue:** No safety checklist displayed
- **Cause:** Hitting old server cached in memory
- **Solution:** Force MCP process restart

---

## Version Info

- **Feature Added:** 2026-04-27
- **MCP Version:** Latest with confirmation feature
- **Binary:** `devtools-mcp/bin/devtools-mcp.exe`
- **Last Build:** After confirmation feature implementation
