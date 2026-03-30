package middlewares

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsSNSTopicPolicyExpander explodes policy found in aws_sns_topic from state resources to aws_sns_topic_policy resources
type AwsSNSTopicPolicyExpander struct {
	resourceFactory          resource.Factory
	resourceSchemaRepository dctlresource.SchemaRepositoryInterface
}

// NewAwsSNSTopicPolicyExpander creates a AwsSNSTopicPolicyExpander.
func NewAwsSNSTopicPolicyExpander(resourceFactory resource.Factory, resourceSchemaRepository dctlresource.SchemaRepositoryInterface) AwsSNSTopicPolicyExpander {
	return AwsSNSTopicPolicyExpander{
		resourceFactory,
		resourceSchemaRepository,
	}
}

// Execute applies the AwsSNSTopicPolicyExpander middleware.
func (m AwsSNSTopicPolicyExpander) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	for _, res := range *remoteResources {
		if res.ResourceType() != aws.AwsSnsTopicResourceType {
			continue
		}
		res.Attrs.SafeDelete([]string{"policy"})
	}

	newList := make([]*resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than sns_topic
		if res.ResourceType() != aws.AwsSnsTopicResourceType {
			newList = append(newList, res)
			continue
		}

		newList = append(newList, res)

		if m.hasPolicyAttached(res, resourcesFromState) {
			res.Attrs.SafeDelete([]string{"policy"})
			continue
		}

		err := m.splitPolicy(res, &newList)
		if err != nil {
			return err
		}
	}
	*resourcesFromState = newList
	return nil
}

func (m *AwsSNSTopicPolicyExpander) splitPolicy(topic *resource.Resource, results *[]*resource.Resource) error {
	policy, exist := topic.Attrs.Get("policy")
	if !exist || policy == "" {
		return nil
	}

	arn, exist := topic.Attrs.Get("arn")
	if !exist || arn == "" {
		return errors.Errorf("No arn found for resource %s (%s)", topic.ID, topic.Type)
	}

	data := map[string]interface{}{
		"arn":    arn,
		"id":     topic.ID,
		"policy": policy,
	}

	newPolicy := m.resourceFactory.CreateAbstractResource("aws_sns_topic_policy", topic.ID, data)

	*results = append(*results, newPolicy)
	logrus.WithFields(logrus.Fields{
		"id": newPolicy.ResourceID(),
	}).Debug("Created new policy from sns_topic")

	topic.Attrs.SafeDelete([]string{"policy"})
	return nil
}

func (m *AwsSNSTopicPolicyExpander) hasPolicyAttached(topic *resource.Resource, resourcesFromState *[]*resource.Resource) bool {
	for _, res := range *resourcesFromState {
		if res.ResourceType() == aws.AwsSnsTopicPolicyResourceType &&
			res.ResourceID() == topic.ID {
			return true
		}
	}
	return false
}
