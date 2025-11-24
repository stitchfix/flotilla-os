# ARA command_hash Bug Report

## Executive Summary

The Auto Resource Adjustment (ARA) feature has a **critical bug** where `command_hash` is calculated from the **description** field instead of the actual command, causing:

1. **21,357 runs** (23 definitions) with NULL command_hash receive **no ARA benefit**
2. **Hundreds of thousands of runs** share ARA data across **completely different commands** that happen to have the same description

This means jobs can inherit resource allocations from unrelated workloads, leading to incorrect over- or under-provisioning.

## The Bug

### How command_hash Should Work

`command_hash` is used by ARA to match similar jobs and apply historical OOM data. The intent is to group jobs running the **same command**.

### How It Actually Works

**Location:** `flotilla/endpoints.go:451-453, 514-516, 592-593`

```go
if lr.CommandHash == nil && lr.Description != nil {
    lr.CommandHash = aws.String(fmt.Sprintf("%x", md5.Sum([]byte(*lr.Description))))
}
```

**Problems:**
1. Hash is MD5 of **Description**, not Command
2. If Description is NULL, command_hash stays NULL
3. NULL command_hash never matches anything in SQL (`command_hash = NULL` always FALSE)

## Impact by the Numbers

### Bug #1: NULL command_hash (No ARA)

```sql
SELECT COUNT(*) as total_runs, COUNT(DISTINCT definition_id) as definitions_affected
FROM task WHERE command_hash IS NULL;
```

**Result:**
- **21,357 runs** have NULL command_hash
- **23 definitions** affected
- These jobs **never benefit from ARA** despite it being enabled

**Example:** Definition `sf-base_python-3_11-...` has **55 different commands**, all with NULL command_hash, none sharing ARA data.

### Bug #2: Description-based Hash (Incorrect ARA Sharing)

```sql
-- Find command_hash values with multiple different commands
SELECT definition_id, command_hash,
       COUNT(DISTINCT command) as distinct_commands,
       COUNT(*) FILTER (WHERE exit_code = 137) as oom_count,
       COUNT(*) as total_runs
FROM task
WHERE command_hash IS NOT NULL AND command IS NOT NULL
GROUP BY definition_id, command_hash
HAVING COUNT(DISTINCT command) > 1
ORDER BY oom_count DESC, total_runs DESC
LIMIT 1;
```

**Result:**
- **Worst case:** `command_hash = 407f6885beaec163a742e8c3c8a50d3e`
  - **176 different commands** share the same hash
  - **115 OOMs** across these different commands
  - **287 total runs**
  - All share description: "Calibrate Psale Prod / Calibrate Psale"

**Other severe cases:**
- `a0798e54ea76fb8dc1e743fe37f761e0`: 2 commands, **87,142 runs** affected
- `1eeb37af6d7e0e4bb2a73a0f61ac7a79`: 2 commands, **52,844 runs** affected
- `123fad187daf3847583761f5495e3ce8`: 2 commands, **39,181 runs** affected

## Concrete Example: The Smoking Gun

### Timeline

**November 22-24, 2025** - Daily data processing job with description "Calibrate Psale Prod / Calibrate Psale"

#### OOMs in 3-Day Window (Contributing to ARA):

| Date | Run ID | Memory | Command Differs By |
|------|--------|--------|-------------------|
| Nov 22 | `eks-c662-2a1e-44f7...` | 1024 MB | `--as_of 20251121` |
| Nov 22 | `eks-a9fd-92f6-4fe1...` | 1792 MB | `--as_of 20251121` |
| Nov 23 | `eks-055c-c578-4951...` | 1024 MB | `--as_of 20251122` |

**ARA Calculation:**
- P99([1024, 1792, 1024]) = 1792 MB
- 1792 MB × 1.75 = **3136 MB**

#### Next Day Run (Inherits OOM Data):

| Date | Run ID | Memory | Command Differs By | Exit Code |
|------|--------|--------|-------------------|-----------|
| Nov 24 | `eks-0d33-a443-43b9...` | **3136 MB** | `--as_of 20251123` | 0 (Success) |

### The Commands Are Different!

**Nov 23 OOM Command:**
```bash
python3 /dsn-algo-adhoc/damien/projects/fy25q4_psale_calibration/calibrate.py --as_of 20251122
```

**Nov 24 Command (Got ARA from above):**
```bash
python3 /dsn-algo-adhoc/damien/projects/fy25q4_psale_calibration/calibrate.py --as_of 20251123
```

**Only difference:** The date parameter (`20251122` vs `20251123`)

**Why this matters:** These are daily data processing jobs. Each date's data could have completely different characteristics and memory requirements, but they share ARA data because they have the same description.

### Verification

The exact ARA query for the Nov 24 run returns:

