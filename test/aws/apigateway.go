// Package aws provides test fakes for AWS SDK service interfaces.
package aws

import (
	"github.com/aws/aws-sdk-go/service/apigateway/apigatewayiface"
)

// FakeAPIGateway embeds the API Gateway interface for mock generation.
type FakeAPIGateway interface {
	apigatewayiface.APIGatewayAPI
}
