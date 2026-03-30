package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsDefaultSecurityGroupRule remove default security group rules of the default security group from remote resources
type AwsDefaultSecurityGroupRule struct{}

// NewAwsDefaultSecurityGroupRule creates a AwsDefaultSecurityGroupRule.
func NewAwsDefaultSecurityGroupRule() AwsDefaultSecurityGroupRule {
	return AwsDefaultSecurityGroupRule{}
}

// Execute applies the AwsDefaultSecurityGroupRule middleware.
func (m AwsDefaultSecurityGroupRule) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newRemoteResources := make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		existInState := false

		// Ignore all resources other than security group rules
		if remoteResource.ResourceType() != aws.AwsSecurityGroupRuleResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Ignore if it's not the default ingress or egress rule
		if !isDefaultIngress(remoteResource, remoteResources) && !isDefaultEgress(remoteResource, remoteResources) {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		for _, stateResource := range *resourcesFromState {
			if remoteResource.Equal(stateResource) {
				existInState = true
				break
			}
		}

		if existInState {
			newRemoteResources = append(newRemoteResources, remoteResource)
		}

		if !existInState {
			logrus.WithFields(logrus.Fields{
				"id":   remoteResource.ResourceId(),
				"type": remoteResource.ResourceType(),
			}).Debug("Ignoring default unmanaged security group rule")
		}
	}

	*remoteResources = newRemoteResources

	return nil
}

func isDefaultIngress(rule *resource.Resource, remoteResources *[]*resource.Resource) bool {
	if ty := rule.Attrs.GetString("type"); ty == nil || *ty != "ingress" {
		return false
	}
	if from := rule.Attrs.GetInt("from_port"); from == nil || *from != 0 {
		return false
	}
	if to := rule.Attrs.GetInt("to_port"); to == nil || *to != 0 {
		return false
	}
	if protocol := rule.Attrs.GetString("protocol"); protocol == nil || *protocol != "-1" {
		return false
	}
	if _, exist := rule.Attrs.Get("cidr_blocks"); exist {
		return false
	}
	if _, exist := rule.Attrs.Get("ipv6_cidr_blocks"); exist {
		return false
	}
	if _, exist := rule.Attrs.Get("prefix_list_ids"); exist {
		return false
	}
	if self := rule.Attrs.GetBool("self"); self == nil || !*self {
		return false
	}
	sgID := rule.Attrs.GetString("security_group_id")
	if sgID == nil {
		return false
	}
	return isFromDefaultSecurityGroup(sgID, remoteResources)
}

func isDefaultEgress(rule *resource.Resource, remoteResources *[]*resource.Resource) bool {
	if ty := rule.Attrs.GetString("type"); ty == nil || *ty != "egress" {
		return false
	}
	if from := rule.Attrs.GetInt("from_port"); from == nil || *from != 0 {
		return false
	}
	if to := rule.Attrs.GetInt("to_port"); to == nil || *to != 0 {
		return false
	}
	if protocol := rule.Attrs.GetString("protocol"); protocol == nil || *protocol != "-1" {
		return false
	}
	if ipv4 := rule.Attrs.GetSlice("cidr_blocks"); ipv4 == nil || len(ipv4) != 1 || ipv4[0] != "0.0.0.0/0" {
		return false
	}
	if _, exist := rule.Attrs.Get("ipv6_cidr_blocks"); exist {
		return false
	}
	if _, exist := rule.Attrs.Get("prefix_list_ids"); exist {
		return false
	}
	if self := rule.Attrs.GetBool("self"); self == nil || *self {
		return false
	}
	sgID := rule.Attrs.GetString("security_group_id")
	if sgID == nil {
		return false
	}
	return isFromDefaultSecurityGroup(sgID, remoteResources)
}

func isFromDefaultSecurityGroup(sgID *string, remoteResources *[]*resource.Resource) bool {
	for _, remoteResource := range *remoteResources {
		if remoteResource.ResourceType() != aws.AwsDefaultSecurityGroupResourceType {
			continue
		}
		if *sgID == remoteResource.ResourceId() {
			return true
		}
	}
	return false
}
