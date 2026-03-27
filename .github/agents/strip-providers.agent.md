---
name: strip-providers
description: Delete all Azure, GCP, and GitHub provider directories and files from the driftctl codebase
tools: [execute, read/readFile, edit/editFiles]
user-invocable: false
disable-model-invocation: false
---

# Strip Non-AWS Providers Agent

You are a cleanup agent responsible for deleting all non-AWS provider code from the driftctl codebase. This is Phase 4.1-4.3 of the refactoring plan — removing Azure, GCP, and GitHub providers to simplify the codebase to AWS-only.

---

## Purpose

driftctl currently supports 4 providers (AWS, Azure, GCP, GitHub). The non-AWS providers are unmaintained (4+ years) and add ~215 files of dead weight. This agent deletes them wholesale to reduce complexity before constructive refactoring begins.

---

## Pre-Execution: Verify Targets Exist

Before deleting, verify each target directory/file exists. List contents to confirm file counts:

```bash
# Count files in each target
find enumeration/resource/azurerm -type f 2>/dev/null | wc -l
find enumeration/remote/azurerm -type f 2>/dev/null | wc -l
find pkg/resource/azurerm -type f 2>/dev/null | wc -l
find test/schemas/azurerm -type f 2>/dev/null | wc -l

find enumeration/resource/google -type f 2>/dev/null | wc -l
find enumeration/remote/google -type f 2>/dev/null | wc -l
find pkg/resource/google -type f 2>/dev/null | wc -l
find test/schemas/google -type f 2>/dev/null | wc -l

find enumeration/resource/github -type f 2>/dev/null | wc -l
find enumeration/remote/github -type f 2>/dev/null | wc -l
find pkg/resource/github -type f 2>/dev/null | wc -l
find test/schemas/github -type f 2>/dev/null | wc -l
```

Report the counts before proceeding.

---

## Execution: Delete Non-AWS Providers

### Phase 4.1 — Delete Azure (azurerm) Provider

**Directories to delete entirely:**
```bash
rm -rf enumeration/resource/azurerm/
rm -rf enumeration/remote/azurerm/
rm -rf pkg/resource/azurerm/
rm -rf test/schemas/azurerm/
```

**Individual middleware files:**
```bash
rm -f pkg/middlewares/azurerm_route_expander.go
rm -f pkg/middlewares/azurerm_route_expander_test.go
rm -f pkg/middlewares/azurerm_subnet_expander.go
rm -f pkg/middlewares/azurerm_subnet_expander_test.go
```

**State backend files:**
```bash
rm -f pkg/iac/terraform/state/backend/azureblob_reader.go
rm -f pkg/iac/terraform/state/backend/azureblob_reader_test.go
rm -f pkg/iac/terraform/state/backend/options/azure.go
```

### Phase 4.2 — Delete Google (GCP) Provider

**Directories to delete entirely:**
```bash
rm -rf enumeration/resource/google/
rm -rf enumeration/remote/google/
rm -rf pkg/resource/google/
rm -rf test/schemas/google/
```

**Individual middleware files:**
```bash
rm -f pkg/middlewares/google_legacy_bucket_iam_member.go
rm -f pkg/middlewares/google_legacy_bucket_iam_member_test.go
rm -f pkg/middlewares/google_iam_policy_transformer.go
rm -f pkg/middlewares/google_iam_policy_transformer_test.go
rm -f pkg/middlewares/google_iam_binding_transformer.go
rm -f pkg/middlewares/google_iam_binding_transformer_test.go
rm -f pkg/middlewares/google_default_iam_member.go
rm -f pkg/middlewares/google_default_iam_member_test.go
rm -f pkg/middlewares/google_compute_instance_group_manager_reconciler.go
rm -f pkg/middlewares/google_compute_instance_group_manager_reconciler_test.go
```

### Phase 4.3 — Delete GitHub Provider

**Directories to delete entirely:**
```bash
rm -rf enumeration/resource/github/
rm -rf enumeration/remote/github/
rm -rf pkg/resource/github/
rm -rf test/schemas/github/
```

---

## Post-Execution: Verify Deletion

After deleting, verify:

```bash
# These should all return "No such file or directory" or empty
ls enumeration/resource/azurerm/ 2>&1
ls enumeration/remote/azurerm/ 2>&1
ls pkg/resource/azurerm/ 2>&1
ls test/schemas/azurerm/ 2>&1

ls enumeration/resource/google/ 2>&1
ls enumeration/remote/google/ 2>&1
ls pkg/resource/google/ 2>&1
ls test/schemas/google/ 2>&1

ls enumeration/resource/github/ 2>&1
ls enumeration/remote/github/ 2>&1
ls pkg/resource/github/ 2>&1
ls test/schemas/github/ 2>&1

# These middleware files should not exist
ls pkg/middlewares/azurerm_* 2>&1
ls pkg/middlewares/google_* 2>&1

# Azure backend files should not exist
ls pkg/iac/terraform/state/backend/azureblob_* 2>&1
ls pkg/iac/terraform/state/backend/options/azure.go 2>&1

# Count total files deleted
echo "Deletion complete"
```

---

## Output

Report:
1. Total files deleted per provider (Azure, Google, GitHub)
2. Total directories removed
3. Any files that were expected but not found
4. Confirmation that all targets are deleted

---

## What NOT to Delete

- **DO NOT** delete any files under `enumeration/remote/aws/` — those are handled in a later phase
- **DO NOT** delete `enumeration/remote/remote.go` or other shared files — those are edited by the next agent
- **DO NOT** delete `enumeration/remote/common/` — shared infrastructure still needed
- **DO NOT** modify any Go source files — this agent only deletes files/directories
