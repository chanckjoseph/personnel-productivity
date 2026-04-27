# Debugging Demo: Bank Account Race Condition

## Scenario

You're debugging a banking application. Customers report that their account balances are incorrect after concurrent withdrawals. The app uses a simple `transfer()` function that reads balance, deducts, and writes back—but under high load, money mysteriously disappears.

This is a classic **race condition** bug. Let's walk through the complete 6-step scientific debugging process to find and fix it.

---

## The Bug (What We Don't Know Yet)

**bank.go:**
```go
func (account *Account) Withdraw(amount float64) error {
    balance := account.GetBalance()  // Read balance
    if balance < amount {
        return errors.New("insufficient funds")
    }
    
    // Some processing happens here...
    time.Sleep(10 * time.Millisecond)  // Simulate I/O
    
    account.SetBalance(balance - amount)  // Write new balance
    return nil
}
```

**The Problem:** Between reading and writing, another goroutine can read the old balance and write it back, causing lost updates.

**Example:**
- Account: $1000
- Thread 1 reads: $1000
- Thread 2 reads: $1000
- Thread 1 withdraws $200 → writes $800
- Thread 2 withdraws $300 → writes $700 (should be $500!)
- Result: Lost $200

---

## Complete Debugging Walkthrough

### Step 1: Start a Debugging Session

**Problem:** Customers report incorrect balances after concurrent operations.

```json
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "start",
    "bug_description": "Bank account balance becomes inconsistent during concurrent withdrawals. Money disappears from accounts after high-load testing. Affects ~2% of transactions."
  }
}
```

**Response:**
```json
{
  "session_id": "session_1714286400000000000",
  "step": "start",
  "content": [
    {
      "type": "text",
      "text": "=== DEBUG WORKFLOW STARTED ===\n\nBug: Bank account balance becomes inconsistent during concurrent withdrawals. Money disappears from accounts after high-load testing. Affects ~2% of transactions.\nSession: session_1714286400000000000\n\nNext: Use 'hypothesis' step to formulate a testable hypothesis\n"
    }
  ]
}
```

✅ **Session created.** Now we have a place to track our debugging.

---

### Step 2: Formulate Hypothesis

**Question:** What causes the balance to be wrong?

**Your Analysis:**
- Balance errors only happen under load (many concurrent requests)
- Errors don't happen with single requests
- The errors are inconsistent (unpredictable)

**Hypothesis:**
> "The Withdraw function has a race condition where multiple goroutines can read the same balance concurrently, and the last write wins, causing lost updates"

```json
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "hypothesis",
    "session_id": "session_1714286400000000000",
    "hypothesis": "Multiple goroutines read the same balance value concurrently before any write, and only the last write persists, causing earlier updates to be lost"
  }
}
```

**Response:**
```json
{
  "step": "hypothesis",
  "hypothesis_id": "hyp_1714286401000000000",
  "is_falsifiable": true,
  "validation_note": "Hypothesis is falsifiable",
  "content": [
    {
      "type": "text",
      "text": "Hypothesis recorded:\nMultiple goroutines read the same balance value concurrently before any write, and only the last write persists, causing earlier updates to be lost\n\nExpected outcome: None specified\n\nNext: Use 'predict' step to state what you expect to observe\n"
    }
  ]
}
```

✅ **Hypothesis recorded.** It's falsifiable (we can prove it true or false).

---

### Step 3: Make a Prediction

**Prediction:** If the hypothesis is true, then...

```json
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "predict",
    "session_id": "session_1714286400000000000",
    "prediction": "If I protect the read-modify-write with a mutex (critical section), then concurrent withdrawals will be atomic, and the balance will never become inconsistent, even under load. The total withdrawn will always match the balance reduction."
  }
}
```

**Response:**
```json
{
  "step": "predict",
  "content": [
    {
      "type": "text",
      "text": "=== STEP 3: PREDICTION RECORDED ===\n\nIf hypothesis is true, then: If I protect the read-modify-write with a mutex (critical section), then concurrent withdrawals will be atomic, and the balance will never become inconsistent, even under load. The total withdrawn will always match the balance reduction.\n\nNext: Use 'experiment' step to design and run the test\n"
    }
  ]
}
```

