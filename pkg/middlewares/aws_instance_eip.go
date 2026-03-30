package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsInstanceEIP is a middleware.
type AwsInstanceEIP struct{}

// Execute applies the AwsInstanceEIP middleware.
func (a AwsInstanceEIP) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than aws_instance
		if remoteResource.ResourceType() != aws.AwsInstanceResourceType {
			continue
		}

		if a.hasEIP(remoteResource, resourcesFromState) {
			logrus.WithFields(logrus.Fields{
				"instance": remoteResource.ResourceID(),
			}).Debug("Ignore instance public ip and dns as it has an eip attached")
			a.ignorePublicIPAndDNS(remoteResource, remoteResources, resourcesFromState)
		}
	}

	return nil
}

func (a AwsInstanceEIP) hasEIP(instance *resource.Resource, resources *[]*resource.Resource) bool {
	for _, res := range *resources {
		if res.ResourceType() == aws.AwsEipResourceType {
			if (*res.Attrs)["instance"] == instance.ResourceID() {
				return true
			}
		}
		if res.ResourceType() == aws.AwsEipAssociationResourceType {
			if (*res.Attrs)["instance_id"] == instance.ResourceID() {
				return true
			}
		}
	}

	return false
}

func (a AwsInstanceEIP) ignorePublicIPAndDNS(instance *resource.Resource, resourcesSet ...*[]*resource.Resource) {
	for _, resources := range resourcesSet {
		for _, res := range *resources {
			if res.ResourceType() == instance.ResourceType() &&
				res.ResourceID() == instance.ResourceID() {
				res.Attrs.SafeDelete([]string{"public_dns"})
				res.Attrs.SafeDelete([]string{"public_ip"})
			}
		}
	}
}
