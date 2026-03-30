package aws

import "github.com/aws/aws-sdk-go/service/lambda/lambdaiface"

// FakeLambda is a test interface for the AWS Lambda API.
type FakeLambda interface {
	lambdaiface.LambdaAPI
}
