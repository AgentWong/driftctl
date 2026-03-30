// Package scan provides the CLI scan command and its associated exit codes and options.
package scan

const (
	// ExitInSync is the exit code when infrastructure is in sync.
	ExitInSync = 0
	// ExitNotInSync is the exit code when infrastructure is not in sync.
	ExitNotInSync = 1
	// ExitError is the exit code when an error occurs during scanning.
	ExitError = 2
)
