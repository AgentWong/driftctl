package client

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
)

type AwsClientFactoryInterface interface {
	GetS3Client(optFns ...func(*s3.Options)) *s3.Client
	GetS3ControlClient(optFns ...func(*s3control.Options)) *s3control.Client
}

type AwsClientFactory struct {
	config aws.Config
}

func NewAWSClientFactory(config aws.Config) *AwsClientFactory {
	return &AwsClientFactory{config}
}

func (s AwsClientFactory) GetS3Client(optFns ...func(*s3.Options)) *s3.Client {
	return s3.NewFromConfig(s.config, optFns...)
}

func (s AwsClientFactory) GetS3ControlClient(optFns ...func(*s3control.Options)) *s3control.Client {
	return s3control.NewFromConfig(s.config, optFns...)
}
