package aws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
	"github.com/snyk/driftctl/test/acceptance/awsutils"
)

func TestAcc_Aws_CloudfrontDistribution(t *testing.T) {
	t.Skip("flake")

	var mutatedDistribution string
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion:           "0.15.5",
		Paths:                      []string{"./testdata/acc/aws_cloudfront_distribution"},
		Args:                       []string{"scan"},
		ShouldRefreshBeforeDestroy: true,
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				Check: func(result *test.ScanResult, _ string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					result.AssertManagedCount(1)
					mutatedDistribution = result.Managed()[0].ResourceID()
				},
			},
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				PreExec: func() {
					client := cloudfront.NewFromConfig(awsutils.Config())
					res, err := client.GetDistributionConfig(context.TODO(), &cloudfront.GetDistributionConfigInput{
						Id: aws.String(mutatedDistribution),
					})
					if err != nil {
						t.Fatal(err)
					}
					res.DistributionConfig.IsIPV6Enabled = aws.Bool(true)
					_, err = client.UpdateDistribution(context.TODO(), &cloudfront.UpdateDistributionInput{
						Id:                 aws.String(mutatedDistribution),
						DistributionConfig: res.DistributionConfig,
						IfMatch:            res.ETag,
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
