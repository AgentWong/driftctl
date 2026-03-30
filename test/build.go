// Package test provides test utilities and acceptance test support.
package test

// Build provides test-mode implementations of build metadata queries.
type Build struct{}

// IsRelease always returns false in the test build.
func (b Build) IsRelease() bool {
	return false
}

// IsUsageReportingEnabled always returns false in the test build.
func (b Build) IsUsageReportingEnabled() bool {
	return false
}
