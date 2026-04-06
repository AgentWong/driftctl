package aws_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
	"github.com/snyk/driftctl/test/acceptance/awsutils"
)

func TestAcc_Aws_SQSQueue(t *testing.T) {
	t.Skip("flake")

	var mutatedQueue string
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/aws_sqs_queue"},
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
					result.Equal(2, result.Summary().TotalManaged)
					mutatedQueue = result.Managed()[0].ResourceID()
				},
			},
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				PreExec: func() {
					client := sqs.NewFromConfig(awsutils.Config())
					attributes := map[string]string{
						"DelaySeconds": "200",
					}
					_, err := client.SetQueueAttributes(context.TODO(), &sqs.SetQueueAttributesInput{
						Attributes: attributes,
						QueueUrl:   aws.String(mutatedQueue),
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Check: func(_ *test.ScanResult, _ string, err error) {
					if err != nil {
						t.Fatal(err)
					}
				},
			},
		},
	})
}
