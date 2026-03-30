package aws

import (
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

// FakeSNS embeds the SNS interface for mock generation.
type FakeSNS interface {
	snsiface.SNSAPI
}
