package aws

import (
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
)

// FakeELB is a test interface for the AWS ELB API.
type FakeELB interface {
	elbiface.ELBAPI
}
