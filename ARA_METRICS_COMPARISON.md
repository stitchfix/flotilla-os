# ARA Metrics Implementation Comparison

Comparing `ez/ara-metrics` (HEAD) vs `ez/ara-metrics-alt`

## Overview

Both implementations add instrumentation to track Auto Resource Adjustment (ARA) behavior to identify over-provisioning patterns, particularly the ~300GB memory limit issue. However, they differ significantly in approach, metrics design, logging strategy, and code structure.

---

## Similarities

### Shared Goals
- Track ARA resource adjustments
- Detect when jobs hit maximum resource limits (especially 350GB memory)
- Enable monitoring to identify over-provisioning patterns
- Instrument `adaptiveResources()` function
- Add structured logging for debugging

### Common Changes
- Both modify `execution/adapter/eks_adapter.go`
- Both add new metric constants to `clients/metrics/metrics.go`
- Both track default resources before ARA applies adjustments
- Both detect and report when max bounds are hit
- Both use structured key-value logging format

---

## Key Differences

### 1. **Metric Naming Convention**

**HEAD (`ez/ara-metrics`):**
- Uses hierarchical dot notation: `engine.eks.ara.*`
- Examples: `engine.eks.ara.estimation_attempted`, `engine.eks.ara.memory_increase`
- Consistent with existing codebase pattern (`engine.eks.execute`, etc.)

**Alt (`ez/ara-metrics-alt`):**
- Uses flat namespace: `ara.*`
- Examples: `ara.resource_adjustment`, `ara.memory_increase_ratio`
- Shorter, more concise names

**Winner:** HEAD - Consistent with existing naming conventions

---

### 2. **Metrics Coverage**

**HEAD (10 metrics):**
```go
// Estimation tracking
EngineEKSARAEstimationAttempted  // Counter
EngineEKSARAEstimationSucceeded  // Counter
EngineEKSARAEstimationFailed     // Counter

// Resource tracking
EngineEKSARAMaxResourceHit       // Counter (tagged with resource:memory or resource:cpu)
EngineEKSARAMemoryIncrease       // Distribution
EngineEKSARACPUIncrease          // Distribution
EngineEKSARADefaultMemory        // Distribution
EngineEKSARAARAMemory            // Distribution
EngineEKSARADefaultCPU           // Distribution
EngineEKSARAARACPU               // Distribution
```

**Alt (8 metrics):**
```go
// Core tracking
ARAResourceAdjustment            // Counter (when ARA triggers)
ARANoHistoricalData              // Counter (when no data found)

// Ratio tracking
ARAMemoryIncreaseRatio           // Histogram
ARACPUIncreaseRatio              // Histogram

// Limit detection
ARAHitMaxMemory                  // Counter
ARAHitMaxCPU                     // Counter

// Final distributions
ARAFinalMemoryMB                 // Histogram
ARAFinalCPUMillicores            // Histogram
```

**Comparison:**
- **HEAD:** More granular - separates estimation attempts from successes/failures
- **ALT:** More focused - tracks key ratios and final states
- **HEAD:** Tracks resource increases as absolute values
- **ALT:** Tracks increases as ratios (better for understanding relative growth)

**Winner:** Tie - Both approaches have merit. HEAD provides more granularity; ALT provides better insight into relative growth.

---

### 3. **Logging Strategy**

**HEAD:**
- Logging only occurs when max resource bounds are hit
- Uses stored logger instance (field on `eksAdapter`)
- Separate `emitARAMetrics()` method for structured logging
- Logs once per max-bound-hit event
- Fields: `run_id`, `definition_id`, `executable_id`, `command`, default/final resources, max hit flags

**ALT:**
- **Multiple logging points:**
  1. When ARA triggers adjustments (INFO level)
  2. When max limits hit (WARN level)
  3. In `state/pg_state_manager.go` for historical data lookups (success/no data/error)
- Uses inline `flotillaLog.NewLogger(nil, nil)` - creates new logger instances
- More verbose logging at each step
- Detailed structured fields including ratios, overage amounts, cluster name
- Separate logs for historical data lookup success/failure

