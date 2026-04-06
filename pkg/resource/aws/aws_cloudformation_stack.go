package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsCloudformationStackResourceType is the Terraform resource type for aws_cloudformation_stack.
const AwsCloudformationStackResourceType = "aws_cloudformation_stack"

func initAwsCloudformationStackMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsCloudformationStackResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
}
