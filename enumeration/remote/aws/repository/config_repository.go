package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/snyk/driftctl/enumeration/remote/cache"
)

type ConfigDiscoveredResource struct {
	Type string
	ID   string
	Name string
}

type ConfigRepository interface {
	ListAllDiscoveredResources() ([]*ConfigDiscoveredResource, error)
	GetSupportedResourceTypes() ([]string, error)
}

type configRepository struct {
	client *configservice.Client
	cache  cache.Cache
}

func NewConfigRepository(cfg aws.Config, c cache.Cache) *configRepository {
	return &configRepository{
		configservice.NewFromConfig(cfg),
		c,
	}
}

func (r *configRepository) GetSupportedResourceTypes() ([]string, error) {
	cacheKey := "configGetSupportedResourceTypes"
	v := r.cache.GetAndLock(cacheKey)
	defer r.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]string), nil
	}

	var resourceTypes []string
	input := &configservice.GetDiscoveredResourceCountsInput{}
	paginator := configservice.NewGetDiscoveredResourceCountsPaginator(r.client, input)
	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, rc := range resp.ResourceCounts {
			if rc.ResourceType != "" {
				resourceTypes = append(resourceTypes, string(rc.ResourceType))
			}
		}
	}

	r.cache.Put(cacheKey, resourceTypes)
	return resourceTypes, nil
}

func (r *configRepository) ListAllDiscoveredResources() ([]*ConfigDiscoveredResource, error) {
	cacheKey := "configListAllDiscoveredResources"
	v := r.cache.GetAndLock(cacheKey)
	defer r.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*ConfigDiscoveredResource), nil
	}

	resourceTypes, err := r.GetSupportedResourceTypes()
	if err != nil {
		return nil, err
	}

	var allResources []*ConfigDiscoveredResource
	for _, rt := range resourceTypes {
		resources, err := r.listResourcesByType(rt)
		if err != nil {
			return nil, err
		}
		allResources = append(allResources, resources...)
	}

	r.cache.Put(cacheKey, allResources)
	return allResources, nil
}

func (r *configRepository) listResourcesByType(resourceType string) ([]*ConfigDiscoveredResource, error) {
	var resources []*ConfigDiscoveredResource
	input := &configservice.ListDiscoveredResourcesInput{
		ResourceType:            types.ResourceType(resourceType),
		IncludeDeletedResources: false,
	}

	paginator := configservice.NewListDiscoveredResourcesPaginator(r.client, input)
	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, ri := range resp.ResourceIdentifiers {
			res := &ConfigDiscoveredResource{
				Type: resourceType,
			}
			if ri.ResourceId != nil {
				res.ID = *ri.ResourceId
			}
			if ri.ResourceName != nil {
				res.Name = *ri.ResourceName
			}
			resources = append(resources, res)
		}
	}

	return resources, nil
}
