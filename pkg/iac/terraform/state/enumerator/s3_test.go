package enumerator

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/snyk/driftctl/pkg/iac/config"
)

// mockListObjectsV2Client implements s3.ListObjectsV2APIClient for testing.
// It returns pages sequentially, using NextContinuationToken to chain them.
type mockListObjectsV2Client struct {
	pages []s3.ListObjectsV2Output
	err   error
	calls int
}

func (m *mockListObjectsV2Client) ListObjectsV2(_ context.Context, _ *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.calls >= len(m.pages) {
		return &s3.ListObjectsV2Output{}, nil
	}
	page := m.pages[m.calls]
	m.calls++
	return &page, nil
}

func TestS3Enumerator_NewS3Enumerator(t *testing.T) {
	tests := []struct {
		name   string
		config config.SupplierConfig
		setEnv map[string]string
	}{
		{
			name: "test with no proxy env var",
			config: config.SupplierConfig{
				Key:     "tfstate",
				Backend: "s3",
				Path:    "terraform.tfstate",
			},
			setEnv: map[string]string{
				"AWS_DEFAULT_REGION": "us-east-1",
			},
		},
		{
			name: "test with proxy env var",
			config: config.SupplierConfig{
				Key:     "tfstate",
				Backend: "s3",
				Path:    "terraform.tfstate",
			},
			setEnv: map[string]string{
				"AWS_DEFAULT_REGION":     "us-east-1",
				"DCTL_S3_DEFAULT_REGION": "eu-west-3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.setEnv {
				os.Setenv(key, value)
			}
			// Verify NewS3Enumerator doesn't panic with valid config
			enumerator := NewS3Enumerator(tt.config)
			if enumerator == nil {
				t.Error("NewS3Enumerator() returned nil")
			}
		})
	}
}

// s3Object is a helper to construct s3types.Object for tests.
func s3Object(key string, size int64) s3types.Object {
	return s3types.Object{Key: aws.String(key), Size: aws.Int64(size)}
}

