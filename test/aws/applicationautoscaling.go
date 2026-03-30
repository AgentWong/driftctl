package aws

import (
	"github.com/aws/aws-sdk-go/service/applicationautoscaling/applicationautoscalingiface"
)

// FakeApplicationAutoScaling embeds the Application Auto Scaling interface for mock generation.
type FakeApplicationAutoScaling interface {
	applicationautoscalingiface.ApplicationAutoScalingAPI
}
