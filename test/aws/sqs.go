package aws

import (
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

// FakeSQS is a test interface for the AWS SQS API.
type FakeSQS interface {
	sqsiface.SQSAPI
}