func TestS3Enumerator_Enumerate(t *testing.T) {
	tests := []struct {
		name   string
		config config.SupplierConfig
		client *mockListObjectsV2Client
		want   []string
		err    string
	}{
		{
			name: "no test results are returned",
			config: config.SupplierConfig{
				Path: "bucket-name/a/nested/prefix",
			},
			client: &mockListObjectsV2Client{
				pages: []s3.ListObjectsV2Output{
					{
						IsTruncated:           aws.Bool(true),
						NextContinuationToken: aws.String("token1"),
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/state1", 5),
							s3Object("a/nested/prefix/state2", 2),
							s3Object("a/nested/prefix/state3", 1),
						},
					},
					{
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/state4", 5),
							s3Object("a/nested/prefix/folder1/state5", 5),
							s3Object("a/nested/prefix/folder2/subfolder1/state6", 5),
						},
					},
				},
			},
			want: []string{},
			err:  "no Terraform state was found in bucket-name/a/nested/prefix, exiting",
		},
		{
			name: "one test result is returned",
			config: config.SupplierConfig{
				Path: "bucket-name/a/nested/prefix/state2",
			},
			client: &mockListObjectsV2Client{
				pages: []s3.ListObjectsV2Output{
					{
						IsTruncated:           aws.Bool(true),
						NextContinuationToken: aws.String("token1"),
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/state1", 5),
							s3Object("a/nested/prefix/state2", 2),
							s3Object("a/nested/prefix/state3", 1),
						},
					},
					{
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/state4", 5),
							s3Object("a/nested/prefix/folder1/state5", 5),
							s3Object("a/nested/prefix/folder2/subfolder1/state6", 5),
						},
					},
				},
			},
			want: []string{"bucket-name/a/nested/prefix/state2"},
		},
		{
			name: "test results with simple doublestar glob",
			config: config.SupplierConfig{
				Path: "bucket-name/**/*.tfstate",
			},
			client: &mockListObjectsV2Client{
				pages: []s3.ListObjectsV2Output{
					{
						IsTruncated:           aws.Bool(true),
						NextContinuationToken: aws.String("token1"),
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/1/state1.tfstate", 5),
							s3Object("a/nested/folder1/2/state2.tfstate", 5),
							s3Object("a/nested/prefix/state3.tfstate", 5),
						},
					},
					{
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/4/4/state4.tfstate", 5),
							s3Object("a/nested/state5.tfstate", 5),
							s3Object("a/nested/prefix/state6.tfstate.backup", 5),
						},
					},
				},
			},
			want: []string{
				"bucket-name/a/nested/prefix/1/state1.tfstate",
				"bucket-name/a/nested/folder1/2/state2.tfstate",
				"bucket-name/a/nested/prefix/state3.tfstate",
				"bucket-name/a/nested/prefix/4/4/state4.tfstate",
				"bucket-name/a/nested/state5.tfstate",
			},
			err: "",
		},
		{
			name: "test results with glob and prefix after glob",
			config: config.SupplierConfig{
				Path: "bucket-name/a/**/b/*.tfstate",
			},
			client: &mockListObjectsV2Client{
				pages: []s3.ListObjectsV2Output{
					{
						Contents: []s3types.Object{
							s3Object("a/prefix/b/state1.tfstate", 5),
							s3Object("a/b/state2.tfstate", 5),
							s3Object("a/prefix/state3.tfstate", 5),
							s3Object("a/prefix/state4.tfstate.backup", 5),
						},
					},
				},
			},
			want: []string{
				"bucket-name/a/prefix/b/state1.tfstate",
				"bucket-name/a/b/state2.tfstate",
			},
			err: "",
		},
		{
			name: "test results with glob",
			config: config.SupplierConfig{
				Path: "bucket-name/a/nested/prefix/**/*.tfstate",
			},
			client: &mockListObjectsV2Client{
				pages: []s3.ListObjectsV2Output{
					{
						IsTruncated:           aws.Bool(true),
						NextContinuationToken: aws.String("token1"),
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/1/state1.tfstate", 5),
							s3Object("a/nested/folder1/2/state2.tfstate", 5),
							s3Object("a/nested/prefix/state3.tfstate", 5),
						},
					},
					{
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/4/4/state4.tfstate", 5),
							s3Object("a/nested/state5.state", 5),
							s3Object("a/nested/prefix/state6.tfstate.backup", 5),
						},
					},
				},
			},
			want: []string{
				"bucket-name/a/nested/prefix/1/state1.tfstate",
				"bucket-name/a/nested/prefix/state3.tfstate",
				"bucket-name/a/nested/prefix/4/4/state4.tfstate",
			},
			err: "",
		},
		{
			name: "test results with simple glob",
			config: config.SupplierConfig{
				Path: "bucket-name/a/nested/prefix/*.tfstate",
			},
			client: &mockListObjectsV2Client{
				pages: []s3.ListObjectsV2Output{
					{
						IsTruncated:           aws.Bool(true),
						NextContinuationToken: aws.String("token1"),
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/1/state1.tfstate", 5),
							s3Object("a/nested/prefix/2/state2.tfstate", 5),
							s3Object("a/nested/prefix/state3.tfstate", 5),
						},
					},
					{
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/4/4/state4.tfstate", 5),
							s3Object("a/nested/prefix/state5.state", 5),
							s3Object("a/nested/prefix/state6.tfstate.backup", 5),
						},
					},
				},
			},
			want: []string{"bucket-name/a/nested/prefix/state3.tfstate"},
			err:  "",
		},
		{
			name: "test when invalid config used",
			config: config.SupplierConfig{
				Path: "bucket-name",
			},
			client: &mockListObjectsV2Client{err: errors.New("error when listing")},
			want:   nil,
			err:    "Unable to parse S3 path: bucket-name. Must be BUCKET_NAME/PREFIX",
		},
		{
			name:   "test when empty config used",
			config: config.SupplierConfig{},
			client: &mockListObjectsV2Client{err: errors.New("error when listing")},
			want:   nil,
			err:    "Unable to parse S3 path: . Must be BUCKET_NAME/PREFIX",
		},
		{
			name: "test enumeration return error",
			config: config.SupplierConfig{
				Path: "bucket-name/a/nested/prefix",
			},
			client: &mockListObjectsV2Client{err: errors.New("error when listing")},
			want:   nil,
			err:    "error when listing",
		},
		{
			name: "test no state found with simple path",
			config: config.SupplierConfig{
				Path: "bucket-name/a/nested/prefix",
			},
			client: &mockListObjectsV2Client{
				pages: []s3.ListObjectsV2Output{
					{
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/1/state1.tfstate", 5),
						},
					},
				},
			},
			want: []string{},
			err:  "no Terraform state was found in bucket-name/a/nested/prefix, exiting",
		},
		{
			name: "test no state found with simple glob path",
			config: config.SupplierConfig{
				Path: "bucket-name/a/nested/prefix/*",
			},
			client: &mockListObjectsV2Client{
				pages: []s3.ListObjectsV2Output{
					{
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/1/state1.tfstate", 5),
						},
					},
				},
			},
			want: []string{},
			err:  "no Terraform state was found in bucket-name/a/nested/prefix/*, exiting",
		},
		{
			name: "test no state found with double star glob path",
			config: config.SupplierConfig{
				Path: "bucket-name/a/nested/prefix/**/*.tfstate",
			},
			client: &mockListObjectsV2Client{
				pages: []s3.ListObjectsV2Output{
					{
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/1/dummy.json", 5),
						},
					},
				},
			},
			want: []string{},
			err:  "no Terraform state was found in bucket-name/a/nested/prefix/**/*.tfstate, exiting",
		},
		{
			name: "test folder terraform.tfstate is not recognized as a file",
			config: config.SupplierConfig{
				Path: "bucket-name/a/nested/**/*.tfstate",
			},
			client: &mockListObjectsV2Client{
				pages: []s3.ListObjectsV2Output{
					{
						Contents: []s3types.Object{
							s3Object("a/nested/prefix/terraform.tfstate/terraform.tfstate", 5),
						},
					},
				},
			},
			want: []string{"bucket-name/a/nested/prefix/terraform.tfstate/terraform.tfstate"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &S3Enumerator{
				config: tt.config,
				client: tt.client,
			}
			got, err := s.Enumerate()
			if err != nil && err.Error() != tt.err {
				t.Fatalf("Expected error '%s', got '%s'", tt.err, err.Error())
				return
			}
			if tt.err != "" && err == nil {
				t.Fatalf("Expected error '%s' but got nil", tt.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Enumerate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
