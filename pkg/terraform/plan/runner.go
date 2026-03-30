package plan

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

// Runner executes terraform plan and returns the parsed plan output.
type Runner struct {
	workingDir string
	execPath   string
}

// NewRunner creates a new Runner for the given Terraform working directory and binary path.
func NewRunner(workingDir string, execPath string) *Runner {
	return &Runner{workingDir: workingDir, execPath: execPath}
}

// RunPlan executes terraform init + plan and returns the JSON-parsed plan.
func (r *Runner) RunPlan(ctx context.Context) (*tfjson.Plan, error) {
	tf, err := tfexec.NewTerraform(r.workingDir, r.execPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize terraform: %w", err)
	}

	if err := tf.Init(ctx, tfexec.Upgrade(false)); err != nil {
		return nil, fmt.Errorf("terraform init failed: %w", err)
	}

	// Plan with -detailed-exitcode so we can detect changes vs no-changes,
	// but both cases produce a usable plan file.
	planFile := "driftctl-plan.tfplan"
	_, err = tf.Plan(ctx, tfexec.Out(planFile))
	if err != nil {
		return nil, fmt.Errorf("terraform plan failed: %w", err)
	}

	plan, err := tf.ShowPlanFile(ctx, planFile)
	if err != nil {
		return nil, fmt.Errorf("terraform show failed: %w", err)
	}

	return plan, nil
}
