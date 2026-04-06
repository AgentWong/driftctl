package middlewares

import (
	"fmt"
	"strings"
	"testing"

	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsDefaultInternetGateway_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			"default internet gateway is not ignored when managed by IaC",
			[]*resource.Resource{
				{
					ID: "fake",
				},
				{
					ID:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				{
					ID:   "dummy-vpc",
					Type: aws.AwsVpcResourceType,
				},
				{
					ID:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				{
					ID:   "dummy-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "dummy-vpc",
					},
				},
			},
			[]*resource.Resource{
				{
					ID:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
			},
			[]*resource.Resource{
				{
					ID: "fake",
				},
				{
					ID:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				{
					ID:   "dummy-vpc",
					Type: aws.AwsVpcResourceType,
				},
				{
					ID:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				{
					ID:   "dummy-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "dummy-vpc",
					},
				},
			},
		},
		{
			"default internet gateway is ignored when not managed by IaC",
			[]*resource.Resource{
				{
					ID: "fake",
				},
				{
					ID:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				{
					ID:   "dummy-vpc",
					Type: aws.AwsVpcResourceType,
				},
				{
					ID:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				{
					ID:   "dummy-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "dummy-vpc",
					},
				},
			},
			[]*resource.Resource{},
			[]*resource.Resource{
				{
					ID: "fake",
				},
				{
					ID:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				{
					ID:   "dummy-vpc",
					Type: aws.AwsVpcResourceType,
				},
				{
					ID:   "dummy-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "dummy-vpc",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsDefaultInternetGateway()
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
