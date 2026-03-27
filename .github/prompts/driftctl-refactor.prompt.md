---
name: driftctl-refactor
description: Orchestrate the driftctl refactoring — replace 103 AWS enumerators with Config API, add terraform plan drift detection, add categorization, strip non-AWS providers
tools:
  ['read/readFile', 'edit/createFile', 'edit/editFiles', 'agent']
---

# driftctl Refactoring Orchestrator

This workflow executes the full driftctl refactoring plan in phases using specialized subagents. Each phase builds on the previous, so **all subagents MUST execute serially** — wait for each to fully complete before invoking the next.

## Goal

Replace 103 individual AWS enumerators with a single AWS Config API call, integrate `terraform plan` for true configuration drift detection, add categorization/false-positive filtering, and strip non-AWS providers (Azure, GCP, GitHub) to simplify the codebase.

---

## Architecture: Orchestrator + Subagents (Serial Execution)

```
Orchestrator: Load plan → Execute phases in order → Verify build
    ↓
PHASE A — Provider Cleanup (reduces codebase before constructive work):
    Step 1: Subagent(strip-providers)     → Delete Azure/GCP/GitHub directories
    Step 2: Subagent(aws-only-cleanup)    → Update shared files for AWS-only
    ↓
PHASE B — AWS Config Inventory (replace 103 enumerators):
    Step 3: Subagent(config-inventory)    → Create Config repo, mapping, BulkEnumerator
    Step 4: Subagent(scanner-init-rewrite)→ Update Scanner + rewrite aws/init.go
    Step 5: Subagent(delete-enumerators)  → Delete old AWS enumerator/repo/mock files
    ↓
PHASE C — Terraform Plan Drift Detection:
    Step 6: Subagent(terraform-plan)      → Plan runner, parser, analyzer, DriftCTL integration
    ↓
PHASE D — Categorization + Output:
    Step 7: Subagent(categorizer-output)  → Categorizer framework + output formatter updates
    ↓
PHASE E — Verification:
    Step 8: Subagent(verify-build)        → go build, go test, go vet, go mod tidy
```

> **CRITICAL: SERIAL EXECUTION ONLY**
> All subagent invocations MUST be executed **one at a time, in strict sequential order**. You MUST wait for each `runSubagent()` call to fully return before invoking the next one. NEVER run multiple subagents in parallel. Later phases depend on files created/modified by earlier phases.

---

## Pre-Execution: Load Context

### 1. Read the Refactoring Plan

Read the full refactoring plan for reference:
- [Refactoring Plan](.ai_references/refactoring-plan.md)

### 2. Verify Starting State

Before starting, confirm:
```
git status  # Should be clean or on the refactoring branch
go build ./...  # Should compile (baseline)
```

If the build is already broken, stop and report — do not proceed with refactoring on a broken baseline.

---

## Step 1: Strip Non-AWS Providers (Phase 4.1-4.3)

Invoke the `strip-providers` agent to bulk-delete all Azure, GCP, and GitHub provider directories.

```
result = runSubagent(
    agent: "strip-providers",
    prompt: "Delete all non-AWS provider directories and files as specified in Phase 4.1-4.3 of the refactoring plan. This includes:
        - Azure: enumeration/resource/azurerm/, enumeration/remote/azurerm/, pkg/resource/azurerm/, test/schemas/azurerm/, pkg/middlewares/azurerm_*, pkg/iac/terraform/state/backend/azureblob_reader*, pkg/iac/terraform/state/backend/options/azure.go
        - Google: enumeration/resource/google/, enumeration/remote/google/, pkg/resource/google/, test/schemas/google/, pkg/middlewares/google_*
        - GitHub: enumeration/resource/github/, enumeration/remote/github/, pkg/resource/github/, test/schemas/github/"
)
```

**Expected outcome:** ~215 files deleted across Azure (~90), Google (~100), GitHub (~25) providers.

---

## Step 2: Update Shared Infrastructure for AWS-Only (Phase 4.4)

Invoke the `aws-only-cleanup` agent to surgically edit shared files that reference the deleted providers.

```
result = runSubagent(
    agent: "aws-only-cleanup",
    prompt: "Update shared infrastructure files to remove all non-AWS provider references. Edit these files:
        - enumeration/remote/remote.go — Remove imports/cases for azurerm, github, google; reduce supportedRemotes to AWS only
        - enumeration/terraform/providers.go — Delete GITHUB, GOOGLE, AZURE constants
        - enumeration/remote/common/providers.go — Remove non-AWS remote constants and mappings
        - enumeration/terraform/provider_config.go — Remove non-AWS provider download configs
        - pkg/resource/resource_types.go — Remove all github_*, google_*, azurerm_* resource type entries
        - pkg/driftctl.go — Remove non-AWS middleware registrations (Google IAM, Azure route/subnet expanders)
        - pkg/cmd/scan.go — Remove Azure CLI flags, non-AWS remote validation options"
)
```

**Expected outcome:** All shared files compile with AWS-only references. Non-AWS imports and constants removed.

---

## Step 3: Create AWS Config Inventory System (Phase 1.1-1.3)

