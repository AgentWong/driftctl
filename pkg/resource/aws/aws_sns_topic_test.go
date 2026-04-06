package aws_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
	"github.com/snyk/driftctl/test/acceptance/awsutils"
)

func TestAcc_Aws_SNSTopic(t *testing.T) {
	var mutatedTopicArn string
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/aws_sns_topic"},
		Args:             []string{"scan"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				ShouldRetry: acceptance.LinearBackoff(10 * time.Minute),
				Check: func(result *test.ScanResult, _ string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					result.AssertManagedCount(3)

					for _, resource := range result.Managed() {
						if strings.Contains(resource.ResourceID(), "user-updates-topic3") {
							mutatedTopicArn = resource.ResourceID()
						}
					}
				},
			},
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				PreExec: func() {
					client := sns.NewFromConfig(awsutils.Config())
					_, err := client.SetTopicAttributes(context.TODO(), &sns.SetTopicAttributesInput{
						AttributeName:  aws.String("DisplayName"),
						AttributeValue: aws.String("CHANGED"),
						TopicArn:       &mutatedTopicArn,
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Check: func(result *test.ScanResult, _ string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertDeletedCount(0)
					result.AssertManagedCount(3)
				},
			},
		},
	})
}
