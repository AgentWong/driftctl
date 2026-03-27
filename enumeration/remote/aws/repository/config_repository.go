package repository

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/configservice/configserviceiface"
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
	client configserviceiface.ConfigServiceAPI
	cache  cache.Cache
}

func NewConfigRepository(session *session.Session, c cache.Cache) *configRepository {
	return &configRepository{
		configservice.New(session),
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
	err := r.client.GetDiscoveredResourceCountsPages(input, func(resp *configservice.GetDiscoveredResourceCountsOutput, lastPage bool) bool {
		for _, rc := range resp.ResourceCounts {
			if rc.ResourceType != nil {
				resourceTypes = append(resourceTypes, *rc.ResourceType)
			}
		}
		return !lastPage
	})
	if err != nil {
		return nil, err
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
		ResourceType:           aws.String(resourceType),
		IncludeDeletedResources: aws.Bool(false),
	}

	err := r.client.ListDiscoveredResourcesPages(input, func(resp *configservice.ListDiscoveredResourcesOutput, lastPage bool) bool {
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
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	return resources, nil
}
