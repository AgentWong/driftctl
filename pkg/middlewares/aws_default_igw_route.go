package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsDefaultInternetGatewayRoute Each region has a default vpc which has an internet gateway attached and thus the route table of this
// same vpc has a default route (0.0.0.0/0) that should not be seen as unmanaged if not managed by IaC
// AwsDefaultInternetGatewayRoute this middleware ignores the above route from unmanaged resources if not managed by IaC
type AwsDefaultInternetGatewayRoute struct{}

// NewAwsDefaultInternetGatewayRoute creates a AwsDefaultInternetGatewayRoute.
func NewAwsDefaultInternetGatewayRoute() AwsDefaultInternetGatewayRoute {
	return AwsDefaultInternetGatewayRoute{}
}

// Execute applies the AwsDefaultInternetGatewayRoute middleware.
func (m AwsDefaultInternetGatewayRoute) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newRemoteResources := make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than routes
		if remoteResource.ResourceType() != aws.AwsRouteResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Ignore all routes except the one that came from the default internet gateway
		if !isDefaultInternetGatewayRoute(remoteResource, remoteResources) {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Check if route is managed by IaC
		existInState := false
		for _, stateResource := range *resourcesFromState {
			if remoteResource.Equal(stateResource) {
				existInState = true
				break
			}
		}

		// Include resource if it's managed in IaC
		if existInState {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Else, resource is not added to newRemoteResources slice so it will be ignored
		logrus.WithFields(logrus.Fields{
			"id":   remoteResource.ResourceID(),
			"type": remoteResource.ResourceType(),
		}).Debug("Ignoring default internet gateway route as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

// isDefaultInternetGatewayRoute return true if the route's target is the default internet gateway (e.g. attached to the default vpc)
func isDefaultInternetGatewayRoute(route *resource.Resource, remoteResources *[]*resource.Resource) bool {
	for _, remoteResource := range *remoteResources {
		if remoteResource.ResourceType() == aws.AwsInternetGatewayResourceType &&
			isDefaultInternetGateway(remoteResource, remoteResources) {
			gtwID, gtwIDExist := route.Attrs.Get("gateway_id")
			destCIDRBlock, destCIDRBlockExist := route.Attrs.Get("destination_cidr_block")
			return gtwIDExist && destCIDRBlockExist && gtwID == remoteResource.ResourceID() && destCIDRBlock == "0.0.0.0/0"
		}
	}
	return false
}
