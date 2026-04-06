// Package terraform provides Terraform provider management, schema handling, and resource reading.
package terraform

import (
	"github.com/snyk/driftctl/enumeration/parallel"
	"github.com/zclconf/go-cty/cty"
)

// ParallelResourceReader reads multiple Terraform resources concurrently using a parallel.Runner.
type ParallelResourceReader struct {
	runner *parallel.Runner
}

// NewParallelResourceReader creates a ParallelResourceReader backed by the given runner.
func NewParallelResourceReader(runner *parallel.Runner) *ParallelResourceReader {
	return &ParallelResourceReader{
		runner: runner,
	}
}

// Wait blocks until all submitted reads complete and returns their results.
func (p *ParallelResourceReader) Wait() ([]cty.Value, error) {
	results := make([]cty.Value, 0)
Loop:
	for {
		select {
		case res, ok := <-p.runner.Read():
			if !ok {
				break Loop
			}
			ctyVal := res.(cty.Value)
			if !ctyVal.IsNull() {
				results = append(results, ctyVal)
			}
		case <-p.runner.DoneChan():
			break Loop
		}
	}
	return results, p.runner.Err()
}

// Run submits a resource read function for concurrent execution.
func (p *ParallelResourceReader) Run(runnable func() (cty.Value, error)) {
	p.runner.Run(func() (interface{}, error) {
		return runnable()
	})
}
