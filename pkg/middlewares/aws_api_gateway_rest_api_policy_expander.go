package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsAPIGatewayRestAPIPolicyExpander explodes policy found in aws_api_gateway_rest_api.policy from state resources to dedicated resources
type AwsAPIGatewayRestAPIPolicyExpander struct {
	resourceFactory resource.Factory
}

// NewAwsAPIGatewayRestAPIPolicyExpander creates a AwsAPIGatewayRestAPIPolicyExpander.
func NewAwsAPIGatewayRestAPIPolicyExpander(resourceFactory resource.Factory) AwsAPIGatewayRestAPIPolicyExpander {
	return AwsAPIGatewayRestAPIPolicyExpander{
		resourceFactory: resourceFactory,
	}
}

// Execute applies the AwsAPIGatewayRestAPIPolicyExpander middleware.
func (m AwsAPIGatewayRestAPIPolicyExpander) Execute(_, resourcesFromState *[]*resource.Resource) error {
	newList := make([]*resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than api_gateway_rest_api
		if res.ResourceType() != aws.AwsAPIGatewayRestAPIResourceType {
			newList = append(newList, res)
			continue
		}

		newList = append(newList, res)

		if hasRestAPIPolicyAttached(res.ResourceId(), resourcesFromState) {
			res.Attrs.SafeDelete([]string{"policy"})
			continue
		}

		err := m.handlePolicy(res, &newList)
		if err != nil {
			return err
		}
	}
	*resourcesFromState = newList
	return nil
}

func (m *AwsAPIGatewayRestAPIPolicyExpander) handlePolicy(api *resource.Resource, results *[]*resource.Resource) error {
	policy, exist := api.Attrs.Get("policy")
	if !exist || policy == nil || policy == "" {
		return nil
	}

	data := map[string]interface{}{
		"id":          api.ResourceId(),
		"rest_api_id": api.ResourceId(),
		"policy":      policy,
	}

	newPolicy := m.resourceFactory.CreateAbstractResource(aws.AwsAPIGatewayRestAPIPolicyResourceType, api.ResourceId(), data)

	*results = append(*results, newPolicy)
	logrus.WithFields(logrus.Fields{
		"id": newPolicy.ResourceId(),
	}).Debug("Created new policy from api gateway rest api")

	api.Attrs.SafeDelete([]string{"policy"})
	return nil
}

// Return true if the rest api has a aws_api_gateway_rest_api_policy resource attached to itself.
// It is mandatory since it's possible to have a aws_api_gateway_rest_api with an inline policy
// AND a aws_api_gateway_rest_api_policy resource at the same time. At the end, on the AWS console,
// hasRestAPIPolicyAttached the aws_api_gateway_rest_api_policy will be used.
func hasRestAPIPolicyAttached(api string, resourcesFromState *[]*resource.Resource) bool {
	for _, res := range *resourcesFromState {
		if res.ResourceType() == aws.AwsAPIGatewayRestAPIPolicyResourceType &&
			res.ResourceId() == api {
			return true
		}
	}
	return false
}
