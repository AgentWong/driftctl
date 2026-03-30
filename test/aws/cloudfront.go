package aws

import "github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"

// FakeCloudFront is a test interface for the AWS CloudFront API.
type FakeCloudFront interface {
	cloudfrontiface.CloudFrontAPI
}
