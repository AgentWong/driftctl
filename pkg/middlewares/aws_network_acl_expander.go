package middlewares

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsNetworkACLExpander this middelware goal is to explode aws_network_acl ingress and egress block into a set of aws_network_acl_rule
type AwsNetworkACLExpander struct {
	resourceFactory resource.Factory
}

// NewAwsNetworkACLExpander creates a AwsNetworkACLExpander.
func NewAwsNetworkACLExpander(resourceFactory resource.Factory) AwsNetworkACLExpander {
	return AwsNetworkACLExpander{resourceFactory}
}

// Execute applies the AwsNetworkACLExpander middleware.
func (m AwsNetworkACLExpander) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newResourcesFromState := make([]*resource.Resource, 0, len(*resourcesFromState))

	for _, stateResource := range *resourcesFromState {
		// Ignore all resources other than network acl
		if stateResource.ResourceType() != aws.AwsNetworkACLResourceType &&
			stateResource.ResourceType() != aws.AwsDefaultNetworkACLResourceType {
			newResourcesFromState = append(newResourcesFromState, stateResource)
			continue
		}

		newResourcesFromState = append(newResourcesFromState, m.expandBlock(
			resourcesFromState,
			stateResource.ResourceId(),
			false,
			stateResource.Attrs.GetSlice("ingress"),
		)...)
		stateResource.Attrs.SafeDelete([]string{"ingress"})

		newResourcesFromState = append(newResourcesFromState, m.expandBlock(
			resourcesFromState,
			stateResource.ResourceId(),
			true,
			stateResource.Attrs.GetSlice("egress"),
		)...)
		stateResource.Attrs.SafeDelete([]string{"egress"})

		newResourcesFromState = append(newResourcesFromState, stateResource)
	}

	// Then we need to remove ingress and egress block from remote resource too
	newRemoteResources := make([]*resource.Resource, 0, len(*remoteResources))
	for _, remoteResource := range *remoteResources {
		if remoteResource.ResourceType() != aws.AwsNetworkACLResourceType &&
			remoteResource.ResourceType() != aws.AwsDefaultNetworkACLResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		remoteResource.Attrs.SafeDelete([]string{"ingress"})
		remoteResource.Attrs.SafeDelete([]string{"egress"})

		newRemoteResources = append(newRemoteResources, remoteResource)
	}

	*resourcesFromState = newResourcesFromState
	*remoteResources = newRemoteResources

	return nil
}

func (m *AwsNetworkACLExpander) expandBlock(resourcesFromState *[]*resource.Resource, networkACLID string, egress bool, ruleBlock []interface{}) []*resource.Resource {
	results := make([]*resource.Resource, 0, len(ruleBlock))

	for _, rule := range ruleBlock {
		attrs := rule.(map[string]interface{})

		attrs["rule_number"] = int64(attrs["rule_no"].(float64))
		delete(attrs, "rule_no")

		attrs["egress"] = egress

		attrs["network_acl_id"] = networkACLID

		attrs["rule_action"] = attrs["action"]
		delete(attrs, "action")

		res := m.resourceFactory.CreateAbstractResource(
			aws.AwsNetworkACLRuleResourceType,
			aws.CreateNetworkACLRuleID(
				networkACLID,
				attrs["rule_number"].(int64),
				egress,
				attrs["protocol"].(string),
			),
			attrs,
		)

		existInState := false
		for _, stateRes := range *resourcesFromState {
			if stateRes.Equal(res) {
				existInState = true
				break
			}
		}

		if !existInState {
			results = append(results, res)
		}
	}

	return results
}
