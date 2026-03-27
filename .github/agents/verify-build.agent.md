---
name: verify-build
description: Run comprehensive build verification — go build, go vet, go mod tidy, test suite, and check for stale non-AWS references
tools: [execute, read/readFile, edit/editFiles]
user-invocable: false
disable-model-invocation: false
---

# Verify Build Agent

You are a verification agent responsible for running comprehensive checks on the refactored driftctl codebase. All previous phases have completed — provider deletion, shared file updates, Config inventory creation, terraform plan integration, and categorizer/output work. Your job is to ensure everything compiles and passes checks.

---

## Purpose

This is the final phase of the refactoring. Run all verification checks, fix any issues found, and report the final status.

---

## Step 1: Verify File Structure

Confirm all expected files exist:

```bash
echo "=== New Config Inventory Files ==="
ls -la enumeration/remote/aws/repository/config_repository.go
ls -la enumeration/remote/aws/config_resource_mapping.go
ls -la enumeration/remote/aws/config_enumerator.go

echo "=== New Terraform Plan Files ==="
ls -la pkg/terraform/plan/runner.go
ls -la pkg/terraform/plan/parser.go

echo "=== New Analyzer Files ==="
ls -la pkg/analyser/plan_analyzer.go

echo "=== New Categorizer Files ==="
ls -la pkg/categorizer/categorizer.go
ls -la pkg/categorizer/cloudformation.go
ls -la pkg/categorizer/service_linked.go
ls -la pkg/categorizer/unsupported.go

echo "=== Deleted Provider Dirs (should not exist) ==="
ls -d enumeration/remote/azurerm 2>&1 || echo "OK: azurerm deleted"
ls -d enumeration/remote/google 2>&1 || echo "OK: google deleted"
ls -d enumeration/remote/github 2>&1 || echo "OK: github deleted"
ls -d pkg/resource/azurerm 2>&1 || echo "OK: azurerm deleted"
ls -d pkg/resource/google 2>&1 || echo "OK: google deleted"
ls -d pkg/resource/github 2>&1 || echo "OK: github deleted"
```

---

## Step 2: Check for Stale Non-AWS References

```bash
echo "=== Checking for stale non-AWS references in Go files ==="

# Check for Azure references
grep -r "azurerm" --include="*.go" pkg/ enumeration/ 2>/dev/null | grep -v "_test.go" | grep -v "vendor/" || echo "OK: No azurerm references"

# Check for Google references (exclude Go standard library google imports)
grep -r "RemoteGoogleTerraform\|google_compute\|google_cloud\|google_dns\|google_storage\|google_project\|enumeration/remote/google\|enumeration/resource/google\|pkg/resource/google" --include="*.go" pkg/ enumeration/ 2>/dev/null || echo "OK: No Google provider references"

# Check for GitHub references (exclude Go standard library github imports used for real dependencies)
grep -r "RemoteGithubTerraform\|github_repository\|github_branch\|github_team\|enumeration/remote/github\|enumeration/resource/github\|pkg/resource/github" --include="*.go" pkg/ enumeration/ 2>/dev/null || echo "OK: No GitHub provider references"

# Check for deleted middleware references
grep -r "AzurermRouteExpander\|AzurermSubnetExpander\|GoogleIAMBinding\|GoogleIAMPolicy\|GoogleComputeInstanceGroup\|GoogleLegacyBucket\|GoogleDefaultIAM" --include="*.go" pkg/ enumeration/ 2>/dev/null || echo "OK: No deleted middleware references"
```

If stale references are found, **fix them** by reading the file and removing the offending lines/imports.

---

## Step 3: Run go mod tidy

```bash
go mod tidy
```

This should automatically prune:
- ~15 Azure SDK modules (`github.com/Azure/...`)
- ~5 Google Cloud modules (`cloud.google.com/...`)
- GitHub GraphQL client (`github.com/shurcooL/githubv4`)
- Their transitive dependencies

Verify cleanup:
```bash
# These should return no results
grep -c "azure-sdk" go.mod go.sum 2>/dev/null || echo "OK: No Azure SDK"
grep -c "cloud.google.com" go.mod 2>/dev/null || echo "OK: No Google Cloud"
grep -c "githubv4" go.mod 2>/dev/null || echo "OK: No GitHub GraphQL"
```

---

## Step 4: Run go build

```bash
go build ./...
```

If build fails:
1. **Read the error messages** — identify which files and what's wrong
2. **Common fixes:**
   - Unused imports → remove the import line
   - Undefined references → check if a type/function was deleted that shouldn't have been, or find the correct replacement
   - Type mismatches → check interface contracts
   - Missing packages → check if `go mod tidy` removed something still needed
3. **Fix the errors** by editing the relevant files
4. **Re-run** `go build ./...`
5. **Repeat** until build passes (max 5 iterations)

---

## Step 5: Run go vet

```bash
go vet ./...
```

If vet reports issues:
1. Read each issue
2. Fix if straightforward (unused variables, wrong printf verbs, etc.)
3. Re-run until clean

---

## Step 6: Run Tests

```bash
# Run tests with timeout and verbose output
go test ./... -timeout 120s -count=1 2>&1 | head -200
```

Report test results:
- Total tests run
- Tests passed
- Tests failed (list the failing test names and packages)
- Tests skipped

**Note:** Some test failures are expected since we deleted test dependencies. Report them but don't consider them blocking unless they indicate compilation errors.

If there are test compilation errors (not just test logic failures), fix them:
- Remove test files that reference deleted packages/types
- Update test files that reference changed interfaces

---

## Step 7: Final Verification Summary

```bash
echo "=== Final File Counts ==="
echo "AWS enumerator files (should be few):"
find enumeration/remote/aws -name "*.go" -type f | wc -l
echo "Config files:"
ls enumeration/remote/aws/config_*.go enumeration/remote/aws/repository/config_*.go 2>/dev/null
echo "Plan files:"
ls pkg/terraform/plan/*.go 2>/dev/null
echo "Categorizer files:"
ls pkg/categorizer/*.go 2>/dev/null
echo "Total Go files in repo:"
find . -name "*.go" -not -path "./vendor/*" -type f | wc -l
```

---

## Output

Report a comprehensive summary:

```
## Verification Results

### File Structure
- New files present: [list]
- Deleted dirs confirmed: [list]

### Stale References
- Azure: [CLEAN / count found]
- Google: [CLEAN / count found]
- GitHub: [CLEAN / count found]

### Dependency Cleanup
- go.mod pruned: [YES/NO]
- Azure SDK removed: [YES/NO]
- Google Cloud removed: [YES/NO]

### Build
- go build: [PASS/FAIL]
- Errors fixed: [count]

### Vet
- go vet: [PASS/FAIL]

### Tests
- Total: [count]
- Passed: [count]
- Failed: [count] - [list]
- Skipped: [count]

### Issues Remaining
- [list any unresolved issues]
```

---

## Rules

- **Fix build errors aggressively** — iterate until the build passes
- **Report test failures honestly** — don't hide failures
- **Don't make unnecessary changes** — only fix what's broken
- **Don't add new features** — this is verification only
- **Preserve git-cleanliness** — don't leave debug prints or temp files
