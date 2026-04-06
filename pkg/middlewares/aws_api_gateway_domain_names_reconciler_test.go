package middlewares

import (
	"fmt"
	"strings"
	"testing"

	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsApiGatewayDomainNamesReconciler_Execute(t *testing.T) {
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
					ID:   "domain1",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
				{
					ID:   "domain2",
					Type: aws.AwsAPIGatewayV2DomainNameResourceType,
				},
			},
			remoteResources: []*resource.Resource{
				{
					ID:   "domain1",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
				{
					ID:   "domain1",
					Type: aws.AwsAPIGatewayV2DomainNameResourceType,
				},
				{
					ID:   "domain2",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
				{
					ID:   "domain2",
					Type: aws.AwsAPIGatewayV2DomainNameResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					ID:   "domain1",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
				{
					ID:   "domain2",
					Type: aws.AwsAPIGatewayV2DomainNameResourceType,
				},
			},
		},
		{
			name:               "with unmanaged resources",
			resourcesFromState: []*resource.Resource{},
			remoteResources: []*resource.Resource{
				{
					ID:   "domain1",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
				{
					ID:   "domain1",
					Type: aws.AwsAPIGatewayV2DomainNameResourceType,
				},
				{
					ID:   "domain2",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
				{
					ID:   "domain2",
					Type: aws.AwsAPIGatewayV2DomainNameResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					ID:   "domain1",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
				{
					ID:   "domain2",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
			},
		},
		{
			name: "with deleted resources",
			resourcesFromState: []*resource.Resource{
				{
					ID:   "domain1",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
				{
					ID:   "domain2",
					Type: aws.AwsAPIGatewayV2DomainNameResourceType,
				},
			},
			remoteResources: []*resource.Resource{},
			expected:        []*resource.Resource{},
		},
		{
			name: "with a mix of managed, unmanaged and deleted resources",
			resourcesFromState: []*resource.Resource{
				{
					ID:   "domain1",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
				{
					ID:   "domain2",
					Type: aws.AwsAPIGatewayV2DomainNameResourceType,
				},
				{
					ID:   "domain4",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
			},
			remoteResources: []*resource.Resource{
				{
					ID:   "domain1",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
				{
					ID:   "domain1",
					Type: aws.AwsAPIGatewayV2DomainNameResourceType,
				},
				{
					ID:   "domain2",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
				{
					ID:   "domain2",
					Type: aws.AwsAPIGatewayV2DomainNameResourceType,
				},
				{
					ID:   "domain3",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
				{
					ID:   "domain3",
					Type: aws.AwsAPIGatewayV2DomainNameResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					ID:   "domain1",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
				{
					ID:   "domain2",
					Type: aws.AwsAPIGatewayV2DomainNameResourceType,
				},
				{
					ID:   "domain3",
					Type: aws.AwsAPIGatewayDomainNameResourceType,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsAPIGatewayDomainNamesReconciler()
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
