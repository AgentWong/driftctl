package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
)

type ConfigEnumerator struct {
	repo    repository.ConfigRepository
	mapping map[string]string
	factory resource.ResourceFactory
}

func NewConfigEnumerator(repo repository.ConfigRepository, factory resource.ResourceFactory) *ConfigEnumerator {
	return &ConfigEnumerator{
		repo:    repo,
		mapping: ConfigToTerraformMapping,
		factory: factory,
	}
}

func (e *ConfigEnumerator) SupportedTypes() []resource.ResourceType {
	seen := make(map[string]struct{}, len(e.mapping))
	types := make([]resource.ResourceType, 0, len(e.mapping))
	for _, tfType := range e.mapping {
		if _, dup := seen[tfType]; dup {
			continue
		}
		seen[tfType] = struct{}{}
		types = append(types, resource.ResourceType(tfType))
	}
	return types
}

func (e *ConfigEnumerator) Enumerate(filter common.EnumerationFilter) ([]*resource.Resource, error) {
	discovered, err := e.repo.ListAllDiscoveredResources()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, "aws_config_discovered_resources")
	}

	var results []*resource.Resource
	for _, r := range discovered {
		tfType, ok := e.mapping[r.Type]
		if !ok {
			continue
		}

		if filter != nil && filter.IsTypeIgnored(resource.ResourceType(tfType)) {
			continue
		}

		res := e.factory.CreateAbstractResource(tfType, r.ID, map[string]interface{}{})

		if filter != nil && filter.IsResourceIgnored(res) {
			continue
		}

		results = append(results, res)
	}

	return results, nil
}
