// Package repository provides AWS service repository implementations for resource enumeration.
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/snyk/driftctl/enumeration/remote/cache"
)

// ConfigDiscoveredResource represents a single resource discovered via AWS Config advanced queries.
type ConfigDiscoveredResource struct {
	Type string
	ID   string
	Name string
	ARN  string
	Tags map[string]string
}

// ConfigRepository defines the interface for querying resources discovered by AWS Config.
type ConfigRepository interface {
	ListAllDiscoveredResources(resourceTypes []string) ([]*ConfigDiscoveredResource, error)
}

type configRepository struct {
	client *configservice.Client
	cache  cache.Cache
}

// NewConfigRepository creates a new ConfigRepository backed by an AWS Config service client.
func NewConfigRepository(cfg aws.Config, c cache.Cache) ConfigRepository {
	return &configRepository{
		configservice.NewFromConfig(cfg),
		c,
	}
}

// selectResourceResult maps the JSON returned by SelectResourceConfig.
type selectResourceResult struct {
	ResourceType string           `json:"resourceType"`
	ResourceID   string           `json:"resourceId"`
	ResourceName string           `json:"resourceName"`
	ARN          string           `json:"arn"`
	Tags         []selectTagEntry `json:"tags"`
}

// selectTagEntry represents a single tag from the Config query result.
type selectTagEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ListAllDiscoveredResources uses the SelectResourceConfig (Advanced Query)
// API, which queries Config's resource index directly and returns results
// immediately — unlike ListDiscoveredResources + GetDiscoveredResourceCounts,
// which lag on newly-started recorders.
func (r *configRepository) ListAllDiscoveredResources(resourceTypes []string) ([]*ConfigDiscoveredResource, error) {
	cacheKey := "configListAllDiscoveredResources"
	v := r.cache.GetAndLock(cacheKey)
	defer r.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*ConfigDiscoveredResource), nil
	}

	var allResources []*ConfigDiscoveredResource

	// chunk types to stay under the 4 KB SQL expression limit
	const chunkSize = 50
	for i := 0; i < len(resourceTypes); i += chunkSize {
		end := i + chunkSize
		if end > len(resourceTypes) {
			end = len(resourceTypes)
		}
		resources, err := r.selectResources(resourceTypes[i:end])
		if err != nil {
			return nil, err
		}
		allResources = append(allResources, resources...)
	}

	r.cache.Put(cacheKey, allResources)
	return allResources, nil
}

// selectResources runs a single SelectResourceConfig query for a batch of types.
func (r *configRepository) selectResources(resourceTypes []string) ([]*ConfigDiscoveredResource, error) {
	quoted := make([]string, len(resourceTypes))
	for i, rt := range resourceTypes {
		quoted[i] = fmt.Sprintf("'%s'", rt)
	}
	expr := fmt.Sprintf(
		"SELECT resourceId, resourceType, resourceName, arn, tags WHERE resourceType IN (%s)",
		strings.Join(quoted, ", "),
	)

	var resources []*ConfigDiscoveredResource
	paginator := configservice.NewSelectResourceConfigPaginator(r.client, &configservice.SelectResourceConfigInput{
		Expression: aws.String(expr),
	})
	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, raw := range resp.Results {
			var parsed selectResourceResult
			if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
				continue
			}
			// flatten tag list into a map for easier downstream consumption
			tags := make(map[string]string, len(parsed.Tags))
			for _, t := range parsed.Tags {
				tags[t.Key] = t.Value
			}
			resources = append(resources, &ConfigDiscoveredResource{
				Type: parsed.ResourceType,
				ID:   parsed.ResourceID,
				Name: parsed.ResourceName,
				ARN:  parsed.ARN,
				Tags: tags,
			})
		}
	}

	return resources, nil
}
