package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsKeyPairResourceType is the Terraform resource type for aws_key_pair.
const AwsKeyPairResourceType = "aws_key_pair"

func initAwsKeyPairMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsKeyPairResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"key_name_prefix"})
		val.SafeDelete([]string{"public_key"})
	})
}
