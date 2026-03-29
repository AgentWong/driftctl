package enumerator

import (
	"context"
	"fmt"
	"path"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/envproxy"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/snyk/driftctl/pkg/iac/config"
)

type S3Enumerator struct {
	config config.SupplierConfig
	client s3.ListObjectsV2APIClient
}

func NewS3Enumerator(config config.SupplierConfig) *S3Enumerator {
	envProxy := envproxy.NewEnvProxy("DCTL_S3_", "AWS_")
	envProxy.Apply()
	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	envProxy.Restore()
	if err != nil {
		// Maintain same behavior as session.Must - panic on config error
		panic("failed to load AWS config: " + err.Error())
	}
	return &S3Enumerator{
		config,
		s3.NewFromConfig(cfg),
	}
}

func (s *S3Enumerator) Origin() string {
	return s.config.String()
}

func (s *S3Enumerator) Enumerate() ([]string, error) {
	bucketPath := strings.Split(s.config.Path, "/")
	if len(bucketPath) < 2 {
		return nil, errors.Errorf("Unable to parse S3 path: %s. Must be BUCKET_NAME/PREFIX", s.config.Path)
	}

	bucket := bucketPath[0]
	// prefix should contains everything that does not have a glob pattern
	// Pattern should be the glob matcher string
	prefix, pattern := extractPrefixAndPattern(strings.Join(bucketPath[1:], "/"))

	// We combine the prefix and pattern to match file names against.
	fullPattern := path.Join(prefix, pattern)

	files := make([]string, 0)
	input := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	}
	paginator := s3.NewListObjectsV2Paginator(s.client, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, metadata := range output.Contents {
			if metadata.Size != nil && *metadata.Size > 0 {
				key := *metadata.Key
				if match, _ := doublestar.Match(fullPattern, key); match {
					files = append(files, strings.Join([]string{bucket, key}, "/"))
				}
			}
		}
	}

	if len(files) == 0 {
		return files, fmt.Errorf("no Terraform state was found in %s, exiting", s.config.Path)
	}

	return files, nil
}
