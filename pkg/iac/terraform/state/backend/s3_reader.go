package backend

import (
	"context"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/envproxy"
)

const BackendKeyS3 = "s3"

// S3GetObjectAPI abstracts the S3 GetObject operation for testability.
type S3GetObjectAPI interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type S3Backend struct {
	bucket   string
	key      string
	reader   io.ReadCloser
	S3Client S3GetObjectAPI
}

func NewS3Reader(path string) (*S3Backend, error) {

	backend := S3Backend{}
	bucketPath := strings.Split(path, "/")
	if len(bucketPath) < 2 {
		return nil, errors.Errorf("Unable to parse S3 path: %s. Must be BUCKET_NAME/PATH/TO/OBJECT", path)
	}
	backend.bucket = bucketPath[0]
	backend.key = strings.Join(bucketPath[1:], "/")

	envProxy := envproxy.NewEnvProxy("DCTL_S3_", "AWS_")
	envProxy.Apply()
	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	envProxy.Restore()
	if err != nil {
		return nil, err
	}
	backend.S3Client = s3.NewFromConfig(cfg)
	return &backend, nil
}

func (s *S3Backend) Read(p []byte) (n int, err error) {
	if s.reader == nil {
		response, err := s.S3Client.GetObject(context.Background(), &s3.GetObjectInput{
			Key:    aws.String(s.key),
			Bucket: aws.String(s.bucket),
		})
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				return 0, errors.Errorf(
					"Error reading state '%s' from s3 bucket '%s': %s",
					s.key,
					s.bucket,
					apiErr.ErrorMessage(),
				)
			}
			return 0, err
		}
		s.reader = response.Body
	}
	return s.reader.Read(p)
}

func (s *S3Backend) Close() error {
	if s.reader != nil {
		return s.reader.Close()
	}
	return errors.New("Unable to close reader as nothing was opened")
}
