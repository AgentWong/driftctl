package aws

import (
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
)

// FakeElastiCache is a test interface for the AWS ElastiCache API.
type FakeElastiCache interface {
	elasticacheiface.ElastiCacheAPI
}
