package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsIamRoleResourceType is the Terraform resource type for aws_iam_role.
const AwsIamRoleResourceType = "aws_iam_role"

func initAwsIAMRoleMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsIamRoleResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"force_detach_policies"})
	})
	resourceSchemaRepository.UpdateSchema(AwsIamRoleResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"assume_role_policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JSONString = true
		},
	})
}
