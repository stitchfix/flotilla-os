# ARA command_hash Fix: Implementation Locations

## ✅ STATUS: IMPLEMENTED

**All code changes have been completed.** This document now serves as a record of what was changed.

**Changes made:**
1. ✅ Added command_hash calculation from command in `services/execution.go`
2. ✅ Removed description-based hash calculation from `flotilla/endpoints.go` (3 locations)
3. ✅ Optimized SQL query in `state/pg_queries.go` to use direct parameter
4. ✅ Updated call site in `execution/adapter/eks_adapter.go` with NULL check

**Remaining work:**
- ⏳ Add unit tests (see Testing Plan section)
- ⏳ Deploy and monitor (see Success Criteria section)

---

## Executive Summary

The `command_hash` bug required moving hash calculation from the API layer (where only description is available) to the execution service layer (where the actual command is finalized).

## Current Broken Flow

```
1. API Layer (flotilla/endpoints.go:451-453, 514-516, 592-593)
   ├─ Receives execution request
   ├─ Sets: lr.CommandHash = MD5(description)  ❌ WRONG
   └─ Passes to execution service

2. Execution Service (services/execution.go:320-327)
   ├─ Constructs final command from template/request
   ├─ Command is now finalized  ✓
   └─ But hash was already set from description  ❌

3. Database (state/pg_state_manager.go:1168)
   └─ Stores the wrong hash from step 1  ❌

4. EKS Adapter (execution/adapter/eks_adapter.go:109)
   ├─ Final command formatting
   └─ Hash still wrong  ❌

5. ARA Lookup (execution/adapter/eks_adapter.go:369)
   └─ Uses wrong hash to query historical data  ❌
```

## Fixed Flow

```
1. API Layer (flotilla/endpoints.go)
   ├─ Receives execution request
   └─ Does NOT set command_hash (remove this code)  ✓

2. Execution Service (services/execution.go:359)
   ├─ Constructs final command
   ├─ Calculates: fields.CommandHash = MD5(command)  ✓ NEW
   └─ Passes to CreateRun

3. Database (state/pg_state_manager.go:1168)
   └─ Stores correct hash  ✓

4. EKS Adapter (execution/adapter/eks_adapter.go:109)
   └─ Command already hashed correctly  ✓

5. ARA Lookup (execution/adapter/eks_adapter.go:369)
   └─ Uses correct hash  ✓
```

## Code Changes Required

### 1. PRIMARY FIX: Add hash calculation in services/execution.go

**Location:** `services/execution.go:359` (right before constructing the Run object)

**Current code (lines 319-381):**
```go
if *fields.Engine == state.EKSEngine {
    executableCmd, err := executable.GetExecutableCommand(req)
    if err != nil {
        return run, err
    }

    if (fields.Command == nil || len(*fields.Command) == 0) && (len(executableCmd) > 0) {
        fields.Command = aws.String(executableCmd)
    }
    executableID := executable.GetExecutableID()
    // ... spot/ondemand logic ...
}

if *fields.Engine == state.EKSSparkEngine {
    // ... spark setup ...
}

if fields.NodeLifecycle == nil {
    fields.NodeLifecycle = &state.SpotLifecycle
}

run = state.Run{
    RunID:          runID,
    // ...
    Command:        fields.Command,
    CommandHash:    fields.CommandHash,  // ❌ Uses wrong hash from API layer
    // ...
}
```

**New code (insert at line ~359, before `run = state.Run{...}`):**
```go
if fields.NodeLifecycle == nil {
    fields.NodeLifecycle = &state.SpotLifecycle
}

// Calculate command_hash from actual command (FIX for ARA bug)
// This ensures jobs with different commands have different hashes,
// even if they share the same description.
if fields.Command != nil && len(*fields.Command) > 0 {
    fields.CommandHash = aws.String(fmt.Sprintf("%x", md5.Sum([]byte(*fields.Command))))
}
// If command is NULL/empty, command_hash remains NULL (malformed job)
// Do NOT fall back to description - that was the bug we're fixing

run = state.Run{
    RunID:          runID,
    // ...
    Command:        fields.Command,
    CommandHash:    fields.CommandHash,  // ✓ Now has correct hash
    // ...
}
```

**Why this location:**
- Command is finalized (line 326 for EKS, or from request)
- Before `CreateRun` is called (line 653)
- Works for both EKS standard and Spark engines
- No database update needed (hash set correctly from start)

**Imports needed:**
```go
import (
    "crypto/md5"
    // ... existing imports ...
)
```

### 2. CLEANUP: Remove broken hash calculation from endpoints.go

