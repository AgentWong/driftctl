// Package common provides shared types and helpers for remote resource enumeration.
package common

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

// EnumerationFilter is a local interface that mirrors enumeration.Filter,
// avoiding a circular import (common -> enumeration -> ... -> common).
// enumeration.Filter satisfies this interface via structural typing.
type EnumerationFilter interface {
	IsTypeIgnored(ty resource.Type) bool
	IsResourceIgnored(res *resource.Resource) bool
}

// BulkEnumerator discovers multiple resource types in a single API call.
type BulkEnumerator interface {
	SupportedTypes() []resource.Type
	Enumerate(filter EnumerationFilter) ([]*resource.Resource, error)
}

// RemoteLibrary holds the registered bulk enumerators.
type RemoteLibrary struct {
	bulkEnumerators []BulkEnumerator
}

// NewRemoteLibrary creates an empty RemoteLibrary.
func NewRemoteLibrary() *RemoteLibrary {
	return &RemoteLibrary{
		bulkEnumerators: make([]BulkEnumerator, 0),
	}
}

// AddBulkEnumerator registers a bulk enumerator.
func (r *RemoteLibrary) AddBulkEnumerator(b BulkEnumerator) {
	r.bulkEnumerators = append(r.bulkEnumerators, b)
}

// GetBulkEnumerators returns all registered bulk enumerators.
func (r *RemoteLibrary) GetBulkEnumerators() []BulkEnumerator {
	return r.bulkEnumerators
}
