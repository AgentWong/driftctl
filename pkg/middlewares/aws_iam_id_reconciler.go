package middlewares

import (
	"regexp"

	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// awsIamUniqueIDPattern matches AWS-internal IAM unique IDs that Config
// returns instead of the friendly name Terraform uses.
// Roles start with AROA, users with AIDA, policies with ANPA/ANVA.
var awsIamUniqueIDPattern = regexp.MustCompile(`^(AROA|AIDA|ANPA|ANVA)[A-Z0-9]{17,}$`)

// AwsIamIDReconciler rewrites remote IAM resource IDs so they match
// the Terraform convention. AWS Config returns internal unique IDs
// (AROA… for roles, AIDA… for users) but Terraform uses the friendly
// name. For IAM policies Terraform uses the ARN.
type AwsIamIDReconciler struct{}

func NewAwsIamIDReconciler() *AwsIamIDReconciler {
	return &AwsIamIDReconciler{}
}

func (m AwsIamIDReconciler) Execute(remoteResources, _ *[]*resource.Resource) error {
	// types where Terraform uses the name as the ID
	nameTypes := map[string]bool{
		aws.AwsIamRoleResourceType: true,
		aws.AwsIamUserResourceType: true,
	}
	// types where Terraform uses the ARN as the ID
	arnTypes := map[string]bool{
		aws.AwsIamPolicyResourceType: true,
	}

	for _, res := range *remoteResources {
		// Only rewrite IDs that look like AWS internal unique identifiers
		if !awsIamUniqueIDPattern.MatchString(res.ResourceId()) {
			continue
		}

		if nameTypes[res.ResourceType()] {
			name := res.DisplayName()
			if name != "" {
				logrus.WithFields(logrus.Fields{
					"type":   res.ResourceType(),
					"old_id": res.ResourceId(),
					"new_id": name,
				}).Debug("Reconciling IAM resource ID from Config unique ID to name")
				res.Id = name
			}
			continue
		}
		if arnTypes[res.ResourceType()] {
			if res.Attrs == nil {
				continue
			}
			arn, ok := (*res.Attrs)["arn"].(string)
			if ok && arn != "" {
				logrus.WithFields(logrus.Fields{
					"type":   res.ResourceType(),
					"old_id": res.ResourceId(),
					"new_id": arn,
				}).Debug("Reconciling IAM policy ID from Config unique ID to ARN")
				res.Id = arn
			}
		}
	}
	return nil
}
