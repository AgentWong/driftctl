package errors

// InfrastructureNotInSync indicates that the scan found drift.
type InfrastructureNotInSync struct{}

func (i InfrastructureNotInSync) Error() string {
	return "Infrastructure is not in sync"
}
