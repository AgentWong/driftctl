package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsIamUserResourceType is the Terraform resource type for IAM users.
const AwsIamUserResourceType = "aws_iam_user"

func initAwsIAMUserMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsIamUserResourceType, func(res *resource.Resource) {
		val := res.Attrs
		permissionsBoundary, exist := val.Get("permissions_boundary")
		if exist && permissionsBoundary == "" {
			val.SafeDelete([]string{"permissions_boundary"})
		}
		val.SafeDelete([]string{"force_destroy"})
	})
}
