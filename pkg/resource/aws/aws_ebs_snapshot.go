package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsEbsSnapshotResourceType is the Terraform resource type for aws_ebs_snapshot.
const AwsEbsSnapshotResourceType = "aws_ebs_snapshot"

func initAwsEbsSnapshotMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsEbsSnapshotResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
}