**Winner:** ALT - More comprehensive logging provides better debugging capability

---

### 4. **Logger Management**

**HEAD:**
```go
type eksAdapter struct {
    logger flotillaLog.Logger  // Stored as field
}

func NewEKSAdapter(logger flotillaLog.Logger) (EKSAdapter, error) {
    adapter := eksAdapter{logger: logger}
    return &adapter, nil
}

// Usage in HEAD
if a.logger == nil {
    return
}
a.logger.Log(logFields...)
```

**ALT:**
```go
// No logger field stored
// Creates new logger instances inline
_ = flotillaLog.NewLogger(nil, nil).Log(...)
```

**Comparison:**
- **HEAD:** Dependency injection pattern - logger passed via constructor, stored as field
- **ALT:** Creates new logger instances inline (less efficient, harder to test)
- **HEAD:** Requires updating `eks_engine.go` to pass logger (which it does)
- **ALT:** No changes needed to constructor/initialization

**Winner:** HEAD - Better design pattern (dependency injection), more testable

---

### 5. **Tagging Strategy**

**HEAD:**
- No tags used on metrics (empty `[]string{}`)
- Simpler, avoids cardinality concerns
- May limit filtering/grouping capabilities in DataDog

**ALT:**
- Uses cluster tags: `[]string{fmt.Sprintf("cluster:%s", run.ClusterName)}`
- Explicitly documented as "low-cardinality tags to avoid excessive volume"
- Enables per-cluster analysis

**Winner:** ALT - Tags enable better filtering and per-cluster analysis

---

### 6. **Metric Types**

**HEAD:**
- Uses `Distribution()` for all numeric metrics
- Uses `Increment()` for counters

**ALT:**
- Uses `Histogram()` for ratios and final values
- Uses `Increment()` for counters

**Comparison:**
- DataDog treats Histogram and Distribution similarly for most use cases
- Both approaches are valid

**Winner:** Tie - No significant difference

---

### 7. **Code Structure**

**HEAD:**
- Cleaner separation: detects max hits after bounds checking
- Uses helper method `emitARAMetrics()` to centralize logging logic
- More modular: logging logic separate from bounds checking

**ALT:**
- Metrics/logging embedded directly in `checkResourceBounds()` 
- Requires passing additional parameters (`run`, `executable`, `defaultCPU`, etc.) to `checkResourceBounds()`
- More invasive changes to function signatures
- Inline logging at multiple points

**Winner:** HEAD - Better code organization, less invasive changes

---

### 8. **State Manager Instrumentation**

**HEAD:**
- No changes to `state/pg_state_manager.go`
- Only instruments the adapter layer

**ALT:**
- **Adds instrumentation to `state/pg_state_manager.go`**
- Logs when historical data is found/not found/errors occur
- Provides visibility into the data lookup layer
- Helps debug issues with historical data queries

**Winner:** ALT - Provides better end-to-end visibility

---

### 9. **Test Coverage**

**HEAD:**
- **Comprehensive test suite** (524 lines in `eks_adapter_test.go`)
- Tests multiple scenarios:
  - ARA enabled with successful estimation
  - GPU jobs (skip ARA)
  - Estimation failures
  - Max resource bounds hitting
  - ARA disabled
  - Logger nil handling
- Mock implementations for logger and state manager

**ALT:**
- No test files included

**Winner:** HEAD - Significantly better test coverage

---

### 10. **Documentation**

**HEAD:**
- Commit message describes changes
- No separate documentation file

**ALT:**
- **Comprehensive 317-line documentation** (`docs/ara-instrumentation.md`)
- Includes:
  - Overview of ARA algorithm
  - Historical context of ARA implementation
  - Detailed explanation of metrics
  - DataDog query examples
  - Alert recommendations
  - Investigation workflow
  - Future improvement suggestions
- Extremely helpful for operators and future developers

**Winner:** ALT - Outstanding documentation

---

### 11. **Detection Logic**

