// Package hcl parses Terraform HCL configuration for backend and cloud block discovery.
package hcl

import (
	"path"

	"github.com/hashicorp/hcl/v2"
	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/snyk/driftctl/pkg/iac/terraform/state"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend"
)

// BackendBlock represents a Terraform backend configuration block.
type BackendBlock struct {
	Name               string   `hcl:"name,label"`
	Path               string   `hcl:"path,optional"`
	WorkspaceDir       string   `hcl:"workspace_dir,optional"`
	Bucket             string   `hcl:"bucket,optional"`
	Key                string   `hcl:"key,optional"`
	Region             string   `hcl:"region,optional"`
	Prefix             string   `hcl:"prefix,optional"`
	ContainerName      string   `hcl:"container_name,optional"`
	WorkspaceKeyPrefix string   `hcl:"workspace_key_prefix,optional"`
	Remain             hcl.Body `hcl:",remain"`
}

// SupplierConfig converts the backend block to a supplier config.
func (b BackendBlock) SupplierConfig(workspace string) *config.SupplierConfig {
	switch b.Name {
	case "local":
		return b.parseLocalBackend()
	case "s3":
		return b.parseS3Backend(workspace)
	}
	return nil
}

func (b BackendBlock) parseLocalBackend() *config.SupplierConfig {
	if b.Path == "" {
		return nil
	}
	return &config.SupplierConfig{
		Key:     state.TerraformStateReaderSupplier,
		Backend: backend.BackendKeyFile,
		Path:    path.Join(b.WorkspaceDir, b.Path),
	}
}

func (b BackendBlock) parseS3Backend(ws string) *config.SupplierConfig {
	if b.Bucket == "" || b.Key == "" {
		return nil
	}

	keyPrefix := b.WorkspaceKeyPrefix
	if ws != DefaultStateName {
		if b.WorkspaceKeyPrefix == "" {
			b.WorkspaceKeyPrefix = "env:"
		}
		keyPrefix = path.Join(b.WorkspaceKeyPrefix, ws)
	}

	return &config.SupplierConfig{
		Key:     state.TerraformStateReaderSupplier,
		Backend: backend.BackendKeyS3,
		Path:    path.Join(b.Bucket, keyPrefix, b.Key),
	}
}