```sql
SELECT cast((percentile_disc(0.99) within GROUP (ORDER BY A.max_memory_used)) * 1.75 as int) as memory
FROM (SELECT memory as max_memory_used FROM TASK
      WHERE queued_at >= '2025-11-21 15:10:01' AND queued_at < '2025-11-24 15:10:01'
        AND (exit_code = 137 or exit_reason = 'OOMKilled')
        AND definition_id = 'sf-base_python-3_9-59ab1a32-cdda-4eb8-5824-49d17d96b1fd'
        AND command_hash = '407f6885beaec163a742e8c3c8a50d3e'
      LIMIT 30) A;
```

**Result:** 3136 MB ← **Exactly what the Nov 24 run received**

## Why This Causes Over-Provisioning

1. **Cross-contamination:** Jobs inherit OOM data from unrelated workloads
2. **Compounding growth:** The 1.75x multiplier compounds across different jobs
3. **Never stabilizes:** Each day's job can trigger growth for the next day's job
4. **Reaches maximum:** Eventually hits the 350GB limit, explaining the "jobs growing to 300GB" issue

## Scale of the Problem

### Definitions with Most Cross-Command OOMs

```sql
SELECT definition_id, command_hash,
       COUNT(DISTINCT command) as distinct_commands,
       COUNT(*) FILTER (WHERE exit_code = 137 OR exit_reason = 'OOMKilled') as oom_count,
       COUNT(*) as total_runs
FROM task
WHERE command_hash IS NOT NULL AND engine = 'eks' AND command IS NOT NULL
GROUP BY definition_id, command_hash
HAVING COUNT(DISTINCT command) > 1
   AND COUNT(*) FILTER (WHERE exit_code = 137 OR exit_reason = 'OOMKilled') > 0
ORDER BY oom_count DESC
LIMIT 10;
```

| Rank | command_hash | Distinct Commands | OOMs | Total Runs |
|------|--------------|-------------------|------|------------|
| 1 | `407f6885beaec163...` | 176 | 115 | 287 |
| 2 | `a5bdb8f3302110219...` | 164 | 87 | 304 |
| 3 | `2344c10bd7229...` | 184 | 83 | 564 |
| 4 | `7803d8faa568610...` | 97 | 82 | 261 |
| 5 | `90ceb0cabff4958...` | 135 | 82 | 230 |

All from the same definition: `sf-base_python-3_9-59ab1a32-cdda-4eb8-5824-49d17d96b1fd`

### Definitions with NULL command_hash (No ARA)

```sql
SELECT definition_id,
       COUNT(DISTINCT command) as distinct_commands,
       COUNT(*) as total_runs
FROM task
WHERE command_hash IS NULL AND command IS NOT NULL
GROUP BY definition_id
HAVING COUNT(DISTINCT command) > 1
ORDER BY total_runs DESC
LIMIT 5;
```

| Definition ID | Distinct Commands | Total Runs |
|---------------|-------------------|------------|
| `sf-base_python-3_11-7449eda4-b8b3-4146-77c5-a47f8caac81b` | 55 | 91 |
| `sf-base_python-3_9-59ab1a32-cdda-4eb8-5824-49d17d96b1fd` | 40 | 49 |
| `data-platform-d834291f-d984-408e-5da4-8646f7e2f5b7` | 4 | 31 |
| `platform-8a651dbe-1794-485b-6ba4-ba58b4a10212` | 5 | 21 |
| `sf-base_pytorch2-24__5-py3-ceef4c9e-6ebc-41e5-6cef-a334aed6e829` | 6 | 17 |

## Root Cause Analysis

### Design Intent vs Implementation

**Intended behavior:**
- Jobs running the **same command** should share ARA data
- Different commands should have separate ARA histories

**Actual behavior:**
- Jobs with the **same description** share ARA data
- Command can be completely different

### Why Description Was Used

Looking at the code flow:

1. API receives execution request with optional `description` field
2. If `command_hash` not provided by client, generate from description
3. **Problem:** Command isn't available yet at this point in the flow
4. Command is constructed later during job submission

**The Disconnect:**
- `command_hash` is set in `flotilla/endpoints.go` (API layer)
- Actual `command` is finalized in `execution/adapter/eks_adapter.go` (execution layer)
- By the time the command is known, the hash is already set

## The Fix

### Recommended Solution

Calculate `command_hash` from the **actual command** that will run:

**Location to fix:** Where the Run object gets its final command, likely in the execution service before calling `EstimateRunResources()`.

**Pseudocode:**
```go
// After command is finalized, before ARA lookup
if run.Command != nil && len(*run.Command) > 0 {
    run.CommandHash = aws.String(fmt.Sprintf("%x", md5.Sum([]byte(*run.Command))))
} else {
    // Fallback: use description if no command (shouldn't happen for EKS jobs)
    if run.Description != nil {
        run.CommandHash = aws.String(fmt.Sprintf("%x", md5.Sum([]byte(*run.Description))))
    }
}
```

