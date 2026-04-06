package enumeration

import (
	"github.com/hashicorp/terraform/terraform"
	"github.com/snyk/driftctl/enumeration/diagnostic"
	"github.com/snyk/driftctl/enumeration/resource"
)

// RefreshInput holds the resources to refresh.
type RefreshInput struct {
	// Resources to refresh
	Resources map[string][]*resource.Resource
}

// RefreshOutput holds the refreshed resources and any diagnostics.
type RefreshOutput struct {
	Resources   map[string][]*resource.Resource
	Diagnostics diagnostic.Diagnostics
}

// GetSchemasOutput holds the provider schema.
type GetSchemasOutput struct {
	Schema *terraform.ProviderSchema
}

// Refresher reads the current state of resources from the provider.
type Refresher interface {
	Refresh(input *RefreshInput) (*RefreshOutput, error)
	GetSchema() (*GetSchemasOutput, error)
}
