package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsLaunchTemplateResourceType is the Terraform resource type for AWS EC2 launch templates.
const AwsLaunchTemplateResourceType = "aws_launch_template"

func initAwsLaunchTemplateMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsLaunchTemplateResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"arn"})
	})
}
