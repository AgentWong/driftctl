package analyser

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/terraform/plan"
)

// PlanAnalyzer combines terraform plan results with cloud inventory
// to produce an Analysis covering drifted, deleted, and unmanaged resources.
type PlanAnalyzer struct {
	planResults     []plan.DriftResult
	configResources []*resource.Resource
	sourceDir       string // Terraform root module directory, stamped onto each resource as its IaC source
}

// NewPlanAnalyzer creates a PlanAnalyzer from plan results and config resources.
// sourceDir is the Terraform root module directory; it is recorded as the IaC
// source on every resource so reports can show which module each resource came from.
func NewPlanAnalyzer(planResults []plan.DriftResult, configResources []*resource.Resource, sourceDir string) *PlanAnalyzer {
	return &PlanAnalyzer{
		planResults:     planResults,
		configResources: configResources,
		sourceDir:       sourceDir,
	}
}

// Analyze produces an Analysis from the plan results and config resources.
func (a *PlanAnalyzer) Analyze() (*Analysis, error) {
	analysis := NewAnalysis()

	// Track resources seen in the plan so we can identify unmanaged ones later
	planResourceIDs := make(map[string]bool)

	for _, pr := range a.planResults {
		if pr.ID != "" {
			planResourceIDs[pr.Type+"."+pr.ID] = true
		}

		res := &resource.Resource{
			Type:   pr.Type,
			ID:     pr.ID,
			Source: resource.NewTerraformStateSource(a.sourceDir, "", pr.Address),
		}

		switch pr.Action {
		case plan.ActionUpdate:
			analysis.AddDrifted(&DriftedResource{
				Res:              res,
				AttributeChanges: convertChanges(pr.AttributeChanges),
			})
		case plan.ActionDelete:
			analysis.AddDeleted(res)
		case plan.ActionCreate, plan.ActionNoOp, plan.ActionRead:
			analysis.AddManaged(res)
		}
	}

	// Resources found in cloud inventory but not in the plan are unmanaged
	for _, cr := range a.configResources {
		key := cr.ResourceType() + "." + cr.ResourceID()
		if !planResourceIDs[key] {
			analysis.AddUnmanaged(cr)
		}
	}

	analysis.SortResources()

	return analysis, nil
}

func convertChanges(planChanges []plan.AttributeChange) []AttributeChange {
	out := make([]AttributeChange, len(planChanges))
	for i, c := range planChanges {
		out[i] = AttributeChange{
			Path:   c.Path,
			Before: c.Before,
			After:  c.After,
		}
	}
	return out
}