Invoke the `config-inventory` agent to create the new Config-based inventory system.

```
result = runSubagent(
    agent: "config-inventory",
    prompt: "Create the AWS Config inventory system (Phase 1.1-1.3):
        1. Create enumeration/remote/aws/repository/config_repository.go — ConfigRepository wrapping configservice SDK with pagination + caching (follow existing repository patterns like s3_repository.go or ec2_repository.go for cache usage)
        2. Create enumeration/remote/aws/config_resource_mapping.go — Static mapping table: AWS Config type -> Terraform type (80+ mappings), reverse map, UnsupportedByConfig() helper
        3. Extend enumeration/remote/common/library.go — Add BulkEnumerator interface with SupportedTypes() and Enumerate(filter) methods; extend RemoteLibrary to hold []BulkEnumerator with AddBulkEnumerator() method
        4. Create enumeration/remote/aws/config_enumerator.go — Single enumerator implementing BulkEnumerator, calls Config API once via ConfigRepository, maps results using the mapping table

        Read existing repository files for patterns (cache usage, client initialization, error handling). The BulkEnumerator interface must coexist with the existing Enumerator interface."
)
```

**Expected outcome:** New files created, BulkEnumerator interface added to library.go, RemoteLibrary extended.

---

## Step 4: Update Scanner + Rewrite aws/init.go (Phase 1.4-1.5)

Invoke the `scanner-init-rewrite` agent to wire up the new system.

```
result = runSubagent(
    agent: "scanner-init-rewrite",
    prompt: "Update the Scanner and rewrite aws/init.go (Phase 1.4-1.5):
        1. Modify enumeration/remote/scanner.go — Update Resources() / scan() to handle both Enumerator (existing parallel runner) and BulkEnumerator (new). BulkEnumerators should be called first, then individual Enumerators for any types not covered by bulk.
        2. Rewrite enumeration/remote/aws/init.go — Replace 103 AddEnumerator() calls with single AddBulkEnumerator(configEnumerator). Keep provider/providerLibrary setup for terraform state reading. Keep the repository initialization for configservice client.

        Read the current scanner.go to understand the ParallelRunner pattern. Read the current aws/init.go to understand provider setup that must be preserved."
)
```

**Expected outcome:** Scanner handles both enumerator types. aws/init.go dramatically simplified.

---

## Step 5: Delete Old AWS Enumerators (Phase 1.6)

Invoke the `delete-enumerators` agent to remove the now-replaced individual enumerators.

```
result = runSubagent(
    agent: "delete-enumerators",
    prompt: "Delete old AWS enumerator code that has been replaced by the Config enumerator (Phase 1.6):
        - Delete all enumeration/remote/aws/*_enumerator.go files (103 files) EXCEPT config_enumerator.go
        - Delete old repository files in enumeration/remote/aws/repository/ EXCEPT config_repository.go — remove ec2_repository.go, s3_repository.go, iam_repository.go, etc. (23 files)
        - Delete associated mock files (mock_*.go) in the repository directory
        - Delete associated test files (*_test.go) for deleted enumerators and repositories
        - Do NOT delete: config_enumerator.go, config_repository.go, config_resource_mapping.go, init.go, provider.go, or any files needed by the new Config system

        Use 'find' or 'ls' to list files before deleting. Verify config_* files are preserved after deletion."
)
```

**Expected outcome:** ~126+ files deleted. Only Config-related files, init.go, and provider.go remain in the AWS directories.

---

## Step 6: Terraform Plan Drift Detection (Phase 2.1-2.5)

Invoke the `terraform-plan` agent to build the plan-based drift detection system.

```
result = runSubagent(
    agent: "terraform-plan",
    prompt: "Build the terraform plan-based drift detection system (Phase 2.1-2.5):
        1. Create pkg/terraform/plan/runner.go — TerraformPlanRunner using terraform-exec (already in go.mod) for Init() + Plan() + ShowPlanFile(). Accept terraform dir path.
        2. Create pkg/terraform/plan/parser.go — Parse tfjson.Plan (terraform-json already in go.mod) into DriftResult structs: type, ID, action (create/update/delete/no-op), attribute changes.
        3. Modify pkg/analyser/analysis.go — Add 'drifted []*DriftedResource' field, TotalDrifted to Summary, update IsSync() to check TotalDrifted == 0.
        4. Create pkg/analyser/plan_analyzer.go — Combine plan results with Config inventory: plan 'update' = drifted, plan 'delete' = deleted, Config resources not in plan = unmanaged.
        5. Modify pkg/driftctl.go — Add plan-based flow to Run(): new --mode=plan flag (default --mode=inventory for backward compat). Plan mode skips middleware chain.
        6. Modify pkg/cmd/scan.go — Add --terraform-dir and --mode flags.

        Read existing analysis.go and driftctl.go to understand current structures before modifying. Read go.mod to verify terraform-exec and terraform-json versions."
)
```

**Expected outcome:** New pkg/terraform/plan/ directory with runner and parser. Analysis model extended. DriftCTL supports plan mode.

---

