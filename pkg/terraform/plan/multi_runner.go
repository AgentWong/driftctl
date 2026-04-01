package plan

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
)

// ModuleResult holds the plan results for a single Terraform root module.
type ModuleResult struct {
	Dir     string
	Results []DriftResult
	Err     error
}

// RunParallel initialises each module sequentially (to avoid SSO token cache
// races when multiple processes refresh the same credential file simultaneously)
// and then runs terraform plan for all successfully initialised modules in
// parallel.  execPath is the path or name of the terraform binary to use.
func RunParallel(ctx context.Context, dirs []string, execPath string) []ModuleResult {
	results := make([]ModuleResult, len(dirs))

	// Phase 1 – sequential init to prevent concurrent SSO token file races.
	initOK := make([]bool, len(dirs))
	for i, dir := range dirs {
		logrus.WithField("dir", dir).Info("Running terraform init")
		runner := NewRunner(dir, execPath)
		if err := runner.Init(ctx); err != nil {
			logrus.WithField("dir", dir).WithError(err).Warn("Init failed")
			results[i] = ModuleResult{Dir: dir, Err: err}
		} else {
			initOK[i] = true
		}
	}

	// Phase 2 – parallel plan across modules that initialised successfully.
	var wg sync.WaitGroup
	for i, dir := range dirs {
		if !initOK[i] {
			continue
		}
		wg.Add(1)
		go func(idx int, d string) {
			defer wg.Done()

			logrus.WithField("dir", d).Info("Running terraform plan")
			runner := NewRunner(d, execPath)
			tfPlan, err := runner.RunPlanOnly(ctx)
			if err != nil {
				logrus.WithField("dir", d).WithError(err).Warn("Plan failed")
				results[idx] = ModuleResult{Dir: d, Err: err}
				return
			}

			driftResults, err := ParsePlan(tfPlan)
			if err != nil {
				logrus.WithField("dir", d).WithError(err).Warn("Plan parse failed")
				results[idx] = ModuleResult{Dir: d, Err: err}
				return
			}

			logrus.WithField("dir", d).WithField("resources", len(driftResults)).Info("Plan complete")
			results[idx] = ModuleResult{Dir: d, Results: driftResults}
		}(i, dir)
	}

	wg.Wait()
	return results
}
