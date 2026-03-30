package resource

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

// Factory creates abstract resources from raw data.
type Factory interface {
	CreateAbstractResource(ty, id string, data map[string]interface{}) *resource.Resource
}

// DriftctlResourceFactory implements Factory using a schema repository.
type DriftctlResourceFactory struct {
	resourceSchemaRepository SchemaRepositoryInterface
}

// NewDriftctlResourceFactory creates a DriftctlResourceFactory.
func NewDriftctlResourceFactory(resourceSchemaRepository SchemaRepositoryInterface) *DriftctlResourceFactory {
	return &DriftctlResourceFactory{
		resourceSchemaRepository: resourceSchemaRepository,
	}
}

// CreateAbstractResource builds a resource from the given type, id, and attributes.
func (r *DriftctlResourceFactory) CreateAbstractResource(ty, id string, data map[string]interface{}) *resource.Resource {
	attributes := resource.Attributes(data)
	attributes.SanitizeDefaults()

	schema, _ := r.resourceSchemaRepository.GetSchema(ty)
	res := resource.Resource{
		Id:    id,
		Type:  ty,
		Attrs: &attributes,
		Sch:   schema,
	}

	schema, exist := r.resourceSchemaRepository.GetSchema(ty)
	if exist && schema.NormalizeFunc != nil {
		schema.NormalizeFunc(&res)
	}

	return &res
}
