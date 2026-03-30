package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsBucketPolicyExpander explodes policy found in aws_s3_bucket.policy from state resources to dedicated resources
type AwsBucketPolicyExpander struct {
	resourceFactory resource.Factory
}

// NewAwsBucketPolicyExpander creates a AwsBucketPolicyExpander.
func NewAwsBucketPolicyExpander(resourceFactory resource.Factory) AwsBucketPolicyExpander {
	return AwsBucketPolicyExpander{
		resourceFactory: resourceFactory,
	}
}

// Execute applies the AwsBucketPolicyExpander middleware.
func (m AwsBucketPolicyExpander) Execute(_, resourcesFromState *[]*resource.Resource) error {
	newList := make([]*resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than s3_bucket
		if res.ResourceType() != aws.AwsS3BucketResourceType {
			newList = append(newList, res)
			continue
		}

		newList = append(newList, res)

		if hasPolicyAttached(res.ResourceID(), resourcesFromState) {
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

func (m *AwsBucketPolicyExpander) handlePolicy(bucket *resource.Resource, results *[]*resource.Resource) error {
	policyAttr, exist := bucket.Attrs.Get("policy")
	if !exist || policyAttr == nil || policyAttr == "" {
		return nil
	}

	data := map[string]interface{}{
		"id":     bucket.ResourceID(),
		"bucket": (*bucket.Attrs)["bucket"],
		"policy": (*bucket.Attrs)["policy"],
	}

	newPolicy := m.resourceFactory.CreateAbstractResource(aws.AwsS3BucketPolicyResourceType, bucket.ResourceID(), data)

	*results = append(*results, newPolicy)
	logrus.WithFields(logrus.Fields{
		"id": newPolicy.ResourceID(),
	}).Debug("Created new policy from bucket")

	bucket.Attrs.SafeDelete([]string{"policy"})
	return nil
}

// Return true if the bucket has a aws_bucket_policy resource attached to itself.
// It is mandatory since it's possible to have a aws_bucket with an inline policy
// AND a aws_bucket_policy resource at the same time. At the end, on the AWS console,
// hasPolicyAttached the aws_bucket_policy will be used.
func hasPolicyAttached(bucket string, resourcesFromState *[]*resource.Resource) bool {
	for _, res := range *resourcesFromState {
		if res.ResourceType() == aws.AwsS3BucketPolicyResourceType &&
			res.ResourceID() == bucket {
			return true
		}
	}
	return false
}
