package aws

import (
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
)

// FakeELBV2 is a test interface for the AWS ELBv2 API.
type FakeELBV2 interface {
	elbv2iface.ELBV2API
}
