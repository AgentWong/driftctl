// Package error defines error types for remote resource scanning operations.
package error

import "fmt"

// RemoteError is the interface for errors that occurred while listing remote resources.
type RemoteError interface {
	ListedTypeError() string
}

// ResourceScanningError represents an error that occurred while scanning a specific resource or resource type.
type ResourceScanningError struct {
	err             error
	resourceType    string
	resourceID      string
	listedTypeError string
}

func (b *ResourceScanningError) Error() string {
	if b.resourceID != "" {
		return fmt.Sprintf("error scanning resource %s: %s", b.Resource(), b.err)
	}
	return fmt.Sprintf("error scanning resource type %s: %s", b.Resource(), b.err)
}

// RootCause returns the underlying error that caused the scanning error.
func (b *ResourceScanningError) RootCause() error {
	return b.err
}

// ResourceType returns the Terraform resource type associated with the error.
func (b *ResourceScanningError) ResourceType() string {
	return b.resourceType
}

// NewResourceScanningError creates a ResourceScanningError for a specific resource.
func NewResourceScanningError(err error, resourceType string, resourceID string) *ResourceScanningError {
	return &ResourceScanningError{
		err:             err,
		resourceType:    resourceType,
		resourceID:      resourceID,
		listedTypeError: resourceType,
	}
}

// NewResourceListingError creates a ResourceScanningError for a resource type listing failure.
func NewResourceListingError(err error, resourceType string) *ResourceScanningError {
	return NewResourceListingErrorWithType(err, resourceType, resourceType)
}

// NewResourceListingErrorWithType creates a ResourceScanningError with a distinct listed type.
func NewResourceListingErrorWithType(err error, resourceType, listedTypeError string) *ResourceScanningError {
	return &ResourceScanningError{
		err:             err,
		resourceType:    resourceType,
		listedTypeError: listedTypeError,
	}
}

// ListedTypeError returns the resource type that was being listed when the error occurred.
func (b *ResourceScanningError) ListedTypeError() string {
	return b.listedTypeError
}

// Resource returns a string identifying the resource or resource type that caused the error.
func (b *ResourceScanningError) Resource() string {
	if b.resourceID != "" {
		return fmt.Sprintf("%s.%s", b.resourceType, b.resourceID)
	}
	return b.resourceType
}

func (b *ResourceScanningError) String() string {
	return fmt.Sprintf("%s.%s (%s)", b.resourceType, b.resourceID, b.listedTypeError)
}
