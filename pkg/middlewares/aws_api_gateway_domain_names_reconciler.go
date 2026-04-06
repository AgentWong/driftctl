package middlewares

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsAPIGatewayDomainNamesReconciler Used to reconcile API Gateway domain names (v1 and v2) from both remote
// and state resources because v1|v2 AWS SDK list endpoints return all domain
// AwsAPIGatewayDomainNamesReconciler names without distinction
type AwsAPIGatewayDomainNamesReconciler struct{}

// NewAwsAPIGatewayDomainNamesReconciler creates a AwsAPIGatewayDomainNamesReconciler.
func NewAwsAPIGatewayDomainNamesReconciler() AwsAPIGatewayDomainNamesReconciler {
	return AwsAPIGatewayDomainNamesReconciler{}
}

// Execute applies the AwsAPIGatewayDomainNamesReconciler middleware.
func (m AwsAPIGatewayDomainNamesReconciler) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newRemoteResources := make([]*resource.Resource, 0)
	managedDomainNames := make([]*resource.Resource, 0)
	unmanagedDomainNames := make([]*resource.Resource, 0)
	for _, res := range *remoteResources {
		// Ignore all resources other than aws_api_gateway_domain_name and aws_apigatewayv2_domain_name
		if res.ResourceType() != aws.AwsAPIGatewayDomainNameResourceType &&
			res.ResourceType() != aws.AwsAPIGatewayV2DomainNameResourceType {
			newRemoteResources = append(newRemoteResources, res)
			continue
		}

		// Find a matching state resources
		existInState := false
		for _, stateResource := range *resourcesFromState {
			if res.Equal(stateResource) {
				existInState = true
				break
			}
		}

		// Keep track of the resource if it's managed in IaC
		if existInState {
			managedDomainNames = append(managedDomainNames, res)
			continue
		}

		// If we're here, it means that we are left with unmanaged domain names
		// in both v1 and v2 format. Let's group real and duplicate domain names
		// in a slice
		unmanagedDomainNames = append(unmanagedDomainNames, res)
	}

	// We only want to show to our end users unmanaged v1 domain names
	// To do that, we're going to loop over unmanaged domain names to delete duplicates
	// and leave after that only v1 domain names (e.g. remove v2 ones)
	deduplicatedUnmanagedDomains := make([]*resource.Resource, 0, len(unmanagedDomainNames))
	for _, unmanaged := range unmanagedDomainNames {
		// Remove duplicates (e.g. same id, the opposite type)
		isDuplicate := false
		for _, managed := range managedDomainNames {
			if managed.ResourceID() == unmanaged.ResourceID() {
				isDuplicate = true
				break
			}
		}
		if isDuplicate {
			continue
		}

		// Now keep only v1 domain names
		if unmanaged.ResourceType() == aws.AwsAPIGatewayDomainNameResourceType {
			deduplicatedUnmanagedDomains = append(deduplicatedUnmanagedDomains, unmanaged)
		}
	}

	// Finally, add both managed and unmanaged resources to remote resources
	newRemoteResources = append(newRemoteResources, managedDomainNames...)
	newRemoteResources = append(newRemoteResources, deduplicatedUnmanagedDomains...)

	*remoteResources = newRemoteResources
	return nil
}
