package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/helpers"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsIamPolicyResourceType is the Terraform resource type for aws_iam_policy.
const AwsIamPolicyResourceType = "aws_iam_policy"

func initAwsIAMPolicyMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsIamPolicyResourceType, func(res *resource.Resource) {
		val := res.Attrs
		jsonString, err := helpers.NormalizeJSONString((*val)["policy"])
		if err == nil {
			_ = val.SafeSet([]string{"policy"}, jsonString)
		}

		val.SafeDelete([]string{"name_prefix"})
	})
	resourceSchemaRepository.UpdateSchema(AwsIamPolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JSONString = true
		},
	})
}
