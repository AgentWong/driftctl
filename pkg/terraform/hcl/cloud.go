package hcl

import (
	"path"

	"github.com/hashicorp/hcl/v2"
	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/snyk/driftctl/pkg/iac/terraform/state"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend"
)

// CloudWorkspacesBlock holds the workspaces configuration within a cloud block.
type CloudWorkspacesBlock struct {
	Name   string   `hcl:"name,optional"`
	Tags   []string `hcl:"tags,optional"`
	Remain hcl.Body `hcl:",remain"`
}

// CloudBlock represents a Terraform Cloud configuration block.
type CloudBlock struct {
	Organization string               `hcl:"organization"`
	Workspaces   CloudWorkspacesBlock `hcl:"workspaces,block"`
	Remain       hcl.Body             `hcl:",remain"`
}

// SupplierConfig converts the cloud block to a supplier config.
func (c CloudBlock) SupplierConfig(workspace string) *config.SupplierConfig {
	// If a workspace is specified in HCL, use it rather than the current environment
	if c.Workspaces.Name != "" {
		workspace = c.Workspaces.Name
	}
	return &config.SupplierConfig{
		Key:     state.TerraformStateReaderSupplier,
		Backend: backend.BackendKeyTFCloud,
		Path:    path.Join(c.Organization, workspace),
	}
}
