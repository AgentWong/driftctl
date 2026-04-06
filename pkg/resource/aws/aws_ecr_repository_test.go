package aws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
	"github.com/snyk/driftctl/test/acceptance/awsutils"
)

func TestAcc_Aws_ECRRepository(t *testing.T) {
	var mutatedRepositoryID string
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/aws_ecr_repository"},
		Args:             []string{"scan"},
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

					mutatedRepositoryID = result.Managed()[0].ResourceID()
				},
			},
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				PreExec: func() {
					client := ecr.NewFromConfig(awsutils.Config())
					_, err := client.PutImageTagMutability(context.TODO(), &ecr.PutImageTagMutabilityInput{
						RepositoryName:     &mutatedRepositoryID,
						ImageTagMutability: ecrtypes.ImageTagMutabilityImmutable,
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
					result.AssertManagedCount(1)
					result.AssertUnmanagedCount(0)
				},
			},
		},
	})
}
