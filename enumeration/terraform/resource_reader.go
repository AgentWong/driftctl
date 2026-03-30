package terraform

import (
	"github.com/snyk/driftctl/enumeration/resource"

	"github.com/zclconf/go-cty/cty"
)

// ResourceReader reads individual resources from a Terraform provider.
type ResourceReader interface {
	ReadResource(args ReadResourceArgs) (*cty.Value, error)
}

// ReadResourceArgs holds the arguments for reading a single resource.
type ReadResourceArgs struct {
	Ty         resource.Type
	ID         string
	Attributes map[string]string
}