✅ **Prediction recorded.** This sets clear expectations for our test.

---

### Step 4: Design Experiment

**Experiment Design:** We'll run concurrent withdrawals and measure whether balances stay consistent.

```json
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "experiment",
    "session_id": "session_1714286400000000000",
    "steps": [
      "Create test account with initial balance $10,000",
      "Spawn 100 concurrent goroutines",
      "Each goroutine withdraws $50 (total: 5000 if no errors)",
      "Record final balance and sum of withdrawals",
      "Repeat 10 times and count inconsistencies",
      "Then add mutex protection to Withdraw() function",
      "Repeat the same test 10 times",
      "Compare error rates before and after"
    ],
    "independent_vars": [
      "Mutex protection (with/without)",
      "Concurrency level (100 goroutines)",
      "Withdrawal amount ($50 each)"
    ],
    "dependent_vars": [
      "Final balance",
      "Sum of withdrawals",
      "Inconsistency count",
      "Total money missing"
    ],
    "control_vars": [
      "Initial balance ($10,000)",
      "Number of iterations (10)",
      "No other transactions"
    ]
  }
}
```

**Response:**
```json
{
  "step": "experiment",
  "experiment_id": "exp_1714286402000000000",
  "content": [
    {
      "type": "text",
      "text": "=== STEP 4: EXPERIMENT DESIGNED ===\n\nSteps: 8\nIndependent vars: [\"Mutex protection (with/without)\", \"Concurrency level (100 goroutines)\", \"Withdrawal amount ($50 each)\"]\nDependent vars: [\"Final balance\", \"Sum of withdrawals\", \"Inconsistency count\", \"Total money missing\"]\nControl vars: [\"Initial balance ($10,000)\", \"Number of iterations (10)\", \"No other transactions\"]\n\nNext: Use 'analyze' step to record observations\n"
    }
  ]
}
```

✅ **Experiment designed.** We have a clear, controlled test plan.

---

### Step 5: Run Experiment & Analyze Results

**Test Phase 1: WITHOUT mutex protection**

```
Initial: $10,000
Run 1: Final = $9,850   (withdrew $150, expected $5,000)  ← LOST $4,850 ❌
Run 2: Final = $8,200   (withdrew $1,800, expected $5,000) ← LOST $3,200 ❌
Run 3: Final = $9,100   (withdrew $900, expected $5,000)   ← LOST $4,100 ❌
...
Average inconsistency: $4,100 lost per run (82% error rate)
```

**Test Phase 2: WITH mutex protection**

```
Initial: $10,000
Run 1: Final = $5,000   (withdrew exactly $5,000)  ✅
Run 2: Final = $5,000   (withdrew exactly $5,000)  ✅
Run 3: Final = $5,000   (withdrew exactly $5,000)  ✅
...
All 10 runs: Perfect consistency, 0 errors
```

**Record the Analysis:**

```json
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "analyze",
    "session_id": "session_1714286400000000000",
    "observations": "WITHOUT mutex: 10 test runs, average final balance $8,200 (should be $5,000). Total money lost: ~$41,000. Error rate: 82%. Errors are NOT consistent—sometimes lose $850, sometimes $4,850. WITH mutex: 10 test runs, all final balances exactly $5,000. Total lost: $0. Error rate: 0%. Difference is statistically significant (p < 0.001). Conclusion: Atomic operations eliminate ALL race conditions.",
    "conclusion": "supported"
  }
}
```

**Response:**
```json
{
  "step": "analyze",
  "conclusion": "supported",
  "content": [
    {
      "type": "text",
      "text": "=== STEP 5: ANALYSIS COMPLETE ===\n\nConclusion: SUPPORTED\n\nHypothesis was SUPPORTED by the evidence.\nNext: Use 'fix' step to apply the bug fix\n"
    }
  ]
}
```

