package backend

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Satisfy the unused import for s3types (used to ensure we can reference S3 types if needed)
var _ s3types.BucketLocationConstraint

type mockS3GetObjectAPI struct {
	mock.Mock
}

func (m *mockS3GetObjectAPI) GetObject(ctx context.Context, params *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func TestNewS3ReaderInvalid(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *S3Backend
		wantErr error
	}{
		{
			name: "invalid path",
			args: args{
				path: "foobar",
			},
			want:    nil,
			wantErr: fmt.Errorf("Unable to parse S3 path: foobar. Must be BUCKET_NAME/PATH/TO/OBJECT"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewS3Reader(tt.args.path)
			if err.Error() != tt.wantErr.Error() {
				t.Errorf("NewS3Reader() error = '%s', wantErr '%s'", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewS3Reader() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewS3Reader(t *testing.T) {
	assert := assert.New(t)
	reader, err := NewS3Reader("sample_bucket/path/to/state.tfstate")
	if err != nil {
		t.Error(err)
	}

	assert.Equal("path/to/state.tfstate", reader.key)
	assert.Equal("sample_bucket", reader.bucket)
}

func TestS3Backend_ReadWithError(t *testing.T) {
	assert := assert.New(t)
	fakeS3 := &mockS3GetObjectAPI{}
	fakeErr := &smithy.GenericAPIError{
		Code:    "InternalError",
		Message: "Request failed on aws side",
	}
	fakeS3.On("GetObject", mock.Anything, mock.Anything).Return(nil, fakeErr)

	reader, err := NewS3Reader("foobar/path/to/state")
	if err != nil {
		t.Error(err)
	}
	reader.S3Client = fakeS3
	var b []byte
	n, err := reader.Read(b)
	assert.Empty(n)
	assert.Equal("Error reading state 'path/to/state' from s3 bucket 'foobar': Request failed on aws side", err.Error())
}

func TestS3Backend_Read(t *testing.T) {
	assert := assert.New(t)
	fakeS3 := &mockS3GetObjectAPI{}
	fakeResponse, _ := os.Open("testdata/valid.tfstate")
	defer func() { _ = fakeResponse.Close() }()
	fakeS3.On("GetObject", mock.Anything, &s3.GetObjectInput{
		Bucket: aws.String("foobar"),
		Key:    aws.String("path/to/state"),
	}).Return(&s3.GetObjectOutput{Body: fakeResponse}, nil).Once()

	reader, err := NewS3Reader("foobar/path/to/state")
	if err != nil {
		t.Error(err)
	}
	reader.S3Client = fakeS3
	_, err = io.ReadAll(reader)
	assert.Nil(err)
}
