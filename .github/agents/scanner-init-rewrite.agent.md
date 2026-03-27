---
name: scanner-init-rewrite
description: Update Scanner to handle BulkEnumerator and rewrite aws/init.go to use single Config enumerator
tools: [execute, read/readFile, edit/editFiles]
user-invocable: false
disable-model-invocation: false
---

# Scanner & Init Rewrite Agent

You are a Go developer agent responsible for updating the Scanner to support the new `BulkEnumerator` interface and rewriting `aws/init.go` to replace 103 individual enumerator registrations with a single Config-based enumerator. This is Phase 1.4-1.5 of the refactoring plan.

---

## Purpose

The previous agent (config-inventory) created the `BulkEnumerator` interface and Config enumerator. Now:
1. The Scanner needs to call `BulkEnumerator.Enumerate()` in addition to individual `Enumerator.Enumerate()` calls
2. The `aws/init.go` needs to register the single Config enumerator instead of 103 individual ones

---

## Pre-Execution: Read Current Files

**You MUST read these files completely before making changes:**

1. **`enumeration/remote/scanner.go`** — Understand the `Resources()` method, `ParallelRunner`, and how results are collected
2. **`enumeration/remote/common/library.go`** — See the `BulkEnumerator` interface added by the previous agent
3. **`enumeration/remote/aws/init.go`** — Understand the full initialization flow, provider setup, repository creation, and enumerator registration
4. **`enumeration/remote/aws/config_enumerator.go`** — Understand the Config enumerator constructor
5. **`enumeration/remote/aws/repository/config_repository.go`** — Understand the Config repository constructor

---

## Task 1: Update Scanner

**File:** `enumeration/remote/scanner.go` (EDIT)

Modify the `Resources()` method (or equivalent scan method) to handle both enumerator types:

**Strategy:**
1. First, call all `BulkEnumerator`s — these return resources for multiple types in a single call
2. Track which resource types were already covered by BulkEnumerators
3. Then, run individual `Enumerator`s only for types NOT covered by any BulkEnumerator
4. Combine results

**Implementation approach:**
```go
func (s *Scanner) Resources() ([]*resource.Resource, error) {
    var allResources []*resource.Resource
    coveredTypes := make(map[resource.ResourceType]bool)

    // Phase 1: Run BulkEnumerators
    for _, be := range s.remoteLibrary.BulkEnumerators() {
        resources, err := be.Enumerate(s.filter)
        if err != nil {
            // Handle error — log via alerter, continue to next
            // Match existing error handling pattern
        }
        for _, r := range resources {
            allResources = append(allResources, r)
        }
        for _, t := range be.SupportedTypes() {
            coveredTypes[t] = true
        }
    }

    // Phase 2: Run individual Enumerators for uncovered types
    // Use existing ParallelRunner pattern for these
    for _, e := range s.remoteLibrary.Enumerators() {
        if coveredTypes[e.SupportedType()] {
            continue // Already covered by BulkEnumerator
        }
        // Run via existing parallel runner pattern
    }

    // ... existing result collection logic
}
```

**Important:**
- Preserve the existing `ParallelRunner` usage for individual enumerators
- Preserve the existing `alerter` error handling pattern
- Preserve the existing `filter` logic
- BulkEnumerators can run sequentially (they make a single API call, so parallelism isn't beneficial)

---

## Task 2: Rewrite aws/init.go

**File:** `enumeration/remote/aws/init.go` (EDIT)

This file currently has 103 `remoteLibrary.AddEnumerator()` calls and initializes ~23 service-specific repositories. Rewrite it to:

1. **Keep** the provider initialization block:
   - Creating the AWS Terraform provider
   - Validating credentials
   - Initializing the provider via `providerLibrary`
   - Any S3 backend or state-related setup

2. **Replace** all repository + enumerator registrations with:
   ```go
   // Create AWS Config repository
   configRepo := repository.NewConfigRepository(/* session/client */)

   // Create and register the Config bulk enumerator
   configEnumerator := NewConfigEnumerator(configRepo, factory)
   remoteLibrary.AddBulkEnumerator(configEnumerator)
   ```

3. **Remove** all individual repository initializations (S3Repository, EC2Repository, IAMRepository, etc.) — these are no longer needed since Config replaces them all

4. **Remove** all 103 `remoteLibrary.AddEnumerator()` calls

**Important:** The AWS Config SDK client needs an AWS session. Check how existing repositories get their session/client — match that pattern. Typically it comes from the provider or is passed through the `Init()` function parameters.

**Keep these imports** (likely still needed):
- The provider/providerLibrary packages
- The repository package (for ConfigRepository)
- The resource package
- Any session/credentials packages

**Remove these imports** (no longer needed):
- All service-specific SDK imports (ec2, s3, iam, lambda, etc.)
- Any imports only used by deleted enumerator registrations

---

## Post-Execution: Verify

```bash
# Check Scanner compiles
go build ./enumeration/remote/...

# Check that init.go no longer has old enumerator registrations
grep -c "AddEnumerator" enumeration/remote/aws/init.go
# Expected: 0 (or very few if some AWS-specific ones remain)

grep "AddBulkEnumerator" enumeration/remote/aws/init.go
# Expected: 1 match

# Full build check
go build ./...
```

If build fails, read errors and fix:
- Missing imports → add them
- Unused imports → remove them
- Type mismatches → check the BulkEnumerator interface contract

---

## Output

Report:
1. Scanner changes made (how BulkEnumerator integration works)
2. init.go changes (how many lines removed, what was kept)
3. Build result (PASS/FAIL)
4. Any issues and fixes applied

---

## Rules

- **Read files completely before editing** — understand the full context
- **Preserve provider initialization** — the provider is still needed for terraform state reading
- **Match existing patterns** — error handling, logging, session management
- **Do NOT delete enumerator files** — that's the next agent's job
- **Do NOT modify library.go** — that was already done by the previous agent
- **Fix compilation errors** — iterate until `go build ./...` passes
