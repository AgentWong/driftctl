// Package test_tfe provides mock-friendly interfaces for Terraform Enterprise.
package test_tfe

import "github.com/hashicorp/go-tfe"

// Workspaces extends tfe.Workspaces for mock generation.
type Workspaces interface {
	tfe.Workspaces
}

// StateVersions extends tfe.StateVersions for mock generation.
type StateVersions interface {
	tfe.StateVersions
}
