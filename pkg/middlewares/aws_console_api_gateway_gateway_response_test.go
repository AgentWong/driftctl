package middlewares

import (
	"fmt"
	"strings"
	"testing"

	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsConsoleApiGatewayGatewayResponse_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			name: "console rest api gateway response is not ignored when managed by IaC",
			remoteResources: []*resource.Resource{
				{
					Id:   "rest-api",
					Type: aws.AwsAPIGatewayRestAPIResourceType,
				},
				{
					Id:   "gtw-response",
					Type: aws.AwsAPIGatewayGatewayResponseResourceType,
				},
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "rest-api",
					Type: aws.AwsAPIGatewayRestAPIResourceType,
				},
				{
					Id:   "gtw-response",
					Type: aws.AwsAPIGatewayGatewayResponseResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "rest-api",
					Type: aws.AwsAPIGatewayRestAPIResourceType,
				},
				{
					Id:   "gtw-response",
					Type: aws.AwsAPIGatewayGatewayResponseResourceType,
				},
			},
		},
		{
			name: "console rest api gateway response is ignored when not managed by IaC",
			remoteResources: []*resource.Resource{
				{
					Id:   "rest-api",
					Type: aws.AwsAPIGatewayRestAPIResourceType,
				},
				{
					Id:   "gtw-response",
					Type: aws.AwsAPIGatewayGatewayResponseResourceType,
				},
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "rest-api",
					Type: aws.AwsAPIGatewayRestAPIResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "rest-api",
					Type: aws.AwsAPIGatewayRestAPIResourceType,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsConsoleAPIGatewayGatewayResponse()
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
