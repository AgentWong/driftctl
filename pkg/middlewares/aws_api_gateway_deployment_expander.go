package middlewares

import (
	"strings"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsAPIGatewayDeploymentExpander Create a aws_api_gateway_stage resource from a aws_api_gateway_deployment resource and ignore the latter resource
// AwsAPIGatewayDeploymentExpander since we don't support it
type AwsAPIGatewayDeploymentExpander struct {
	resourceFactory resource.Factory
}

// NewAwsAPIGatewayDeploymentExpander creates a AwsAPIGatewayDeploymentExpander.
func NewAwsAPIGatewayDeploymentExpander(resourceFactory resource.Factory) AwsAPIGatewayDeploymentExpander {
	return AwsAPIGatewayDeploymentExpander{resourceFactory}
}

// Execute applies the AwsAPIGatewayDeploymentExpander middleware.
func (m AwsAPIGatewayDeploymentExpander) Execute(_, resourcesFromState *[]*resource.Resource) error {
	var newResources []*resource.Resource
	for _, res := range *resourcesFromState {
		if res.ResourceType() != aws.AwsAPIGatewayDeploymentResourceType {
			newResources = append(newResources, res)
			continue
		}

		stageName := res.Attributes().GetString("stage_name")
		if stageName == nil || *stageName == "" {
			continue
		}

		newStage := m.resourceFactory.CreateAbstractResource(
			aws.AwsAPIGatewayStageResourceType,
			strings.Join([]string{"ags", *(res.Attributes().GetString("rest_api_id")), *stageName}, "-"),
			map[string]interface{}{},
		)

		newResources = append(newResources, newStage)
	}
	*resourcesFromState = newResources

	return nil
}