## Step 7: Categorization + Output Formatters (Phase 2.6 + 3.1-3.5)

Invoke the `categorizer-output` agent to build the categorizer framework and update output formatters.

```
result = runSubagent(
    agent: "categorizer-output",
    prompt: "Build the categorizer framework and update output formatters (Phase 2.6 + 3.1-3.5):

        CATEGORIZERS:
        1. Create pkg/categorizer/categorizer.go — Categorizer interface with Categorize(*resource.Resource) Category method. Categories: cloudformation_managed, service_linked, unsupported, managed, unmanaged. Chain of categorizers applied to each resource.
        2. Create pkg/categorizer/cloudformation.go — Check aws:cloudformation:stack-name tag.
        3. Create pkg/categorizer/service_linked.go — Path-based check: /aws-service-role/ prefix or AWSServiceRoleFor* pattern.
        4. Create pkg/categorizer/unsupported.go — Check against Config mapping table for gaps.
        5. Integrate --exclude-category flag in pkg/cmd/scan.go.

        OUTPUT FORMATTERS:
        6. Update console output in pkg/cmd/scan/output/ to render drift details (attribute changes from plan) and category grouping.
        7. Update JSON output to include drifted resources and categories.
        8. Update HTML output to include drifted resources and categories.

        Read existing output formatters in pkg/cmd/scan/output/ to understand current patterns before modifying."
)
```

**Expected outcome:** New pkg/categorizer/ directory. Output formatters show drift details and categories.

---

## Step 8: Verification (Final)

Invoke the `verify-build` agent to run comprehensive checks.

```
result = runSubagent(
    agent: "verify-build",
    prompt: "Run comprehensive verification of the refactored codebase:
        1. Run 'go build ./...' — must compile cleanly
        2. Run 'go vet ./...' — must pass
        3. Run 'go mod tidy' — clean up dependencies (should prune Azure/Google/GitHub SDK modules)
        4. Verify no stale non-AWS references: grep -r 'azurerm\|RemoteGoogleTerraform\|RemoteGithubTerraform\|RemoteAzureTerraform' --include='*.go' pkg/ enumeration/
        5. Verify config_* files exist: ls enumeration/remote/aws/repository/config_repository.go enumeration/remote/aws/config_enumerator.go enumeration/remote/aws/config_resource_mapping.go
        6. Verify new directories exist: ls pkg/terraform/plan/ pkg/categorizer/
        7. Run 'go test ./...' — report results (some tests may need fixing)

        If build or vet fails, attempt to fix compilation errors. Report all results."
)
```

**Expected outcome:** Clean build, vet passes, no stale non-AWS references, all new files in place.

---

## Step 9: Update CHANGELOG.md

After all phases complete and verification passes, update `CHANGELOG.md` at the project root to reflect the actual changes made. The file already contains a pre-written v1.0.0 entry based on the planned changes. Review it against what was actually done and adjust if needed:

- If a planned feature was **not implemented** or was implemented differently, update or remove the entry
- If **additional changes** were made (e.g., extra files modified, unexpected fixes), add them
- Update the date to the current date if it differs
- Ensure file counts and names match reality

```
Read CHANGELOG.md, compare against the actual changes from phases A-E, and edit to reflect what was actually done.
```

**Expected outcome:** CHANGELOG.md accurately reflects all v1.0.0 changes.

---

## Post-Execution: Final Report

After all subagents complete, summarize:

```
## driftctl Refactoring Complete

### Phase A: Provider Cleanup
- Files deleted: ~215 (Azure ~90, Google ~100, GitHub ~25)
- Shared files updated: 7

### Phase B: AWS Config Inventory
- New files created: 3 (config_repository.go, config_resource_mapping.go, config_enumerator.go)
- Files modified: 2 (library.go, scanner.go)
- Files rewritten: 1 (aws/init.go)
- Old enumerators deleted: ~126

### Phase C: Terraform Plan
- New files created: 4 (runner.go, parser.go, plan_analyzer.go)
- Files modified: 3 (analysis.go, driftctl.go, scan.go)

### Phase D: Categorization + Output
- New files created: 4 (categorizer.go, cloudformation.go, service_linked.go, unsupported.go)
- Output formatters updated: 3 (console, JSON, HTML)

### Phase E: Verification
- Build: [PASS/FAIL]
- Vet: [PASS/FAIL]
- Tests: [X passed, Y failed]
- Stale references: [NONE/list]
```

---

## Error Handling

| Error | Action |
|-------|--------|
| Subagent fails to complete | Log the error, report which phase failed, stop execution |
| Build breaks after a phase | Attempt to fix in verify-build agent; if unfixable, report the phase that broke it |
| Files not found | Agent should report missing files; orchestrator adjusts instructions |
| Merge conflicts | Should not occur (sequential execution on same branch) |

---

## Important Notes

- **Do NOT skip phases** — each builds on the previous
- **Do NOT run subagents in parallel** — serial execution only
- **Do NOT modify files outside the scope** of each phase
- **Commit between phases** is optional but recommended for safety
- The refactoring plan is at `.ai_references/refactoring-plan.md` — subagents should read it for detailed context when needed
