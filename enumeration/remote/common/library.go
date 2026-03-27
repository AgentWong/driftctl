package common

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

type Enumerator interface {
	SupportedType() resource.ResourceType
	Enumerate() ([]*resource.Resource, error)
}

// EnumerationFilter is a local interface that mirrors enumeration.Filter,
// avoiding a circular import (common -> enumeration -> ... -> common).
// enumeration.Filter satisfies this interface via structural typing.
type EnumerationFilter interface {
	IsTypeIgnored(ty resource.ResourceType) bool
	IsResourceIgnored(res *resource.Resource) bool
}

// BulkEnumerator discovers multiple resource types in a single API call,
// replacing N individual enumerators with one batch operation.
type BulkEnumerator interface {
	SupportedTypes() []resource.ResourceType
	Enumerate(filter EnumerationFilter) ([]*resource.Resource, error)
}

type RemoteLibrary struct {
	enumerators     []Enumerator
	bulkEnumerators []BulkEnumerator
}

func NewRemoteLibrary() *RemoteLibrary {
	return &RemoteLibrary{
		enumerators:     make([]Enumerator, 0),
		bulkEnumerators: make([]BulkEnumerator, 0),
	}
}

func (r *RemoteLibrary) AddEnumerator(enumerator Enumerator) {
	r.enumerators = append(r.enumerators, enumerator)
}

func (r *RemoteLibrary) Enumerators() []Enumerator {
	return r.enumerators
}

func (r *RemoteLibrary) AddBulkEnumerator(b BulkEnumerator) {
	r.bulkEnumerators = append(r.bulkEnumerators, b)
}

func (r *RemoteLibrary) GetBulkEnumerators() []BulkEnumerator {
	return r.bulkEnumerators
}