✅ **Hypothesis CONFIRMED!** The race condition is the root cause.

---

### Step 6: Apply the Fix

**The Fix:** Add mutex protection to make the operation atomic.

**bank.go (BEFORE):**
```go
func (account *Account) Withdraw(amount float64) error {
    balance := account.GetBalance()
    if balance < amount {
        return errors.New("insufficient funds")
    }
    time.Sleep(10 * time.Millisecond)  // Simulate I/O
    account.SetBalance(balance - amount)
    return nil
}
```

**bank.go (AFTER):**
```go
var balanceMutex sync.Mutex

func (account *Account) Withdraw(amount float64) error {
    balanceMutex.Lock()  // ← ATOMIC SECTION START
    defer balanceMutex.Unlock()
    
    balance := account.GetBalance()
    if balance < amount {
        return errors.New("insufficient funds")
    }
    time.Sleep(10 * time.Millisecond)
    account.SetBalance(balance - amount)
    
    return nil  // ← ATOMIC SECTION END
}
```

**Record the Fix:**

```json
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "fix",
    "session_id": "session_1714286400000000000",
    "fix_description": "Added sync.Mutex around Withdraw() read-modify-write operation to make balance update atomic. Only one goroutine can hold lock at a time, preventing concurrent reads of stale balance. Lock released after write completes."
  }
}
```

**Response:**
```json
{
  "step": "fix",
  "bug_fixed": true,
  "content": [
    {
      "type": "text",
      "text": "=== STEP 6: BUG FIXED ===\n\nFix: Added sync.Mutex around Withdraw() read-modify-write operation to make balance update atomic. Only one goroutine can hold lock at a time, preventing concurrent reads of stale balance. Lock released after write completes.\n\nYou can now commit and push the fix using:\n- git_commit tool\n- git_push tool\n"
    }
  ]
}
```

✅ **Fix implemented!** Now commit and push.

---

## Commit the Fix

```json
{
  "tool": "git_commit",
  "arguments": {
    "message": "Fix: Prevent race condition in Account.Withdraw() with mutex protection\n\nDebugging Session: session_1714286400000000000\n\nROOT CAUSE:\nConcurrent goroutines read the same balance value before any goroutine writes.\nOnly the last write persists, causing earlier updates to be lost.\nExample: $1000 initial → Two $500 withdrawals → Final: $500 (should be $0)\n\nSOLUTION:\nProtect read-modify-write sequence with sync.Mutex lock to ensure atomicity.\nOnly one goroutine can enter the critical section at a time.\n\nTESTING:\nBefore fix: 82% failure rate, $41k lost across 100 concurrent withdrawals\nAfter fix: 0% failure rate, perfect consistency across all tests\nVerified with 10 iterations × 100 goroutines = 1000 concurrent ops\n\nImpact: Fixes reported balance inconsistencies affecting ~2% of transactions"
  }
}
```

---

## Complete Debugging Journey Summary

| Step | What We Did | Tool | Result |
|------|------------|------|--------|
| 1 | Created session for "disappearing money" | `debug_workflow start` | Session created ✅ |
| 2 | Hypothesized: "Race condition in read-modify-write" | `debug_workflow hypothesis` | Validated as falsifiable ✅ |
| 3 | Predicted: "Mutex will eliminate errors" | `debug_workflow predict` | Prediction recorded ✅ |
| 4 | Designed: Concurrent withdrawal test (with/without mutex) | `debug_workflow experiment` | 8 controlled steps ✅ |
| 5 | Analyzed: 82% error WITHOUT mutex → 0% error WITH mutex | `debug_workflow analyze` | Hypothesis SUPPORTED ✅ |
| 6 | Fixed: Added sync.Mutex to Withdraw() | `debug_workflow fix` | Bug fixed ✅ |
| 7 | Committed: Git commit with full debugging details | `git_commit` | Pushed to repo ✅ |

---

## Why the Scientific Method Worked

