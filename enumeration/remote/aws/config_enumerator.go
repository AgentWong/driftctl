// Package aws implements the AWS remote provider for resource enumeration.
package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
)

// ConfigEnumerator enumerates AWS Config resources via the Config API.
type ConfigEnumerator struct {
	repo    repository.ConfigRepository
	mapping map[string]string
	factory resource.Factory
}

// NewConfigEnumerator creates a ConfigEnumerator backed by the given repository.
func NewConfigEnumerator(repo repository.ConfigRepository, factory resource.Factory) *ConfigEnumerator {
	return &ConfigEnumerator{
		repo:    repo,
		mapping: ConfigToTerraformMapping,
		factory: factory,
	}
}

// SupportedTypes returns the Terraform resource types this enumerator can discover.
func (e *ConfigEnumerator) SupportedTypes() []resource.Type {
	seen := make(map[string]struct{}, len(e.mapping))
	types := make([]resource.Type, 0, len(e.mapping))
	for _, tfType := range e.mapping {
		if _, dup := seen[tfType]; dup {
			continue
		}
		seen[tfType] = struct{}{}
		types = append(types, resource.Type(tfType))
	}
	return types
}

// Enumerate discovers resources from AWS Config and converts them to Terraform resources.
func (e *ConfigEnumerator) Enumerate(filter common.EnumerationFilter) ([]*resource.Resource, error) {
	// pass Config type keys so the repo queries those directly instead of
	// relying on GetDiscoveredResourceCounts (which lags on new recorders)
	configTypes := make([]string, 0, len(e.mapping))
	for ct := range e.mapping {
		configTypes = append(configTypes, ct)
	}
	discovered, err := e.repo.ListAllDiscoveredResources(configTypes)
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, "aws_config_discovered_resources")
	}

	var results []*resource.Resource
	for _, r := range discovered {
		tfType, ok := e.mapping[r.Type]
		if !ok {
			continue
		}

		if filter != nil && filter.IsTypeIgnored(resource.Type(tfType)) {
			continue
		}

		// seed resource attributes with data from Config so downstream
		// categorizers and output formatters have access to tags/ARN/name
		attrs := map[string]interface{}{}
		if len(r.Tags) > 0 {
			// store as map[string]interface{} so the CloudFormation categorizer
			// can type-assert against map[string]interface{}
			tagMap := make(map[string]interface{}, len(r.Tags))
			for k, v := range r.Tags {
				tagMap[k] = v
			}
			attrs["tags"] = tagMap
		}
		if r.ARN != "" {
			attrs["arn"] = r.ARN
		}
		if r.Name != "" {
			attrs["config_name"] = r.Name
		}

		res := e.factory.CreateAbstractResource(tfType, r.ID, attrs)

		if filter != nil && filter.IsResourceIgnored(res) {
			continue
		}

		results = append(results, res)
	}

	return results, nil
}
