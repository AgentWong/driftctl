// Package mocks provides test doubles and mock implementations used across the test suite.
package mocks

// MockBuild is a configurable mock implementation of the build metadata interface.
type MockBuild struct {
	Release        bool
	UsageReporting bool
}

// IsRelease returns the configured release flag.
func (m MockBuild) IsRelease() bool {
	return m.Release
}

// IsUsageReportingEnabled returns the configured usage reporting flag.
func (m MockBuild) IsUsageReportingEnabled() bool {
	return m.UsageReporting
}
