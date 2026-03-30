package middlewares

import (
	"fmt"
	"strings"
	"testing"

	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsDefaultApiGatewayAccount_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			"test that default account is not ignored when managed by IaC",
			[]*resource.Resource{
				{
					ID: "fake",
				},
				{
					ID:    "a-dummy-account",
					Type:  aws.AwsAPIGatewayAccountResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "default-managed-by-IaC",
					Type:  aws.AwsAPIGatewayAccountResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			[]*resource.Resource{
				{
					ID:    "default-managed-by-IaC",
					Type:  aws.AwsAPIGatewayAccountResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			[]*resource.Resource{
				{
					ID: "fake",
				},
				{
					ID:    "default-managed-by-IaC",
					Type:  aws.AwsAPIGatewayAccountResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			"test that default account is ignored when not managed by IaC",
			[]*resource.Resource{
				{
					ID: "fake",
				},
				{
					ID:    "a-dummy-account",
					Type:  aws.AwsAPIGatewayAccountResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "default-managed-by-IaC",
					Type:  aws.AwsAPIGatewayAccountResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			[]*resource.Resource{},
			[]*resource.Resource{
				{
					ID: "fake",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsDefaultAPIGatewayAccount()
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
