---
name: categorizer-output
description: Create categorizer framework for false positive filtering and update output formatters for drift details
tools: [execute, read/readFile, edit/createFile, edit/editFiles]
user-invocable: false
disable-model-invocation: false
---

# Categorizer & Output Agent

You are a Go developer agent responsible for building the categorization framework for false positive filtering and updating the output formatters to render drift details. This covers Phase 2.6 (output formatters) and Phase 3 (categorization) of the refactoring plan.

---

## Purpose

### Categorization (Phase 3)
Resources detected as "unmanaged" by the inventory scan may be false positives — CloudFormation-managed resources, service-linked roles, or resources unsupported by Terraform. The categorizer framework classifies each resource so users can filter out expected noise.

### Output Formatters (Phase 2.6)
The existing output formatters only show managed/unmanaged/deleted counts. They need to be updated to show:
- Drifted resources with attribute-level changes (from terraform plan)
- Category grouping (from categorizer)

---

## Pre-Execution: Study Existing Code

**You MUST read these files before writing code:**

1. **`pkg/cmd/scan/output/`** — List and read all output formatter files (console, JSON, HTML, plan) to understand the current output patterns
2. **`pkg/analyser/analysis.go`** — Understand the Analysis struct (including the drifted field added by the previous agent)
3. **`enumeration/remote/aws/config_resource_mapping.go`** — Understand the mapping table (for unsupported resource detection)
4. **`pkg/cmd/scan.go`** — Understand how flags are registered and passed to formatters
5. **`pkg/resource/resource.go`** or similar — Understand the Resource type structure

```bash
ls pkg/cmd/scan/output/
```

---

## Part 1: Categorizer Framework

### Task 1: Create Categorizer Interface

**File:** `pkg/categorizer/categorizer.go` (CREATE — ensure `pkg/categorizer/` directory is created)

```go
package categorizer

import "github.com/snyk/driftctl/enumeration/resource"

// Category represents why a resource appears as unmanaged
type Category string

const (
    CategoryManaged              Category = "managed"
    CategoryUnmanaged            Category = "unmanaged"
    CategoryCloudFormationManaged Category = "cloudformation_managed"
    CategoryServiceLinked        Category = "service_linked"
    CategoryUnsupported          Category = "unsupported"
)

// Categorizer classifies a resource into a category
type Categorizer interface {
    Categorize(r *resource.Resource) (Category, bool)
    // Returns (category, matched). If matched=false, the next categorizer in the chain is tried.
}

// Chain applies categorizers in order, returning the first match
type Chain struct {
    categorizers []Categorizer
}

func NewChain(categorizers ...Categorizer) *Chain {
    return &Chain{categorizers: categorizers}
}

func (c *Chain) Categorize(r *resource.Resource) Category {
    for _, cat := range c.categorizers {
        if category, matched := cat.Categorize(r); matched {
            return category
        }
    }
    return CategoryUnmanaged // Default if no categorizer matches
}
```

### Task 2: CloudFormation-Managed Detection

**File:** `pkg/categorizer/cloudformation.go` (CREATE)

```go
package categorizer

import "github.com/snyk/driftctl/enumeration/resource"

// CloudFormationCategorizer detects resources managed by CloudFormation stacks
type CloudFormationCategorizer struct{}

func NewCloudFormationCategorizer() *CloudFormationCategorizer {
    return &CloudFormationCategorizer{}
}

func (c *CloudFormationCategorizer) Categorize(r *resource.Resource) (Category, bool) {
    // Check for aws:cloudformation:stack-name tag
    // Resources managed by CloudFormation have this tag set
    attrs := r.Attributes()
    if attrs == nil {
        return "", false
    }

    tags, ok := (*attrs)["tags"]
    if !ok {
        return "", false
    }

    tagsMap, ok := tags.(map[string]interface{})
    if !ok {
        return "", false
    }

    if _, hasStackTag := tagsMap["aws:cloudformation:stack-name"]; hasStackTag {
        return CategoryCloudFormationManaged, true
    }

    return "", false
}
```

Adjust the attribute access pattern to match how `resource.Resource` actually stores attributes (read the Resource type definition).

### Task 3: Service-Linked Role Detection

**File:** `pkg/categorizer/service_linked.go` (CREATE)

