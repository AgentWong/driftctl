package middlewares

import (
	"fmt"
	"strings"
	"testing"

	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsApiGatewayDeploymentExpander_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		mocks              func(*dctlresource.MockResourceFactory)
		expected           []*resource.Resource
	}{
		{
			name: "no stages created from deployment state resources",
			resourcesFromState: []*resource.Resource{
				{
					ID:   "foo",
					Type: aws.AwsAPIGatewayDeploymentResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "api",
					},
				},
				{
					ID:   "bar",
					Type: aws.AwsAPIGatewayDeploymentResourceType,
					Attrs: &resource.Attributes{
						"stage_name":  "",
						"rest_api_id": "api",
					},
				},
				{
					ID:   "ags-api-baz",
					Type: aws.AwsAPIGatewayStageResourceType,
					Attrs: &resource.Attributes{
						"stage_name": "baz",
					},
				},
			},
			expected: []*resource.Resource{
				{
					ID:   "ags-api-baz",
					Type: aws.AwsAPIGatewayStageResourceType,
					Attrs: &resource.Attributes{
						"stage_name": "baz",
					},
				},
			},
		},
		{
			name: "stages created from deployment state resources",
			mocks: func(factory *dctlresource.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsAPIGatewayStageResourceType,
					"ags-api-foo",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					ID:   "ags-api-foo",
					Type: aws.AwsAPIGatewayStageResourceType,
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					ID:   "foo",
					Type: aws.AwsAPIGatewayDeploymentResourceType,
					Attrs: &resource.Attributes{
						"stage_name":  "foo",
						"rest_api_id": "api",
					},
				},
				{
					ID:   "ags-api-baz",
					Type: aws.AwsAPIGatewayStageResourceType,
					Attrs: &resource.Attributes{
						"stage_name": "baz",
					},
				},
			},
			expected: []*resource.Resource{
				{
					ID:   "ags-api-baz",
					Type: aws.AwsAPIGatewayStageResourceType,
					Attrs: &resource.Attributes{
						"stage_name": "baz",
					},
				},
				{
					ID:   "ags-api-foo",
					Type: aws.AwsAPIGatewayStageResourceType,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &dctlresource.MockResourceFactory{}
			if tt.mocks != nil {
				tt.mocks(factory)
			}

			m := NewAwsAPIGatewayDeploymentExpander(factory)
			err := m.Execute(&[]*resource.Resource{}, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}
			changelog, err := diff.Diff(tt.expected, tt.resourcesFromState)
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
