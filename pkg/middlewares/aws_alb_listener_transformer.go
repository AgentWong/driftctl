// Package middlewares transforms resources between enumeration and analysis.
package middlewares

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsALBListenerTransformer is a simple middleware to turn all aws_alb_listener resources into aws_lb_listener ones
// Both types provide the same functionality, but we can't know which one was used to provision cloud resources.
// AwsALBListenerTransformer so we use aws_lb_listener as the common type.
type AwsALBListenerTransformer struct {
	resourceFactory resource.Factory
}

// NewAwsALBListenerTransformer creates a AwsALBListenerTransformer.
func NewAwsALBListenerTransformer(resourceFactory resource.Factory) AwsALBListenerTransformer {
	return AwsALBListenerTransformer{
		resourceFactory: resourceFactory,
	}
}

// Execute applies the AwsALBListenerTransformer middleware.
func (m AwsALBListenerTransformer) Execute(_, resourcesFromState *[]*resource.Resource) error {
	newStateResources := make([]*resource.Resource, 0, len(*resourcesFromState))

	for _, res := range *resourcesFromState {
		if res.ResourceType() != aws.AwsApplicationLoadBalancerListenerResourceType {
			newStateResources = append(newStateResources, res)
			continue
		}

		newStateResources = append(newStateResources, m.resourceFactory.CreateAbstractResource(
			aws.AwsLoadBalancerListenerResourceType,
			res.ResourceID(),
			*res.Attributes(),
		))
	}

	*resourcesFromState = newStateResources
	return nil
}
