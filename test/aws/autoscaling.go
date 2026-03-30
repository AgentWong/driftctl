package aws

import (
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
)

// FakeAutoscaling is a test interface for the AWS Auto Scaling API.
type FakeAutoscaling interface {
	autoscalingiface.AutoScalingAPI
}