```go
package categorizer

import (
    "strings"
    "github.com/snyk/driftctl/enumeration/resource"
)

// ServiceLinkedCategorizer detects AWS service-linked roles
type ServiceLinkedCategorizer struct{}

func NewServiceLinkedCategorizer() *ServiceLinkedCategorizer {
    return &ServiceLinkedCategorizer{}
}

func (c *ServiceLinkedCategorizer) Categorize(r *resource.Resource) (Category, bool) {
    // Only applies to IAM roles
    if r.ResourceType() != "aws_iam_role" {
        return "", false
    }

    id := r.ResourceId()

    // Service-linked roles have paths like /aws-service-role/
    if strings.Contains(id, "/aws-service-role/") {
        return CategoryServiceLinked, true
    }

    // Or names like AWSServiceRoleFor*
    attrs := r.Attributes()
    if attrs != nil {
        if name, ok := (*attrs)["name"].(string); ok {
            if strings.HasPrefix(name, "AWSServiceRoleFor") {
                return CategoryServiceLinked, true
            }
        }
        if path, ok := (*attrs)["path"].(string); ok {
            if strings.HasPrefix(path, "/aws-service-role/") {
                return CategoryServiceLinked, true
            }
        }
    }

    return "", false
}
```

### Task 4: Unsupported Resource Detection

**File:** `pkg/categorizer/unsupported.go` (CREATE)

```go
package categorizer

import "github.com/snyk/driftctl/enumeration/resource"

// UnsupportedCategorizer detects resources not supported by AWS Config or Terraform
type UnsupportedCategorizer struct {
    supportedTypes map[string]bool
}

func NewUnsupportedCategorizer(supportedTypes []string) *UnsupportedCategorizer {
    typeMap := make(map[string]bool, len(supportedTypes))
    for _, t := range supportedTypes {
        typeMap[t] = true
    }
    return &UnsupportedCategorizer{supportedTypes: typeMap}
}

func (c *UnsupportedCategorizer) Categorize(r *resource.Resource) (Category, bool) {
    if !c.supportedTypes[r.ResourceType()] {
        return CategoryUnsupported, true
    }
    return "", false
}
```

### Task 5: Add --exclude-category Flag

**File:** `pkg/cmd/scan.go` (EDIT)

Add flag:
```go
cmd.Flags().StringSliceVar(&opts.ExcludeCategories, "exclude-category", nil,
    "Exclude resources by category: cloudformation_managed, service_linked, unsupported")
```

Add `ExcludeCategories []string` to `ScanOptions` in `pkg/driftctl.go` if not already present.

---

## Part 2: Output Formatters

Read all existing output formatters in `pkg/cmd/scan/output/` to understand the current structure. Then update each:

### Task 6: Update Console Output

Find the console/text output formatter and update it to:
1. Add a "Drifted" section showing resources with attribute changes:
   ```
   Found drifted resources:
     - aws_instance.web (i-1234567890):
       ~ tags.Name: "old-name" => "new-name"
       ~ instance_type: "t2.micro" => "t2.small"
   ```
2. Add category grouping for unmanaged resources:
   ```
   Unmanaged resources by category:
     CloudFormation-managed (3):
       - aws_s3_bucket: my-cfn-bucket
     Service-linked (2):
       - aws_iam_role: AWSServiceRoleForECS
   ```
3. Update the summary line to include drifted count

### Task 7: Update JSON Output

Find the JSON output formatter and update it to include:
- `"drifted"` array with attribute changes
- `"categories"` grouping for unmanaged resources
- Updated summary with `total_drifted`

### Task 8: Update HTML Output

Find the HTML output formatter and update it to include:
- Drifted resources section with attribute change table
- Category grouping for unmanaged resources

---

## Post-Execution: Verify

```bash
# Verify new directory and files
ls -la pkg/categorizer/

# Verify new files exist
ls -la pkg/categorizer/categorizer.go
ls -la pkg/categorizer/cloudformation.go
ls -la pkg/categorizer/service_linked.go
ls -la pkg/categorizer/unsupported.go

# Build check
go build ./...
```

---

## Output

Report:
1. Categorizer files created and their interfaces
2. Output formatters modified and what changed
3. Build result (PASS/FAIL)
4. Any issues and fixes applied

---

## Rules

- **Read output formatters before modifying** — understand current patterns
- **Read resource.Resource type** — understand how attributes are accessed
- **Preserve existing output format** — add to it, don't replace it
- **Match existing code style** — imports, error handling, formatting
- **Fix compilation errors** — iterate until `go build ./...` passes
