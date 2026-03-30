package aws

import (
	"github.com/aws/aws-sdk-go/service/apigatewayv2/apigatewayv2iface"
)

// FakeAPIGatewayV2 embeds the API Gateway V2 interface for mock generation.
type FakeAPIGatewayV2 interface {
	apigatewayv2iface.ApiGatewayV2API
}