**HEAD:**
```go
// After bounds checking
cpuRequestBeforeBounds := cpuRequest
memRequestBeforeBounds := memRequest
cpuRequest, memRequest = a.checkResourceBounds(...)

// Then detect hits
if memRequestBeforeBounds > maxMem {
    maxMemHit = true
    // emit metrics/logs
}
```

**ALT:**
```go
// Inside checkResourceBounds()
if mem > maxMem {
    // Emit metrics and logs immediately
    _ = metrics.Increment(metrics.ARAHitMaxMemory, ...)
    // ... logging ...
    mem = maxMem
}
```

**Comparison:**
- **HEAD:** Two-step process - check bounds, then detect if hit
- **ALT:** Single-step - detect and log during bounds checking
- **ALT:** More straightforward, less code

**Winner:** ALT - Simpler, more direct approach

---

### 12. **ARA Trigger Detection**

**HEAD:**
- No explicit "ARA triggered" detection
- Only tracks estimation attempts/success/failure
- Doesn't distinguish between "ARA found same values" vs "ARA actually changed resources"

**ALT:**
```go
araTriggered := (estimatedResources.Cpu != cpuRequest || 
                estimatedResources.Memory != memRequest)
```
- Explicitly detects when ARA actually changes resources
- Only logs/increments metrics when resources actually change
- More precise tracking

**Winner:** ALT - More accurate tracking of actual ARA adjustments

---

## Best-of-Breed Recommendation

**The ideal solution would combine:**

### From HEAD:
1. ? **Metric naming convention** - Use `engine.eks.ara.*` pattern
2. ? **Logger as dependency** - Store logger as field, inject via constructor
3. ? **Code organization** - Separate `emitARAMetrics()` method
4. ? **Test coverage** - Include comprehensive test suite
5. ? **Granular metrics** - Track estimation attempts/success/failure separately

### From ALT:
1. ? **Logging strategy** - Log when ARA triggers AND when limits hit
2. ? **State manager instrumentation** - Add logging in `pg_state_manager.go`
3. ? **Documentation** - Include comprehensive docs file
4. ? **Tagging** - Use cluster tags for filtering
5. ? **Ratio metrics** - Track ratios instead of/in addition to absolute increases
6. ? **ARA trigger detection** - Explicitly detect when ARA actually changes resources

### Hybrid Approach:
```go
// Metrics (combine both approaches)
- engine.eks.ara.estimation_attempted     // Counter
- engine.eks.ara.estimation_succeeded     // Counter  
- engine.eks.ara.estimation_failed         // Counter
- engine.eks.ara.resource_adjustment       // Counter (only when changed)
- engine.eks.ara.memory_increase_ratio     // Histogram (ALT's approach)
- engine.eks.ara.cpu_increase_ratio        // Histogram
- engine.eks.ara.hit_max_memory            // Counter
- engine.eks.ara.hit_max_cpu               // Counter
- engine.eks.ara.final_memory_mb           // Histogram
- engine.eks.ara.final_cpu_millicores      // Histogram

// Logging (ALT's comprehensive approach)
- Log when ARA triggers (INFO)
- Log when limits hit (WARN)
- Log in state manager for historical lookups

// Code structure (HEAD's approach)
- Store logger as field
- Separate emitARAMetrics() method
- Use cluster tags on metrics

// Documentation
- Include ALT's comprehensive docs

// Tests
- Include HEAD's comprehensive test suite
```

---

## Verdict

**Best Overall:** Neither solution is perfect alone. **ALT is closer to production-ready** due to:
- Comprehensive documentation
- Better logging strategy
- End-to-end instrumentation
- Ratio-based metrics (easier to understand)

**But HEAD has better engineering practices:**
- Dependency injection
- Test coverage
- Code organization

**Recommendation:** Start with ALT as the base, then incorporate HEAD's improvements:
1. Store logger as field (HEAD)
2. Add test suite (HEAD)
3. Optionally adjust metric names to match HEAD's convention
4. Keep ALT's logging and documentation

This hybrid would be the best-of-breed solution.
