// Package iac provides IaC (Infrastructure as Code) state reading and supplier logic.
package iac

import (
	"fmt"
	"strings"
)

// StateReadingError aggregates multiple errors encountered while reading state files.
type StateReadingError struct {
	errors []error
}

// NewStateReadingError returns an empty StateReadingError.
func NewStateReadingError() *StateReadingError {
	return &StateReadingError{}
}

// Add appends an error to the StateReadingError collection.
func (s *StateReadingError) Add(err error) {
	s.errors = append(s.errors, err)
}

func (s *StateReadingError) Error() string {
	var err strings.Builder
	_, _ = fmt.Fprint(&err, "There were errors reading your states files : \n")
	for _, e := range s.errors {
		_, _ = fmt.Fprintf(&err, "   - %s\n", e.Error())
	}
	return err.String()
}