**Locations to modify:**
- `flotilla/endpoints.go:451-453` (CreateRunV2)
- `flotilla/endpoints.go:514-516` (CreateRunV4)
- `flotilla/endpoints.go:592-594` (CreateRunByAlias)

**Current code (appears in 3 places):**
```go
if lr.CommandHash == nil && lr.Description != nil {
    lr.CommandHash = aws.String(fmt.Sprintf("%x", md5.Sum([]byte(*lr.Description))))
}
```

**Action:** **REMOVED these 3 blocks entirely** ✅ COMPLETED

**Rationale:**
- This was the source of the bug (hashing description instead of command)
- Hash will now be calculated correctly in execution service
- API clients already don't pass command_hash, so removal has no client impact
- No fallback to description - that perpetuates the bug

### 3. OPTIMIZATION: Update SQL query to use direct parameter ✅ COMPLETED

**File:** `state/pg_queries.go`
**Location:** Line 64

**Changed from:**
```sql
AND command_hash = (SELECT command_hash FROM task WHERE run_id = $2)
```

**Changed to:**
```sql
AND command_hash = $2
```

**Benefit:** Eliminates unnecessary subquery, improves performance

### 4. OPTIMIZATION: Update call site to pass command_hash ✅ COMPLETED

**File:** `execution/adapter/eks_adapter.go`
**Location:** Lines 368-422 (in `adaptiveResources` function)

**Changed from:**
```go
if !isGPUJob {
    estimatedResources, err := manager.EstimateRunResources(ctx, *executable.GetExecutableID(), run.RunID)
    if err == nil {
        // ARA found historical data...
    } else {
        // No historical data available
        _ = metrics.Increment(metrics.EngineEKSARANoHistoricalData, metricTags, 1)
    }
}
```

**Changed to:**
```go
if !isGPUJob {
    // Only attempt ARA if we have a command hash
    if run.CommandHash == nil {
        // Command hash is NULL - job has no command (malformed job definition)
        _ = metrics.Increment(metrics.EngineEKSARANullCommandHash, metricTags, 1)
        _ = a.logger.Log(
            "level", "warn",
            "message", "Skipping ARA - NULL command_hash",
            "reason", "Job has no command (malformed definition)",
            "run_id", run.RunID,
            "definition_id", *executable.GetExecutableID(),
        )
    } else {
        estimatedResources, err := manager.EstimateRunResources(ctx, *executable.GetExecutableID(), *run.CommandHash)
        if err == nil {
            // ARA found historical data...
        } else {
            // No historical data available
            _ = metrics.Increment(metrics.EngineEKSARANoHistoricalData, metricTags, 1)
        }
    }
}
```

**Changes:**
- Added NULL check for `run.CommandHash`
- Pass `*run.CommandHash` instead of `run.RunID`
- Added specific metric and logging for NULL case

**Note:** The metric `metrics.EngineEKSARANullCommandHash` may need to be added to the metrics package.

### 5. OPTIONAL: Add validation/logging

**Location:** `state/pg_state_manager.go:1168` (CreateRun, where command_hash is stored)

**Add validation before insert:**
```go
// Validate that command_hash matches command (helps catch bugs)
if r.Command != nil && r.CommandHash != nil {
    expectedHash := fmt.Sprintf("%x", md5.Sum([]byte(*r.Command)))
    if expectedHash != *r.CommandHash {
        // Log mismatch but don't fail (for observability)
        flotillaLog.Log(
            "message", "WARNING: command_hash mismatch",
            "run_id", r.RunID,
            "expected_hash", expectedHash,
            "actual_hash", *r.CommandHash,
        )
    }
}
```

## Migration Considerations

### Clean Break (Recommended)

Since current command_hash values are incorrect, the best approach is:

1. **Deploy the fix** - All new runs get correct hash
2. **Accept loss of history** - New hashes won't match old hashes
3. **Monitor ARA metrics** - Use instrumentation to verify behavior
4. **Expect initial spike** - `ara.no_historical_data` metric will increase temporarily

**Why this is OK:**
- Current ARA data is contaminated anyway
- Better to start fresh with correct data
- New instrumentation will help monitor the recovery

### Alternative: Dual-Hash Lookup (NOT IMPLEMENTED)

**Decision:** We chose the clean break approach. No dual-hash lookup was implemented.

**Reason:** The historical data is contaminated and would perpetuate the bug. Starting fresh with correct hashing is the right approach.

## Testing Plan

### Unit Tests

**Location:** `services/execution_test.go`

