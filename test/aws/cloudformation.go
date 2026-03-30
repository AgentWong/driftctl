package aws

import (
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

// FakeCloudformation is a test interface for the AWS CloudFormation API.
type FakeCloudformation interface {
	cloudformationiface.CloudFormationAPI
}
