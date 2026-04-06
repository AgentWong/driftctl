package terraform

import tfproviders "github.com/hashicorp/terraform/providers"

// SchemaSupplier provides Terraform resource schemas.
type SchemaSupplier interface {
	Schema() map[string]tfproviders.Schema
}
