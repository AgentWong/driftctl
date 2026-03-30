package aws

import "github.com/aws/aws-sdk-go/service/ecr/ecriface"

// FakeECR is a test interface for the AWS ECR API.
type FakeECR interface {
	ecriface.ECRAPI
}
