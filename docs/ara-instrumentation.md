# ARA Instrumentation Guide

## Overview

This document describes the instrumentation added to measure Auto Resource Adjustment (ARA) behavior in Flotilla. The goal is to understand how often ARA causes resource growth and identify potential over-provisioning, particularly when jobs repeatedly hit maximum resource limits (~300GB memory).

## Background: How ARA Works

### What is ARA?

Auto Resource Adjustment (ARA) is a feature that automatically adjusts CPU and memory resources for Kubernetes jobs based on historical usage data from previous runs that experienced Out-Of-Memory (OOM) failures.

### Historical Context

1. **Initial Implementation (~2020):** ARA was introduced as an optional feature controlled by the `adaptive_resource_allocation` field on task definitions
2. **Global Override (Jan 2020):** Added `eks.ara_enabled` config parameter for global control
3. **Always Enabled (Mar 2022, commit 6eb44086):** ARA was hardcoded to always be enabled in `execution/engine/eks_engine.go:70`
   - All jobs now run with ARA regardless of configuration
   - The toggle was removed

### ARA Algorithm

**Location:** `execution/adapter/eks_adapter.go:adaptiveResources()`

**Process:**
1. Job starts with default resources from task definition
2. ARA queries historical data via `EstimateRunResources()` in `state/pg_state_manager.go`
3. SQL query (`state/pg_queries.go:TaskResourcesSelectCommandSQL`) looks for:
   - Jobs from the same definition with matching command hash
   - That OOM'd (exit_code=137 or exit_reason='OOMKilled')
   - Within the last 3 days
   - Up to 30 most recent runs
4. Calculates P99 (99th percentile) of resource usage and applies multipliers:
   - **Memory:** P99 max memory × **1.75**
   - **CPU:** P99 max CPU × **1.25**
5. Ensures request ≤ limit, applies bounds checking

**Resource Limits:**
- Min CPU: 256 millicores
- Max CPU: 60,000 millicores (94,000 for GPU jobs)
- Min Memory: 512 MB
- Max Memory: **350,000 MB** (~341 GB) for standard jobs (376,000 MB for GPU)

### Why Jobs Grow to ~300GB

The 1.75x multiplier compounds with each OOM:
1. Job runs with 10GB → OOMs
2. Next run: 10GB × 1.75 = 17.5GB → OOMs
3. Next run: 17.5GB × 1.75 = 30.6GB → OOMs
4. Pattern continues: 30.6GB → 53.5GB → 93.6GB → 163GB → 285GB → **350GB limit hit**

Each OOM triggers exponential growth until the maximum limit is reached.

## Instrumentation Added

### Metrics (DataDog)

All metrics use low-cardinality tags (`cluster` only) to avoid excessive volume.

#### Counters

| Metric | Description | When to Alert |
|--------|-------------|---------------|
| `ara.resource_adjustment` | Incremented when ARA triggers resource changes | Track frequency of ARA usage |
| `ara.no_historical_data` | Jobs with no ARA historical data (using defaults) | Monitor new job patterns |
| `ara.hit_max_memory` | **Jobs hitting 350GB memory limit** | **Critical: indicates over-provisioning** |
| `ara.hit_max_cpu` | Jobs hitting CPU limit | Monitor CPU exhaustion |

#### Histograms/Distributions

| Metric | Description | Use Case |
|--------|-------------|----------|
| `ara.memory_increase_ratio` | Ratio of adjusted/original memory | Understand typical growth (e.g., 1.75 = 75% increase) |
| `ara.cpu_increase_ratio` | Ratio of adjusted/original CPU | Understand CPU scaling patterns |
| `ara.final_memory_mb` | Final memory allocated (after ARA + bounds) | Distribution of actual allocations |
| `ara.final_cpu_millicores` | Final CPU allocated (after ARA + bounds) | Distribution of CPU allocations |

### Structured Logging

All logs use key-value pairs compatible with standard log aggregation tools.

#### ARA Adjustment Logs (Info Level)

**Location:** `execution/adapter/eks_adapter.go:adaptiveResources()`

**When:** ARA triggers resource changes based on historical data

**Fields:**
```
message: "ARA adjusted resources"
definition_id: <definition UUID>
run_id: <run UUID>
cluster: <cluster name>
default_cpu_millicores: <original CPU>
adjusted_cpu_millicores: <ARA-adjusted CPU>
cpu_ratio: <adjusted/original>
default_memory_mb: <original memory>
adjusted_memory_mb: <ARA-adjusted memory>
memory_ratio: <adjusted/original>
```

#### Limit Hit Logs (Warning Level) - CRITICAL

**Location:** `execution/adapter/eks_adapter.go:checkResourceBounds()`

**When:** Jobs hit maximum memory or CPU limits

**Memory Limit Example:**
```
level: "warn"
message: "ARA memory allocation hit maximum limit - potential over-provisioning"
definition_id: <definition UUID>
run_id: <run UUID>
cluster: <cluster name>
default_memory_mb: <original memory from definition>
requested_memory_mb: <what ARA calculated>
final_memory_mb: 350000
memory_overage_mb: <how much over limit was requested>
ara_triggered: true/false
```

**CPU Limit Example:**
```
level: "warn"
message: "ARA CPU allocation hit maximum limit"
definition_id: <definition UUID>
run_id: <run UUID>
cluster: <cluster name>
default_cpu_millicores: <original CPU>
requested_cpu_millicores: <what ARA calculated>
final_cpu_millicores: 60000
cpu_overage_millicores: <how much over limit>
ara_triggered: true/false
```

