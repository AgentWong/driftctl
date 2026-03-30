package aws

import (
	"github.com/aws/aws-sdk-go/service/cloudtrail/cloudtrailiface"
)

// FakeCloudtrail is a test interface for the AWS CloudTrail API.
type FakeCloudtrail interface {
	cloudtrailiface.CloudTrailAPI
}
