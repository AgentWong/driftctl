package aws

import "github.com/aws/aws-sdk-go/service/route53/route53iface"

// FakeRoute53 is a test interface for the AWS Route 53 API.
type FakeRoute53 interface {
	route53iface.Route53API
}
