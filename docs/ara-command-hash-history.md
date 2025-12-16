# History of command_hash Implementation

## Timeline of Changes

### January 17, 2020 - Original Design (Commit a5d7e0f)
**Author:** Ujjwal Sarin
**PR:** #269
**Title:** "Adding command hash to task"

**What was added:**
1. `command_hash` column added to `task` table
2. Changed ARA query from matching exact `command` text to `command_hash`
3. **Database automatically calculated hash:** `MD5($17)` where `$17` is the command parameter

**Original CreateRun SQL:**
```sql
INSERT INTO task (
  ..., command, ..., command_hash
) VALUES (
  ..., $17, ..., MD5($17)
);
```

**Original UpdateRun SQL:**
```sql
UPDATE task SET
  command = $17, ..., command_hash = MD5($17)
WHERE run_id = $1;
```

**Intent:** Hash was calculated FROM THE COMMAND to group similar jobs for ARA resource estimation.

**Original Query Change:**
```sql
-- BEFORE: Match exact command text
WHERE command = (SELECT command FROM TASK WHERE run_id = $2)

-- AFTER: Match command hash
WHERE command_hash = (SELECT command_hash FROM task WHERE run_id = $2)
```

### January 22, 2020 - Removed Auto-Hashing from UpdateRun (Commit fbe8409)
**Author:** Ujjwal Sarin
**Title:** "removing adding command_hash on updates"

**What changed:**
- Removed `command_hash = MD5($17)` from UpdateRun SQL
- Left CreateRun unchanged (still had MD5 calculation)

**Why this matters:** This suggests the design started shifting toward setting command_hash earlier in the flow, not in the database.

### December 31, 2021 - API Layer Auto-Generation from Description (Commit 7802cfe)
**Author:** Ujjwal Sarin
**Commit message:** "encode lr"

**What was added:**
```go
// In flotilla/endpoints.go - CreateRunV2, CreateRunV4, CreateRunByAlias
if lr.CommandHash == nil && lr.Description != nil {
    lr.CommandHash = aws.String(hex.EncodeToString([]byte(*lr.Description)))
}
```

**THE BUG INTRODUCED:** Changed from hashing the command to hashing the description.

**Why description was used:** At the API layer (endpoints.go), the final command isn't constructed yet. The command gets finalized later during job submission in the execution layer.

**Context:** This commit was for Spark executor estimation feature (see below).

### December 31, 2021 - Same Day: Changed to MD5 (Commit 7e84338)
**Author:** Ujjwal Sarin
**Title:** "adding support for predicting executor"

**What changed:**
```go
// Changed from hex encoding to MD5 (same day, 2 hours later)
if lr.CommandHash == nil && lr.Description != nil {
    lr.CommandHash = aws.String(fmt.Sprintf("%x", md5.Sum([]byte(*lr.Description))))
}
```

**What was added:** Spark executor count estimation using command_hash:
```go
// execution/engine/emr_engine.go
func (emr *EMRExecutionEngine) estimateExecutorCount(run state.Run, manager state.Manager) *int64 {
    if run.Engine != nil && *run.Engine == state.EKSSparkEngine {
        count, err := manager.EstimateExecutorCount(run.DefinitionID, *run.CommandHash)
        if err == nil {
            return aws.Int64(count)
        }
    }
    return aws.Int64(100)
}
```

**New Query Added:**
```sql
const TaskResourcesExecutorCountSQL = `
SELECT COALESCE(cast((percentile_disc(0.99) within GROUP (ORDER BY A.executor_count)) * 1.75 as int), 100)
FROM (SELECT CASE WHEN (exit_reason like '%Exception%') THEN spark_extension->'num_executors' END
      FROM TASK
      WHERE queued_at >= CURRENT_TIMESTAMP - INTERVAL '7 days'
        AND engine = 'eks-spark'
        AND definition_id = $1
        AND command_hash = $2
        AND (exit_code != 0)
      LIMIT 30) A
`
```

**Significance:** This shows command_hash was being used for TWO features:
1. ARA memory/CPU estimation (original, Jan 2020)
2. Spark executor count estimation (new, Dec 2021)

Both rely on grouping similar jobs, but the Dec 2021 implementation broke this by hashing description instead of command.

## Current State (2025)

### API Layer (flotilla/endpoints.go)
```go
// Lines 451-453, 514-516, 592-593
if lr.CommandHash == nil && lr.Description != nil {
    lr.CommandHash = aws.String(fmt.Sprintf("%x", md5.Sum([]byte(*lr.Description))))
}
```

**Problem:** Hashes description, not command.

### Database Layer (state/pg_state_manager.go)
```go
// CreateRun - Line 1168
r.CommandHash  // Just uses whatever was passed in, no calculation
```

**Problem:** No fallback calculation. If API layer provides wrong hash, database accepts it.

