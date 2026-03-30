package aws

import "github.com/aws/aws-sdk-go/service/kms/kmsiface"

// FakeKMS is a test interface for the AWS KMS API.
type FakeKMS interface {
	kmsiface.KMSAPI
}
