// Package awsutils provides shared AWS config helpers for acceptance tests.
package awsutils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

// Config loads the default AWS config using shared credentials and config files.
func Config() aws.Config {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		panic("failed to load AWS config: " + err.Error())
	}
	return cfg
}
