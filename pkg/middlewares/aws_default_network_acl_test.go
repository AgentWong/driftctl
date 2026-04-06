package middlewares

import (
	"fmt"
	"strings"
	"testing"

	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsDefaultNetworkACL_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			"default network ACL is not ignored when managed by IaC",
			[]*resource.Resource{
				{
					ID: "fake",
				},
				{
					ID:   "default-acl",
					Type: aws.AwsDefaultNetworkACLResourceType,
				},
				{
					ID:   "non-default-acl",
					Type: aws.AwsNetworkACLResourceType,
				},
			},
			[]*resource.Resource{
				{
					ID:   "default-acl",
					Type: aws.AwsDefaultNetworkACLResourceType,
				},
			},
			[]*resource.Resource{
				{
					ID: "fake",
				},
				{
					ID:   "default-acl",
					Type: aws.AwsDefaultNetworkACLResourceType,
				},
				{
					ID:   "non-default-acl",
					Type: aws.AwsNetworkACLResourceType,
				},
			},
		},
		{
			"default network acl is ignored when not managed by IaC",
			[]*resource.Resource{
				{
					ID: "fake",
				},
				{
					ID:   "default-acl",
					Type: aws.AwsDefaultNetworkACLResourceType,
				},
				{
					ID:   "non-default-acl",
					Type: aws.AwsNetworkACLResourceType,
				},
			},
			[]*resource.Resource{},
			[]*resource.Resource{
				{
					ID: "fake",
				},
				{
					ID:   "non-default-acl",
					Type: aws.AwsNetworkACLResourceType,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsDefaultNetworkACL()
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
