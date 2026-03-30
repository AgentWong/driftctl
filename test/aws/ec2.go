package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// FakeEC2 is a test interface for the AWS EC2 API.
type FakeEC2 interface {
	ec2iface.EC2API
}
