// Package remote provides test helpers for remote resource scanning.
package remote

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

// SortableScanner wraps a resource.Supplier and returns results in sorted order.
type SortableScanner struct {
	Scanner resource.Supplier
}

// NewSortableScanner creates a SortableScanner wrapping the given supplier.
func NewSortableScanner(scanner resource.Supplier) *SortableScanner {
	return &SortableScanner{
		Scanner: scanner,
	}
}

// Resources returns sorted resources from the wrapped scanner.
func (s *SortableScanner) Resources() ([]*resource.Resource, error) {
	resources, err := s.Scanner.Resources()
	if err != nil {
		return nil, err
	}
	return resource.Sort(resources), nil
}
