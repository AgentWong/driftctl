// Package errors defines error types used by the scan command.
package errors

// UsageError represents a user input error.
type UsageError struct {
	msg string
}

// NewUsageError creates a UsageError with the given message.
func NewUsageError(msg string) UsageError {
	return UsageError{msg}
}

func (u UsageError) Error() string {
	return u.msg
}
