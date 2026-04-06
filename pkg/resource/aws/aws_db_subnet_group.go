package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsDbSubnetGroupResourceType is the Terraform resource type for aws_db_subnet_group.
const AwsDbSubnetGroupResourceType = "aws_db_subnet_group"

func initAwsDbSubnetGroupMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsDbSubnetGroupResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"name_prefix"})
	})
}
