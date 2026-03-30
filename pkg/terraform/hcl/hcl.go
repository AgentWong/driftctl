package hcl

import (
	"os"
	"path"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// DefaultStateName is the default Terraform workspace name.
const DefaultStateName = "default"

// MainBodyBlock is the top-level HCL body containing terraform configuration.
type MainBodyBlock struct {
	Terraform TerraformBlock `hcl:"terraform,block"`
	Remain    hcl.Body       `hcl:",remain"`
}

// TerraformBlock holds the terraform block with optional backend and cloud sub-blocks.
type TerraformBlock struct {
	Backend *BackendBlock `hcl:"backend,block"`
	Cloud   *CloudBlock   `hcl:"cloud,block"`
	Remain  hcl.Body      `hcl:",remain"`
}

// ParseTerraformFromHCL parses a Terraform HCL file and returns the terraform block.
func ParseTerraformFromHCL(filename string) (*TerraformBlock, error) {
	var body MainBodyBlock

	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, diags
	}

	diags = gohcl.DecodeBody(f.Body, nil, &body)
	if diags.HasErrors() {
		return nil, diags
	}

	return &body.Terraform, nil
}

// GetCurrentWorkspaceName reads the active Terraform workspace from the .terraform/environment file.
func GetCurrentWorkspaceName(cwd string) string {
	name := DefaultStateName // See https://github.com/hashicorp/terraform/blob/main/internal/backend/backend.go#L33

	data, err := os.ReadFile(path.Join(cwd, ".terraform/environment")) //nolint:gosec // G304: cwd is from the working directory
	if err != nil {
		return name
	}
	if v := strings.Trim(string(data), "\n"); v != "" {
		name = v
	}
	return name
}
