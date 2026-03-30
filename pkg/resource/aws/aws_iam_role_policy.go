package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsIamRolePolicyResourceType is the Terraform resource type for aws_iam_role_policy.
const AwsIamRolePolicyResourceType = "aws_iam_role_policy"

func initAwsIAMRolePolicyMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.UpdateSchema(AwsIamRolePolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JSONString = true
		},
	})
}
