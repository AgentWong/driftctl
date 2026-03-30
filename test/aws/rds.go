package aws

import "github.com/aws/aws-sdk-go/service/rds/rdsiface"

// FakeRDS embeds the RDS interface for mock generation.
type FakeRDS interface {
	rdsiface.RDSAPI
}
