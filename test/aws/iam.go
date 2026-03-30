package aws

import (
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

// FakeIAM is a test interface for the AWS IAM API.
type FakeIAM interface {
	iamiface.IAMAPI
}