```go
func TestCommandHashCalculatedFromCommand(t *testing.T) {
    // Test that command_hash is MD5 of command, not description
    req := &state.DefinitionExecutionRequest{
        ExecutionRequestCommon: &state.ExecutionRequestCommon{
            Command:     aws.String("python script.py --arg value"),
            Description: aws.String("Different description"),
        },
    }

    run, err := executionService.constructBaseRunFromExecutable(ctx, definition, req)

    expectedHash := fmt.Sprintf("%x", md5.Sum([]byte("python script.py --arg value")))
    assert.Equal(t, expectedHash, *run.CommandHash)
    assert.NotEqual(t, fmt.Sprintf("%x", md5.Sum([]byte("Different description"))), *run.CommandHash)
}

func TestCommandHashWithSameDescriptionDifferentCommands(t *testing.T) {
    // Test that different commands get different hashes even with same description
    description := "Daily processing job"

    req1 := &state.DefinitionExecutionRequest{
        ExecutionRequestCommon: &state.ExecutionRequestCommon{
            Command:     aws.String("python process.py --date 2025-01-01"),
            Description: aws.String(description),
        },
    }

    req2 := &state.DefinitionExecutionRequest{
        ExecutionRequestCommon: &state.ExecutionRequestCommon{
            Command:     aws.String("python process.py --date 2025-01-02"),
            Description: aws.String(description),
        },
    }

    run1, _ := executionService.constructBaseRunFromExecutable(ctx, definition, req1)
    run2, _ := executionService.constructBaseRunFromExecutable(ctx, definition, req2)

    assert.NotEqual(t, run1.CommandHash, run2.CommandHash,
        "Different commands should have different hashes even with same description")
}
```

### Integration Tests

**Verify end-to-end:**

1. Submit two runs with:
   - Same description
   - Different commands (e.g., different dates)

2. Check database:
   ```sql
   SELECT command, command_hash, description
   FROM task
   WHERE run_id IN ('run1', 'run2');
   ```

3. Verify:
   - Different commands → different hashes ✓
   - Same description ✓
   - Hashes are MD5 of commands ✓

### Production Verification

**After deployment, monitor:**

1. **New runs have non-NULL hash:**
   ```sql
   SELECT COUNT(*)
   FROM task
   WHERE queued_at > NOW() - INTERVAL '1 hour'
     AND command_hash IS NULL
     AND command IS NOT NULL;
   ```
   Should be 0.

2. **Hash matches command:**
   ```sql
   SELECT run_id, command, command_hash,
          MD5(command) as expected_hash
   FROM task
   WHERE queued_at > NOW() - INTERVAL '1 hour'
   LIMIT 100;
   ```
   Verify `command_hash = expected_hash`.

3. **ARA metrics (from instrumentation):**
   - `ara.no_historical_data` - will spike initially (expected)
   - `ara.resource_adjustment` - should stabilize over 3-7 days
   - `ara.hit_max_memory` - should decrease for over-provisioned jobs

## Rollback Plan

If the fix causes issues:

1. **Quick rollback:** Revert the code changes and redeploy
2. **Data is safe:** Database schema unchanged, no migrations needed
3. **Monitoring:** New instrumentation continues to work regardless

## Summary of Changes Made

| File | Lines | Action | Status |
|------|-------|--------|--------|
| `services/execution.go` | 5 | **ADD** crypto/md5 import | ✅ COMPLETED |
| `services/execution.go` | 361-368 | **ADD** command_hash calculation | ✅ COMPLETED |
| `flotilla/endpoints.go` | 451-453 | **REMOVE** description-based hash | ✅ COMPLETED |
| `flotilla/endpoints.go` | 510-512 | **REMOVE** description-based hash | ✅ COMPLETED |
| `flotilla/endpoints.go` | 584-586 | **REMOVE** description-based hash | ✅ COMPLETED |
| `state/pg_queries.go` | 64 | **MODIFY** Remove subquery, use $2 directly | ✅ COMPLETED |
| `execution/adapter/eks_adapter.go` | 369-422 | **ADD** NULL check and pass *run.CommandHash | ✅ COMPLETED |
| `services/execution_test.go` | New | **ADD** unit tests (TODO) | ⏳ PENDING |

## Timeline Estimate

- Code changes: 30 minutes
- Unit tests: 1 hour
- Integration testing: 2 hours
- Deployment: Standard release process
- Monitoring period: 3-7 days for ARA to stabilize

## Success Criteria

1. ✓ All new runs have `command_hash = MD5(command)`
2. ✓ Different commands have different hashes
3. ✓ Zero NULL command_hash for new runs (except truly NULL commands)
4. ✓ ARA metrics stabilize within 7 days
5. ✓ OOM rates decrease for previously over-provisioned jobs
