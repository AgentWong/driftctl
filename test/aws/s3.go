package aws

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3control/s3controliface"
)

// FakeS3 is a test interface for the AWS S3 API.
type FakeS3 interface {
	s3iface.S3API
}

// FakeS3Control is a test interface for the AWS S3 Control API.
type FakeS3Control interface {
	s3controliface.S3ControlAPI
}

// FakeRequestFailure is a test interface for AWS S3 request failures.
type FakeRequestFailure interface {
	s3.RequestFailure
}
