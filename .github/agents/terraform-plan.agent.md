---
name: terraform-plan
description: Create terraform plan runner, parser, plan analyzer, and integrate plan-based drift detection into DriftCTL
tools: [execute, read/readFile, edit/createFile, edit/editFiles]
user-invocable: false
disable-model-invocation: false
---

# Terraform Plan Agent

You are a Go developer agent responsible for building the terraform plan-based drift detection system. This is Phase 2.1-2.5 of the refactoring plan. This system uses `terraform plan` output to detect actual configuration drift (attribute-level changes), not just resource existence.

---

## Purpose

The current system only compares resource existence (ID+Type). This agent adds true configuration drift detection by:
1. Running `terraform plan` on the user's Terraform configuration
2. Parsing the plan output to identify drifted, deleted, and unmanaged resources
3. Combining plan results with AWS Config inventory for complete coverage

---

## Pre-Execution: Study Existing Code

**You MUST read these files before writing code:**

1. **`go.mod`** — Verify `terraform-exec` and `terraform-json` versions available
2. **`pkg/analyser/analysis.go`** — Understand the `Analysis` struct, `Summary`, serialization
3. **`pkg/driftctl.go`** — Understand the `DriftCTL` struct, `Run()` method, options, middleware chain
4. **`pkg/cmd/scan.go`** — Understand CLI flag registration and option binding
5. **`pkg/analyser/analyzer.go`** — Understand the existing `Analyzer` interface if one exists

---

## Task 1: Create Plan Runner

**File:** `pkg/terraform/plan/runner.go` (CREATE — ensure `pkg/terraform/plan/` directory is created)

```go
package plan

import (
    "context"
    "fmt"

    "github.com/hashicorp/terraform-exec/tfexec"
    tfjson "github.com/hashicorp/terraform-json"
)

// Runner executes terraform plan and returns the plan output
type Runner struct {
    workingDir string
    execPath   string // path to terraform binary
}

func NewRunner(workingDir string, execPath string) *Runner {
    return &Runner{workingDir: workingDir, execPath: execPath}
}

// RunPlan executes terraform init + plan and returns the parsed plan
func (r *Runner) RunPlan(ctx context.Context) (*tfjson.Plan, error) {
    tf, err := tfexec.NewTerraform(r.workingDir, r.execPath)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize terraform: %w", err)
    }

    // Init
    if err := tf.Init(ctx, tfexec.Upgrade(false)); err != nil {
        return nil, fmt.Errorf("terraform init failed: %w", err)
    }

    // Plan — output to temp file
    planFile := "driftctl-plan.tfplan"
    _, err = tf.Plan(ctx, tfexec.Out(planFile))
    if err != nil {
        return nil, fmt.Errorf("terraform plan failed: %w", err)
    }

    // Show plan as JSON
    plan, err := tf.ShowPlanFile(ctx, planFile)
    if err != nil {
        return nil, fmt.Errorf("terraform show failed: %w", err)
    }

    return plan, nil
}
```

Adjust the implementation based on the actual `terraform-exec` API version in `go.mod`. The API may differ between v0.14 and newer versions.

---

## Task 2: Create Plan Parser

**File:** `pkg/terraform/plan/parser.go` (CREATE)

```go
package plan

import tfjson "github.com/hashicorp/terraform-json"

// Action represents the type of drift detected
type Action string

const (
    ActionCreate Action = "create"  // Resource exists in config but not deployed
    ActionUpdate Action = "update"  // Resource exists but has drifted attributes
    ActionDelete Action = "delete"  // Resource deployed but not in config
    ActionNoOp   Action = "no-op"   // Resource is in sync
)

// AttributeChange represents a single attribute that has drifted
type AttributeChange struct {
    Path     string      // Attribute path (e.g., "tags.Name")
    Before   interface{} // Current value
    After    interface{} // Desired value
}

// DriftResult represents the drift status of a single resource
type DriftResult struct {
    Type             string             // Terraform resource type (e.g., "aws_instance")
    ID               string             // Resource ID
    Address          string             // Terraform address (e.g., "aws_instance.web")
    Action           Action             // What kind of change
    AttributeChanges []AttributeChange  // Specific attributes that changed (for updates)
}

// ParsePlan extracts DriftResults from a terraform plan
func ParsePlan(p *tfjson.Plan) ([]DriftResult, error) {
    var results []DriftResult

    if p.ResourceChanges == nil {
        return results, nil
    }

    for _, rc := range p.ResourceChanges {
        result := DriftResult{
            Type:    rc.Type,
            Address: rc.Address,
        }

        // Extract ID from prior state if available
        if rc.Change != nil && rc.Change.Before != nil {
            if beforeMap, ok := rc.Change.Before.(map[string]interface{}); ok {
                if id, ok := beforeMap["id"].(string); ok {
                    result.ID = id
                }
            }
        }

        // Map terraform plan actions to our Action type
        result.Action = mapActions(rc.Change.Actions)

        // For updates, extract attribute changes
        if result.Action == ActionUpdate && rc.Change != nil {
            result.AttributeChanges = extractAttributeChanges(rc.Change)
        }

        results = append(results, result)
    }

    return results, nil
}
```

Implement `mapActions()` and `extractAttributeChanges()` helper functions. The `tfjson.Actions` type maps to create/update/delete/no-op.

---

## Task 3: Extend Analysis Model

**File:** `pkg/analyser/analysis.go` (EDIT)

Add to the existing structures:

