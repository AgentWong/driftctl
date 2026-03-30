package middlewares

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsAPIGatewayBasePathMappingReconciler is used to reconcile API Gateway base path mapping (v1 and v2)
// from both remote and state resources because v1|v2 AWS SDK list endpoints return all mappings
// AwsAPIGatewayBasePathMappingReconciler without distinction.
type AwsAPIGatewayBasePathMappingReconciler struct{}

// NewAwsAPIGatewayBasePathMappingReconciler creates a AwsAPIGatewayBasePathMappingReconciler.
func NewAwsAPIGatewayBasePathMappingReconciler() AwsAPIGatewayBasePathMappingReconciler {
	return AwsAPIGatewayBasePathMappingReconciler{}
}

// Execute applies the AwsAPIGatewayBasePathMappingReconciler middleware.
func (m AwsAPIGatewayBasePathMappingReconciler) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newRemoteResources := make([]*resource.Resource, 0)
	managedAPIMapping := make([]*resource.Resource, 0)
	unmanagedAPIMapping := make([]*resource.Resource, 0)
	for _, res := range *remoteResources {
		// Ignore all resources other than aws_api_gateway_base_path_mapping and aws_apigatewayv2_api_mapping
		if res.ResourceType() != aws.AwsAPIGatewayBasePathMappingResourceType &&
			res.ResourceType() != aws.AwsAPIGatewayV2MappingResourceType {
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
			managedAPIMapping = append(managedAPIMapping, res)
			continue
		}

		// If we're here, it means that we are left with unmanaged path mappings
		// in both v1 and v2 format. Let's group real and duplicate path mappings
		// in a slice
		unmanagedAPIMapping = append(unmanagedAPIMapping, res)
	}

	// We only want to show to our end users unmanaged v1 path mappings
	// To do that, we're going to loop over unmanaged path mappings to delete duplicates
	// and leave after that only v1 path mappings (e.g. remove v2 ones)
	deduplicatedUnmanagedMappings := make([]*resource.Resource, 0, len(unmanagedAPIMapping))
	for _, unmanaged := range unmanagedAPIMapping {
		// Remove duplicates (e.g. same id, the opposite type)
		isDuplicate := false
		for _, managed := range managedAPIMapping {
			if managed.ResourceID() == unmanaged.ResourceID() {
				isDuplicate = true
				break
			}
		}
		if isDuplicate {
			continue
		}

		// Now keep only v1 path mappings
		if unmanaged.ResourceType() == aws.AwsAPIGatewayBasePathMappingResourceType {
			deduplicatedUnmanagedMappings = append(deduplicatedUnmanagedMappings, unmanaged)
		}
	}

	// Finally, add both managed and unmanaged resources to remote resources
	newRemoteResources = append(newRemoteResources, managedAPIMapping...)
	newRemoteResources = append(newRemoteResources, deduplicatedUnmanagedMappings...)

	*remoteResources = newRemoteResources
	return nil
}
