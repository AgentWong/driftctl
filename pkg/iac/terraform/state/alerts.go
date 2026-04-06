// Package state reads and enumerates Terraform state resources.
package state

import (
	"fmt"

	"github.com/snyk/driftctl/enumeration/resource"
)

// ReadingAlert represents a ReadingAlert.
type ReadingAlert struct {
	key string
	err string
}

// NewReadingAlert creates a new instance.
func NewReadingAlert(key string, err error) *ReadingAlert {
	return &ReadingAlert{key: key, err: err.Error()}
}

// Message implements the ReadingAlert interface.
func (s *ReadingAlert) Message() string {
	return fmt.Sprintf("Your analysis may be incomplete. There was an error reading state file '%s': %s", s.key, s.err)
}

// ShouldIgnoreResource implements the ReadingAlert interface.
func (s *ReadingAlert) ShouldIgnoreResource() bool {
	return false
}

// Resource implements the ReadingAlert interface.
func (s *ReadingAlert) Resource() *resource.Resource {
	return nil
}
