package aws

import "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

// FakeDynamoDB is a test interface for the AWS DynamoDB API.
type FakeDynamoDB interface {
	dynamodbiface.DynamoDBAPI
}