1. **New type `DriftedResource`:**
```go
type DriftedResource struct {
    Resource         resource.Resource
    AttributeChanges []AttributeChange
}

type AttributeChange struct {
    Path   string      `json:"path"`
    Before interface{} `json:"before"`
    After  interface{} `json:"after"`
}
```

2. **Add to `Analysis` struct:**
```go
drifted []*DriftedResource
```

3. **Add to `Summary` struct:**
```go
TotalDrifted int `json:"total_drifted"`
```

4. **Update `IsSync()`:**
```go
func (a *Analysis) IsSync() bool {
    return a.summary.TotalUnmanaged == 0 &&
        a.summary.TotalDeleted == 0 &&
        a.summary.TotalDrifted == 0
}
```

5. **Add accessor methods:**
```go
func (a *Analysis) Drifted() []*DriftedResource { return a.drifted }
func (a *Analysis) AddDrifted(d *DriftedResource) { ... }
```

6. **Update JSON serialization** — add `drifted` to `serializableAnalysis` and `MarshalJSON()`

---

## Task 4: Create Plan Analyzer

**File:** `pkg/analyser/plan_analyzer.go` (CREATE)

```go
package analyser

// PlanAnalyzer combines terraform plan results with Config inventory
type PlanAnalyzer struct {
    planResults  []plan.DriftResult
    configResources []*resource.Resource  // From AWS Config inventory
}

func NewPlanAnalyzer(planResults []plan.DriftResult, configResources []*resource.Resource) *PlanAnalyzer {
    return &PlanAnalyzer{
        planResults:     planResults,
        configResources: configResources,
    }
}

// Analyze produces an Analysis combining plan + inventory data
func (a *PlanAnalyzer) Analyze() (*Analysis, error) {
    analysis := NewAnalysis()

    planResourceIDs := make(map[string]bool)

    for _, pr := range a.planResults {
        planResourceIDs[pr.Type+"."+pr.ID] = true

        switch pr.Action {
        case plan.ActionUpdate:
            // Drifted — has attribute changes
            analysis.AddDrifted(&DriftedResource{
                Resource: resource.Resource{Type: pr.Type, Id: pr.ID},
                AttributeChanges: convertChanges(pr.AttributeChanges),
            })
        case plan.ActionDelete:
            // Deleted from config, still in state
            analysis.AddDeleted(...)
        case plan.ActionCreate:
            // In config but not deployed — managed but missing
            analysis.AddManaged(...)
        case plan.ActionNoOp:
            // In sync
            analysis.AddManaged(...)
        }
    }

    // Resources in Config inventory but NOT in plan = unmanaged
    for _, cr := range a.configResources {
        key := cr.ResourceType() + "." + cr.ResourceId()
        if !planResourceIDs[key] {
            analysis.AddUnmanaged(cr)
        }
    }

    return analysis, nil
}
```

Adjust the resource construction to match the actual `resource.Resource` struct fields (read `analysis.go` and related files to understand the correct types).

---

## Task 5: Integrate into DriftCTL

**File:** `pkg/driftctl.go` (EDIT)

1. Add a `Mode` field to `ScanOptions`:
```go
Mode string // "inventory" (default) or "plan"
```

2. Add a `TerraformDir` field:
```go
TerraformDir string // path to terraform root module
```

3. In the `Run()` method, add plan-based flow:
```go
if d.opts.Mode == "plan" {
    // Run terraform plan
    runner := plan.NewRunner(d.opts.TerraformDir, "terraform")
    tfPlan, err := runner.RunPlan(context.Background())
    if err != nil {
        return nil, err
    }

    // Parse plan
    driftResults, err := plan.ParsePlan(tfPlan)
    if err != nil {
        return nil, err
    }

    // Get Config inventory for unmanaged detection
    // (resources from remote supplier)
    remoteResources, err := d.remoteSupplier.Resources()
    // ...

    // Analyze
    analyzer := analyser.NewPlanAnalyzer(driftResults, remoteResources)
    return analyzer.Analyze()
}

// Existing inventory-based flow continues below...
```

---

## Task 6: Add CLI Flags

**File:** `pkg/cmd/scan.go` (EDIT)

Add flags:
```go
cmd.Flags().StringVar(&opts.Mode, "mode", "inventory", "Scan mode: 'inventory' (default) or 'plan'")
cmd.Flags().StringVar(&opts.TerraformDir, "terraform-dir", "", "Path to Terraform root module (required for plan mode)")
```

Add validation in `PreRunE`:
```go
if opts.Mode == "plan" && opts.TerraformDir == "" {
    return fmt.Errorf("--terraform-dir is required when using --mode=plan")
}
```

---

## Post-Execution: Verify

```bash
# Ensure new directory exists
ls -la pkg/terraform/plan/

# Check new files
ls -la pkg/terraform/plan/runner.go
ls -la pkg/terraform/plan/parser.go
ls -la pkg/analyser/plan_analyzer.go

# Build check
go build ./...
```

---

## Output

Report:
1. Files created and their structure
2. Files modified and what changed
3. Build result (PASS/FAIL)
4. Any issues and fixes applied

---

## Rules

- **Read existing code first** — match patterns for Analysis, Resource types, error handling
- **Use existing dependencies** — terraform-exec and terraform-json are already in go.mod
- **Preserve backward compatibility** — default mode is "inventory", existing behavior unchanged
- **Do NOT modify scanner.go** — that was handled by a previous agent
- **Fix compilation errors** — iterate until `go build ./...` passes