#### Historical Data Lookup Logs

**Location:** `state/pg_state_manager.go:EstimateRunResources()`

**Success:**
```
message: "ARA: Historical resource data found"
definition_id: <definition UUID>
command_hash: <MD5 of command>
estimated_memory_mb: <calculated value>
estimated_cpu_millicores: <calculated value>
```

**No Data (Expected):**
```
message: "ARA: No historical resource data found"
definition_id: <definition UUID>
command_hash: <MD5 of command>
```

**Error:**
```
level: "error"
message: "ARA: Error querying historical resource data"
definition_id: <definition UUID>
command_hash: <MD5 of command>
error: <error message>
```

## Using the Instrumentation

### Key Questions You Can Answer

#### 1. How often does ARA trigger resource increases?

**DataDog Query:**
```
sum:ara.resource_adjustment{*}.as_count()
```

Compare to total job submissions to get percentage.

#### 2. How many jobs are hitting the ~300GB limit? ⭐ MOST IMPORTANT

**DataDog Query:**
```
sum:ara.hit_max_memory{*}.as_count()
```

**Log Query (to identify specific jobs):**
```
message:"ARA memory allocation hit maximum limit - potential over-provisioning"
```

Group by `definition_id` to find which task definitions are affected.

#### 3. What's the typical resource growth ratio?

**DataDog Query:**
```
avg:ara.memory_increase_ratio{*}
p50:ara.memory_increase_ratio{*}
p90:ara.memory_increase_ratio{*}
p99:ara.memory_increase_ratio{*}
```

A ratio of 1.75 means 75% increase, 3.0 means 200% increase, etc.

#### 4. Distribution of final memory allocations

**DataDog Query:**
```
avg:ara.final_memory_mb{*}
p50:ara.final_memory_mb{*}
p90:ara.final_memory_mb{*}
p95:ara.final_memory_mb{*}
p99:ara.final_memory_mb{*}
```

Shows the actual memory being allocated across all jobs.

#### 5. Which specific definitions are over-provisioning?

**Log Filter:**
```
message:"potential over-provisioning"
```

Extract `definition_id` and `memory_overage_mb` to prioritize which jobs need attention.

### Recommended Alerts

#### Critical: Excessive Memory Limit Hits

**Metric:** `ara.hit_max_memory`

**Threshold:** Alert if > 10 hits per hour

**Why:** Indicates jobs are repeatedly hitting the 350GB limit, suggesting either:
- Jobs genuinely need more than 350GB (need larger instances)
- ARA is over-provisioning (need to adjust multipliers)

#### High CPU Limit Hits

**Metric:** `ara.hit_max_cpu`

**Threshold:** Alert if > 5 hits per hour

**Why:** CPU exhaustion can cause job failures or slowdowns.

### Investigation Workflow

When you see high `ara.hit_max_memory` counts:

1. **Identify affected definitions:**
   ```
   Log filter: message:"potential over-provisioning"
   Group by: definition_id
   Sort by: count
   ```

2. **Analyze a specific definition:**
   ```
   Filter: definition_id:"<uuid>" AND message:"ARA"
   Look for patterns:
   - How much overage? (memory_overage_mb)
   - What was the original default? (default_memory_mb)
   - Growth ratio? (memory_ratio)
   ```

3. **Check job success rate:**
   - Are these jobs actually succeeding despite hitting the limit?
   - Or are they still OOM'ing even at max resources?

4. **Decide on action:**
   - If jobs succeed at max limit: Likely over-provisioning, consider:
     - Reducing ARA multiplier from 1.75x to 1.5x or 1.25x
     - Making ARA configurable per-definition again
     - Setting reasonable max limits per definition type
   - If jobs fail even at max limit: Jobs legitimately need more resources:
     - Increase max memory limit
     - Use larger instance types
     - Optimize job code to use less memory

## Code Locations

### Metrics Constants
- File: `clients/metrics/metrics.go`
- Lines: 51-59

### Main Instrumentation
- File: `execution/adapter/eks_adapter.go`
- Functions: `adaptiveResources()`, `checkResourceBounds()`
- Lines: 352-492

### Historical Data Logging
- File: `state/pg_state_manager.go`
- Function: `EstimateRunResources()`
- Lines: 118-162

### ARA SQL Query
- File: `state/pg_queries.go`
- Constant: `TaskResourcesSelectCommandSQL`
- Lines: 54-66

## Future Improvements

Based on instrumentation data, consider:

1. **Make ARA configurable again** - Restore per-definition or global toggles for A/B testing
2. **Adjust multipliers** - If 1.75x is too aggressive, reduce to 1.5x or 1.25x
3. **Per-definition limits** - Set different max memory based on job type
4. **Graduated multipliers** - Use smaller multipliers as resources grow (e.g., 1.75x up to 50GB, then 1.25x)
5. **Decay historical data** - Weight recent OOMs more than old ones
6. **Track actual usage vs allocation** - Compare requested resources to what jobs actually use

## Related Documentation

- ARA Feature Documentation: `docs/ara.md`
- State Models: `state/models.go`
- Resource Queries: `state/pg_queries.go`
- Main CLAUDE.md: Project overview and development guide
