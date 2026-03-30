package middlewares

import (
	"fmt"
	"strings"
	"testing"

	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsApiGatewayBasePathMappingReconciler_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		remoteResources    []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			name: "with managed resources",
			resourcesFromState: []*resource.Resource{
				{
					ID:   "mapping1",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
				{
					ID:   "mapping2",
					Type: aws.AwsAPIGatewayV2MappingResourceType,
				},
			},
			remoteResources: []*resource.Resource{
				{
					ID:   "mapping1",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
				{
					ID:   "mapping1",
					Type: aws.AwsAPIGatewayV2MappingResourceType,
				},
				{
					ID:   "mapping2",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
				{
					ID:   "mapping2",
					Type: aws.AwsAPIGatewayV2MappingResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					ID:   "mapping1",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
				{
					ID:   "mapping2",
					Type: aws.AwsAPIGatewayV2MappingResourceType,
				},
			},
		},
		{
			name:               "with unmanaged resources",
			resourcesFromState: []*resource.Resource{},
			remoteResources: []*resource.Resource{
				{
					ID:   "mapping1",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
				{
					ID:   "mapping1",
					Type: aws.AwsAPIGatewayV2MappingResourceType,
				},
				{
					ID:   "mapping2",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
				{
					ID:   "mapping2",
					Type: aws.AwsAPIGatewayV2MappingResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					ID:   "mapping1",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
				{
					ID:   "mapping2",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
			},
		},
		{
			name: "with deleted resources",
			resourcesFromState: []*resource.Resource{
				{
					ID:   "mapping1",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
				{
					ID:   "mapping2",
					Type: aws.AwsAPIGatewayV2MappingResourceType,
				},
			},
			remoteResources: []*resource.Resource{},
			expected:        []*resource.Resource{},
		},
		{
			name: "with a mix of managed, unmanaged and deleted resources",
			resourcesFromState: []*resource.Resource{
				{
					ID:   "mapping1",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
				{
					ID:   "mapping2",
					Type: aws.AwsAPIGatewayV2MappingResourceType,
				},
				{
					ID:   "mapping4",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
			},
			remoteResources: []*resource.Resource{
				{
					ID:   "mapping1",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
				{
					ID:   "mapping1",
					Type: aws.AwsAPIGatewayV2MappingResourceType,
				},
				{
					ID:   "mapping2",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
				{
					ID:   "mapping2",
					Type: aws.AwsAPIGatewayV2MappingResourceType,
				},
				{
					ID:   "mapping3",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
				{
					ID:   "mapping3",
					Type: aws.AwsAPIGatewayV2MappingResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					ID:   "mapping1",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
				{
					ID:   "mapping2",
					Type: aws.AwsAPIGatewayV2MappingResourceType,
				},
				{
					ID:   "mapping3",
					Type: aws.AwsAPIGatewayBasePathMappingResourceType,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsAPIGatewayBasePathMappingReconciler()
			err := m.Execute(&tt.remoteResources, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}
			changelog, err := diff.Diff(tt.expected, tt.remoteResources)
			if err != nil {
				t.Fatal(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), fmt.Sprintf("%v", change.From), fmt.Sprintf("%v", change.To))
				}
			}
		})
	}
}
