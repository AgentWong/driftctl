package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsAPIGatewayResourceExpander explodes api gateway default resource found in aws_api_gateway_rest_api.root_resource_id from state resources to dedicated resources
type AwsAPIGatewayResourceExpander struct {
	resourceFactory resource.Factory
}

// NewAwsAPIGatewayResourceExpander creates a AwsAPIGatewayResourceExpander.
func NewAwsAPIGatewayResourceExpander(resourceFactory resource.Factory) AwsAPIGatewayResourceExpander {
	return AwsAPIGatewayResourceExpander{
		resourceFactory: resourceFactory,
	}
}

// Execute applies the AwsAPIGatewayResourceExpander middleware.
func (m AwsAPIGatewayResourceExpander) Execute(_, resourcesFromState *[]*resource.Resource) error {
	newStateResources := make([]*resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than aws_api_gateway_rest_api
		if res.ResourceType() != aws.AwsAPIGatewayRestAPIResourceType {
			newStateResources = append(newStateResources, res)
			continue
		}

		newStateResources = append(newStateResources, res)

		err := m.handleResource(res, &newStateResources)
		if err != nil {
			return err
		}
	}
	*resourcesFromState = newStateResources
	return nil
}

func (m *AwsAPIGatewayResourceExpander) handleResource(api *resource.Resource, results *[]*resource.Resource) error {
	resourceID := api.Attrs.GetString("root_resource_id")
	if resourceID == nil || *resourceID == "" {
		return nil
	}

	newResource := m.resourceFactory.CreateAbstractResource(aws.AwsAPIGatewayResourceResourceType, *resourceID, map[string]interface{}{
		"rest_api_id": api.ResourceId(),
		"path":        "/",
	})

	*results = append(*results, newResource)
	logrus.WithFields(logrus.Fields{
		"id": newResource.ResourceId(),
	}).Debug("Created new resource from api gateway rest api")

	return nil
}
