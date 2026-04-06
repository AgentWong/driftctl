package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/helpers"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsSnsTopicPolicyResourceType is the Terraform resource type for aws_sns_topic_policy.
const AwsSnsTopicPolicyResourceType = "aws_sns_topic_policy"

func initSnsTopicPolicyMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsSnsTopicPolicyResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"owner"})
		jsonString, err := helpers.NormalizeJSONString((*val)["policy"])
		if err != nil {
			return
		}
		_ = val.SafeSet([]string{"policy"}, jsonString)
	})
	resourceSchemaRepository.UpdateSchema(AwsSnsTopicPolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JSONString = true
		},
	})
}
