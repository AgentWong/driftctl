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

// Init runs terraform init for the working directory.
func (r *Runner) Init(ctx context.Context) error {
	tf, err := tfexec.NewTerraform(r.workingDir, r.execPath)
	if err != nil {
		return fmt.Errorf("failed to initialize terraform: %w", err)
	}
	if err := tf.Init(ctx, tfexec.Upgrade(false)); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}
	return nil
}

// RunPlanOnly executes terraform plan (assumes Init has already been run) and
// returns the JSON-parsed plan.
//
// -lock=false is passed because driftctl uses ReadOnly credentials which cannot
// perform the s3:PutObject required to acquire a remote state lock.
func (r *Runner) RunPlanOnly(ctx context.Context) (*tfjson.Plan, error) {
	tf, err := tfexec.NewTerraform(r.workingDir, r.execPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize terraform: %w", err)
	}

	planFile := "driftctl-plan.tfplan"
	_, err = tf.Plan(ctx, tfexec.Out(planFile), tfexec.Lock(false))
	if err != nil {
		return nil, fmt.Errorf("terraform plan failed: %w", err)
	}

	p, err := tf.ShowPlanFile(ctx, planFile)
	if err != nil {
		return nil, fmt.Errorf("terraform show failed: %w", err)
	}

	return p, nil
}

// RunPlan executes terraform init + plan and returns the JSON-parsed plan.
// It is a convenience wrapper around Init + RunPlanOnly for single-module use.
func (r *Runner) RunPlan(ctx context.Context) (*tfjson.Plan, error) {
	if err := r.Init(ctx); err != nil {
		return nil, err
	}
	return r.RunPlanOnly(ctx)
}
