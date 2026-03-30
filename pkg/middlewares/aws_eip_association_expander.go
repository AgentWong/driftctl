package middlewares

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

/**
  Fetching eip association from remote return every association but some of them are embedded in eip.
  This middleware will check for every eip_association that here is no corresponding association_id inside eip.
*/

// EipAssociationExpander is a middleware.
type EipAssociationExpander struct {
	resourceFactory resource.Factory
}

// NewEipAssociationExpander creates a EipAssociationExpander.
func NewEipAssociationExpander(resourceFactory resource.Factory) EipAssociationExpander {
	return EipAssociationExpander{resourceFactory}
}

// Execute applies the EipAssociationExpander middleware.
func (m EipAssociationExpander) Execute(_, resourcesFromState *[]*resource.Resource) error {
	var newResources []*resource.Resource
	for _, res := range *resourcesFromState {
		newResources = append(newResources, res)

		if res.ResourceType() != aws.AwsEipResourceType {
			continue
		}
		if m.haveMatchingEipAssociation(res, resourcesFromState) {
			continue
		}
		// This EIP have no association, check if we need to create one
		assocID := res.Attributes().GetString("association_id")
		if assocID == nil || *assocID == "" {
			continue
		}

		attributes := *res.Attributes()
		newAssoc := m.resourceFactory.CreateAbstractResource(
			aws.AwsEipAssociationResourceType,
			*assocID,
			map[string]interface{}{
				"allocation_id":        res.ResourceId(),
				"id":                   *assocID,
				"instance_id":          attributes["instance"],
				"network_interface_id": attributes["network_interface"],
				"private_ip_address":   attributes["private_ip"],
				"public_ip":            attributes["public_ip"],
			},
		)

		newResources = append(newResources, newAssoc)
	}
	*resourcesFromState = newResources

	return nil
}

func (m EipAssociationExpander) haveMatchingEipAssociation(cur *resource.Resource, stateRes *[]*resource.Resource) bool {
	for _, res := range *stateRes {
		if res.ResourceType() != aws.AwsEipAssociationResourceType {
			continue
		}
		assocID := cur.Attributes().GetString("association_id")
		if assocID != nil && res.ResourceId() == *assocID {
			return true
		}
	}
	return false
}
