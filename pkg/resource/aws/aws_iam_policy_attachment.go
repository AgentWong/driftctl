package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsIamPolicyAttachmentResourceType is the Terraform resource type for aws_iam_policy_attachment.
const AwsIamPolicyAttachmentResourceType = "aws_iam_policy_attachment"

func initAwsIAMPolicyAttachmentMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsIamPolicyAttachmentResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"name"})
	})
}