### API Schema (state/models.go)
```go
// LaunchRequestV2 - Line 1235
type LaunchRequestV2 struct {
    Command     *string `json:"command,omitempty"`
    Description *string `json:"description,omitempty"`
    CommandHash *string `json:"command_hash,omitempty"`
    // ...
}
```

**Observation:** `command_hash` IS exposed as an optional API field, but:
1. Clients rarely/never pass it explicitly
2. API layer auto-generates from description as fallback
3. This means nearly all command_hash values in production are MD5(description)

## Root Cause Analysis

### The Design Disconnect

**Layer 1 - API (endpoints.go):**
- Receives execution request
- Command might not be finalized yet
- Needs to set command_hash for downstream use
- Only has description available
- **Decision:** Hash description as proxy for command

**Layer 2 - Execution (execution/adapter/eks_adapter.go):**
- Constructs final command from template + parameters
- Command is now known
- But command_hash was already set in Layer 1
- **Missing:** No code to recalculate hash from actual command

**Layer 3 - Database (state/pg_state_manager.go):**
- Just stores whatever command_hash was provided
- No validation that hash matches command
- **Assumption:** Hash was calculated correctly upstream

### Why This Wasn't Caught

1. **Description often stable:** Many jobs use the same description repeatedly
2. **Worked for simple cases:** Jobs with truly identical descriptions often have identical commands
3. **Gradual degradation:** As users started parameterizing commands (dates, configs), descriptions stayed same but commands diverged
4. **No monitoring:** Until the recent instrumentation patches, there was no visibility into ARA behavior

## Evidence from Production

### NULL command_hash
- **21,357 runs** with NULL command_hash (description also NULL)
- These runs get NO ARA benefit despite feature being enabled

### Cross-Command Contamination
- **Worst case:** 176 different commands sharing one command_hash
- **High-volume case:** 87,142 runs across 2 different commands
- **ML Training catastrophe:** 12 different training configs all sharing 350GB allocation

### The Smoking Gun
From docs/ara-command-hash-bug-report.md:

**Daily jobs differing only by date:**
```bash
# Nov 23 OOM
python3 calibrate.py --as_of 20251122

# Nov 24 (inherited ARA from above)
python3 calibrate.py --as_of 20251123
```

Both have description "Calibrate Psale Prod / Calibrate Psale"
→ Same command_hash
→ Share ARA data
→ Nov 24 job gets 3136 MB from Nov 23 OOM

**The data being processed is completely different** (different dates), but they share resource allocation history.

## The Original Intent vs Reality

### Original Intent (Jan 2020)
- Jobs running the **same command** share ARA data
- Different commands have separate ARA histories
- Performance optimization: hash instead of full text comparison

### Current Reality (Dec 2021 - Present)
- Jobs with the **same description** share ARA data
- Commands can be completely different
- Leads to incorrect resource allocation

## Why Description Was Chosen

Looking at the code flow:

1. API receives execution request (`flotilla/endpoints.go`)
   - Has: description (optional), command template
   - Needs: command_hash for ARA lookup

2. Command construction happens later (`execution/adapter/eks_adapter.go`)
   - Combines template + env vars + parameters
   - Final command not available at API layer

3. Timing problem:
   - `command_hash` needed before `adaptiveResources()` call
   - `command` not finalized until during job construction
   - Description available early, command available late

**The Compromise:** Use description as a "proxy" for command.

**Why it seemed reasonable:**
- Description often correlates with command
- Better than nothing for grouping similar jobs
- Performance: avoid expensive string operations on long commands

**Why it fails:**
- Parameterized commands (dates, configs, data subsets)
- Description captures "what" but not "how"
- Catastrophic cross-contamination at scale

## Related Queries

### Original ARA Query (2020-2021)
```sql
-- Before command_hash
WHERE command = (SELECT command FROM TASK WHERE run_id = $2)
```

### Current ARA Query (2022-Present)
```sql
-- Using command_hash
WHERE command_hash = (SELECT command_hash FROM task WHERE run_id = $2)
```

**Irony:** The query change was meant to make ARA more efficient, but combined with description-based hashing, it made it incorrect.

## Conclusion

The bug wasn't a single mistake but an **architectural mismatch**:

1. **2020:** Designed command_hash to group identical commands
2. **2021:** Needed to set hash early in request flow
3. **2021:** Command not available early, description chosen as proxy
4. **2021-2025:** Production usage reveals proxy doesn't work at scale

The fix requires moving command_hash calculation to **after** command is finalized, or making command available earlier in the flow.

## References

- **Original feature:** Commit a5d7e0f (Jan 17, 2020)
- **Auto-hash removal:** Commit fbe8409 (Jan 22, 2020)
- **Bug introduction:** Commit 7802cfe (Dec 31, 2021)
- **MD5 change:** Commit 7e84338 (Dec 31, 2021)
- **ARA enablement:** Commit 4c0ffc8 (Feb 23, 2022)
- **Bug documentation:** docs/ara-command-hash-bug-report.md (Nov 25, 2025)
