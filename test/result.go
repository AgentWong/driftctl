package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/snyk/driftctl/pkg/analyser"

	"github.com/stretchr/testify/require"
)

// ScanResult wraps a scan Analysis with testify assertion helpers for convenient test assertions.
type ScanResult struct {
	*require.Assertions
	*analyser.Analysis
}

// NewScanResult creates a new ScanResult wrapping the given analysis and test instance.
func NewScanResult(t *testing.T, analysis *analyser.Analysis) *ScanResult {
	return &ScanResult{
		Assertions: require.New(t),
		Analysis:   analysis,
	}
}

// AssertResourceUnmanaged fails the test if no unmanaged resource with the given id and type is found.
func (r *ScanResult) AssertResourceUnmanaged(id, ty string) {
	for _, u := range r.Unmanaged() {
		if u.ResourceType() == ty && u.ResourceId() == id {
			return
		}
	}
	r.Failf("Resource not unmanaged", "%s(%s)", id, ty)
}

// AssertResourceDeleted fails the test if no deleted resource with the given id and type is found.
func (r *ScanResult) AssertResourceDeleted(id, ty string) {
	for _, u := range r.Deleted() {
		if u.ResourceType() == ty && u.ResourceId() == id {
			return
		}
	}
	r.Failf("Resource not deleted", "%s(%s)", id, ty)
}

// AssertCoverage asserts that the scan coverage equals the expected percentage.
func (r *ScanResult) AssertCoverage(expected int) {
	r.Equal(expected, r.Coverage())
}

// AssertDeletedCount asserts that the number of deleted resources equals count.
func (r *ScanResult) AssertDeletedCount(count int) {
	r.Equal(count, len(r.Deleted()))
}

// AssertManagedCount asserts that the number of managed resources equals count.
func (r *ScanResult) AssertManagedCount(count int) {
	r.Equal(count, len(r.Managed()))
}

// AssertUnmanagedCount asserts that the number of unmanaged resources equals count.
func (r *ScanResult) AssertUnmanagedCount(count int) {
	r.Equal(count, len(r.Unmanaged()))
}

// AssertInfrastructureIsInSync fails the test if the scanned infrastructure is not fully in sync.
func (r ScanResult) AssertInfrastructureIsInSync() {
	r.Equal(
		true,
		r.IsSync(),
		fmt.Sprintf(
			"Infrastructure is not in sync: \n%s\n",
			r.printAnalysisResult(),
		),
	)
}

// AssertInfrastructureIsNotSync fails the test if the scanned infrastructure is unexpectedly in sync.
func (r ScanResult) AssertInfrastructureIsNotSync() {
	r.Equal(
		false,
		r.IsSync(),
		fmt.Sprintf(
			"Infrastructure is in sync: \n%s\n",
			r.printAnalysisResult(),
		),
	)
}

func (r *ScanResult) printAnalysisResult() string {
	str, _ := json.MarshalIndent(r.Analysis, "", " ")
	return string(str)
}