❌ **Without Scientific Method:**
- "Add locks everywhere" → Overkill, performance suffers
- "Maybe it's database" → Wrong focus, wastes time
- "Increase server resources" → Doesn't fix the bug
- "Pray and reload" → Nothing changes

✅ **With Scientific Method:**
1. **Observation** → Identified pattern: only happens under load
2. **Hypothesis** → Formed testable explanation: race condition
3. **Prediction** → Set clear expectations: mutex eliminates errors
4. **Experiment** → Measured before/after objectively (82% → 0%)
5. **Analysis** → Conclusion supported by data
6. **Fix** → Applied minimum necessary fix with proof

---

## Key Insights from This Demo

### 1. Falsifiability is Crucial
❌ Vague hypothesis: "Something is broken with concurrency"
✅ Falsifiable hypothesis: "Race condition causes lost updates BECAUSE of concurrent reads of stale balance"

### 2. Controlled Experiments Matter
❌ Random testing: "Run it a few times, seems okay"
✅ Controlled: Independent (with/without mutex), Dependent (final balance), Control (same initial state)

### 3. Metrics Tell the Story
❌ "It's broken": Subjective, non-quantifiable
✅ "82% error rate before, 0% after": Objective proof of fix

### 4. Documentation Enables Learning
The git commit now contains:
- Root cause
- Solution
- Test results
- Impact assessment

Future developers can understand WHY the mutex exists (not just THAT it exists).

---

## What Would Happen If We Iterated

If the hypothesis had been **REFUTED** (mutex didn't fix it):

```json
{
  "tool": "debug_workflow",
  "arguments": {
    "step": "iterate",
    "session_id": "session_1714286400000000000"
  }
}
```

We'd go back to formulate a **new hypothesis**:
- "The database is reading stale data from a cache"
- "The balance table has concurrent inserts, violating constraints"
- "Time synchronization causes phantom withdrawals"

And repeat the cycle until we found the real root cause.

---

## Running This Demo Yourself

1. **Copy the bank.go buggy code** into your project
2. **Start a debug session:**
   ```json
   {"tool": "debug_workflow", "step": "start", "bug_description": "..."}
   ```
3. **Follow the steps above** — formulate hypothesis, predict, experiment
4. **Analyze results** — mutex fix works
5. **Commit fix** with full debugging context

The session will be saved in `.devtools-mcp/sessions/session_1714286400000000000.json` for future reference.

---

## Comparison: Other Bugs You Could Debug This Way

This method works for ANY reproducible bug:

| Bug Type | Hypothesis | Experiment |
|----------|-----------|-----------|
| **Memory leak** | "Event listeners not unregistered on shutdown" | Add listener.clear(), measure heap size |
| **Slow queries** | "N+1 query problem in user loading" | Add query logging, compare counts |
| **Auth failures** | "JWT token expired before refresh" | Mock time, measure expiry vs refresh |
| **Flaky tests** | "Race condition in mock setup teardown" | Add mutex, run 100× |
| **API crashes** | "Null pointer when optional field missing" | Send requests without field, catch panic |

**The same 6-step process works for all of them.**

---

## Real-World Impact

**Before Fix:**
- Customer complaints increasing
- 2% transaction error rate
- Investigation was chaotic (guessing)

**After Fix:**
- 0% error rate
- Root cause documented
- Future developers understand why mutex exists
- Prevents regression (test can verify fix works)

**Time Saved:** Scientific approach identified root cause in hours, vs. days of trial-and-error.

---

## Next Steps

1. **Study the DEBUGGING.md** guide for full tool reference
2. **Try the Layer B modular tools** for more control:
   ```json
   {"tool": "formulate_hypothesis", "session_id": "...", "hypothesis": "..."}
   {"tool": "design_experiment", "session_id": "...", "steps": [...]}
   {"tool": "analyze_results", "session_id": "...", "observations": "..."}
   ```
3. **Track iterations** with `track_iteration` tool
4. **Commit fixes** with git_commit, reference session ID

**See Also:** [DEBUGGING.md](DEBUGGING.md) for tool reference and best practices.
