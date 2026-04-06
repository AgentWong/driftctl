package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsIamUserPolicyResourceType is the Terraform resource type for aws_iam_user_policy.
const AwsIamUserPolicyResourceType = "aws_iam_user_policy"

func initAwsIAMUserPolicyMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.UpdateSchema(AwsIamUserPolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JSONString = true
		},
	})
}
