// Package build provides build configuration and metadata for the driftctl binary.
package build

var env = "dev"

// This flag could be switched to false while building to create a binary without third party network calls
// That mean that following services will be disabled:
// - telemetry
// - version check
var enableUsageReporting = "true"

// Interface defines methods for querying the build environment.
type Interface interface {
	IsRelease() bool
	IsUsageReportingEnabled() bool
}

// Build holds the compiled-in build configuration for a driftctl binary.
type Build struct{}

// IsRelease reports whether the binary was built as a release.
func (b Build) IsRelease() bool {
	return env == "release"
}

// IsUsageReportingEnabled reports whether usage/telemetry reporting is enabled.
func (b Build) IsUsageReportingEnabled() bool {
	return b.IsRelease() && enableUsageReporting == "true"
}
