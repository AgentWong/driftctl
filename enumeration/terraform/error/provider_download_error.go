// Package error defines errors returned during Terraform provider operations.
package error

import "fmt"

// ProviderNotFoundError indicates that a requested provider version does not exist.
type ProviderNotFoundError struct {
	Version string
}

func (p ProviderNotFoundError) Error() string {
	return fmt.Sprintf("Provider version %s does not exist", p.Version)
}