### Migration Strategy

**Challenge:** Changing command_hash breaks ARA history

**Options:**

1. **Clean break (Recommended):**
   - Fix the hash calculation
   - Accept that ARA starts fresh for all jobs
   - Monitor via new instrumentation to ensure it works correctly

2. **Dual-hash lookup:**
   - Try command-based hash first
   - Fall back to description-based hash for historical data
   - Gradually phase out old hashes

3. **Per-definition rollout:**
   - Fix hash for definitions most affected by the bug
   - Leave others on old behavior temporarily
   - Migrate gradually

### Testing Plan

1. **Verify hash calculation:**
   - Unit tests ensuring hash comes from command, not description
   - Integration tests with various command/description combinations

2. **Verify ARA still works:**
   - Test that identical commands share ARA data
   - Test that different commands DON'T share data

3. **Monitor after deployment:**
   - Use new `ara.*` metrics to track behavior
   - Watch for unexpected resource changes
   - Check logs for `ara.no_historical_data` - should increase initially

## Impact on Current Investigation

This bug significantly impacts the "jobs growing to 300GB" investigation:

1. **Over-provisioning is worse than thought:**
   - Jobs inherit OOMs from unrelated workloads
   - The 1.75x multiplier compounds across different jobs
   - Growth isn't just from retrying the same job, but cross-contamination

2. **Instrumentation still valuable:**
   - The new ARA metrics will help measure the bug's impact
   - After fixing, metrics will show if ARA works correctly

3. **Fix priority:**
   - This bug should be fixed **before** tuning ARA multipliers
   - Otherwise, you're tuning a broken system

## Queries for Further Investigation

### Find your most affected definitions

```sql
-- Definitions with most OOM cross-contamination
SELECT
    definition_id,
    command_hash,
    COUNT(DISTINCT MD5(command)) as distinct_commands,
    COUNT(*) FILTER (WHERE exit_code = 137 OR exit_reason = 'OOMKilled') as oom_count,
    COUNT(*) as total_runs,
    MAX(memory) as max_memory_allocated
FROM task
WHERE command_hash IS NOT NULL
  AND engine = 'eks'
  AND command IS NOT NULL
  AND queued_at >= CURRENT_TIMESTAMP - INTERVAL '30 days'
GROUP BY definition_id, command_hash
HAVING COUNT(DISTINCT MD5(command)) > 1
   AND COUNT(*) FILTER (WHERE exit_code = 137 OR exit_reason = 'OOMKilled') > 0
ORDER BY oom_count * distinct_commands DESC
LIMIT 20;
```

### Find jobs hitting memory limits with cross-command contamination

```sql
-- Jobs at max memory (350GB) that share command_hash with different commands
SELECT DISTINCT t1.definition_id, t1.command_hash
FROM task t1
JOIN task t2 ON t1.definition_id = t2.definition_id
            AND t1.command_hash = t2.command_hash
            AND MD5(t1.command) != MD5(t2.command)
WHERE t1.memory >= 300000  -- Close to or at max
  AND t1.queued_at >= CURRENT_TIMESTAMP - INTERVAL '7 days'
GROUP BY t1.definition_id, t1.command_hash
HAVING COUNT(DISTINCT MD5(t1.command)) > 1;
```

## Recommendations

1. **Immediate:**
   - Review the examples in this report with the team
   - Decide on fix approach (clean break vs dual-hash)
   - Prioritize this fix before tuning ARA parameters

2. **Short-term:**
   - Implement command-based hash calculation
   - Deploy with new instrumentation
   - Monitor via `ara.*` metrics

3. **Long-term:**
   - Consider whether description should exist separately from command
   - Review if ARA should use command hash at all, or something more semantic
   - Add validation to prevent command_hash from being NULL

## Related Files

- **Bug location:** `flotilla/endpoints.go:451-453, 514-516, 592-593`
- **ARA query:** `state/pg_queries.go:54-66` (TaskResourcesSelectCommandSQL)
- **ARA lookup:** `state/pg_state_manager.go:118-162` (EstimateRunResources)
- **Resource adjustment:** `execution/adapter/eks_adapter.go:352-421` (adaptiveResources)
- **New instrumentation:** `docs/ara-instrumentation.md`

## Database Evidence

All evidence in this report is from production database queries run on 2025-11-24.

Key run IDs for reproduction:
- OOM: `eks-055c-c578-4951-75d8-3f5a0bb15b37` (Nov 23, 1024 MB, OOM)
- Inherited: `eks-0d33-a443-43b9-45f9-04b780868880` (Nov 24, 3136 MB, Success)
- Command hash: `407f6885beaec163a742e8c3c8a50d3e`
- Definition: `sf-base_python-3_9-59ab1a32-cdda-4eb8-5824-49d17d96b1fd`
