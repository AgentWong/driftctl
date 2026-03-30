package backend

import (
	"testing"

	"github.com/snyk/driftctl/test/mocks"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestTFCloudBackend_Read(t *testing.T) {
	type args struct {
		workspaceID string
		options     *Options
	}
	tests := []struct {
		name     string
		args     args
		wantErr  error
		expected string
		mock     func(*mocks.Workspaces, *mocks.StateVersions)
	}{
		{
			name: "Should fetch URL with auth header",
			args: args{
				workspaceID: "ws-ABCDEFG12345678",
				options: &Options{
					TFCloudToken:    "TOKEN",
					TFCloudEndpoint: "https://app.terraform.io/api/v2",
				},
			},
			wantErr:  nil,
			expected: "{}",
			mock: func(_ *mocks.Workspaces, StateVersions *mocks.StateVersions) {
				retDownloadURL := "https://archivist.terraform.io/v1/object/test"
				StateVersions.On("Current", mock.Anything, "ws-ABCDEFG12345678").Return(&tfe.StateVersion{DownloadURL: retDownloadURL}, nil)
				StateVersions.On("Download", mock.Anything, retDownloadURL).Return([]byte(`{}`), nil)
			},
		},
		{
			name: "Should resolve path and return state",
			args: args{
				workspaceID: "some-org/some-workspace",
				options: &Options{
					TFCloudToken:    "TOKEN",
					TFCloudEndpoint: "https://app.terraform.io/api/v2",
				},
			},
			wantErr:  nil,
			expected: "{}",
			mock: func(Workspaces *mocks.Workspaces, StateVersions *mocks.StateVersions) {
				Workspaces.On("Read", mock.Anything, "some-org", "some-workspace").Return(&tfe.Workspace{ID: "ws-ABCDEFG12345678"}, nil)
				retDownloadURL := "https://archivist.terraform.io/v1/object/test"
				StateVersions.On("Current", mock.Anything, "ws-ABCDEFG12345678").Return(&tfe.StateVersion{DownloadURL: retDownloadURL}, nil)
				StateVersions.On("Download", mock.Anything, retDownloadURL).Return([]byte(`{}`), nil)
			},
		},
		{
			name: "Should fail with wrong workspaceId",
			args: args{
				workspaceID: "ws-ABCDEFG12345678",
				options: &Options{
					TFCloudToken:    "TOKEN",
					TFCloudEndpoint: "https://app.terraform.io/api/v2",
				},
			},
			mock: func(_ *mocks.Workspaces, StateVersions *mocks.StateVersions) {
				retDownloadURL := "https://archivist.terraform.io/v1/object/test"
				StateVersions.On("Current", mock.Anything, "ws-ABCDEFG12345678").Return(&tfe.StateVersion{DownloadURL: retDownloadURL}, errors.New("resource not found"))
			},
			wantErr: errors.New("unable to read current state version: resource not found"),
		},
		{
			name: "Should fail with download error",
			args: args{
				workspaceID: "ws-ABCDEFG12345678",
				options: &Options{
					TFCloudToken:    "TOKEN",
					TFCloudEndpoint: "https://app.terraform.io/api/v2",
				},
			},
			mock: func(_ *mocks.Workspaces, StateVersions *mocks.StateVersions) {
				retDownloadURL := "https://archivist.terraform.io/v1/object/test"
				StateVersions.On("Current", mock.Anything, "ws-ABCDEFG12345678").Return(&tfe.StateVersion{DownloadURL: retDownloadURL}, nil)
				StateVersions.On("Download", mock.Anything, retDownloadURL).Return([]byte(`{}`), errors.New("connection terminated"))
			},
			wantErr: errors.New("unable to download current state content: connection terminated"),
		},
		{
			name: "Should fail with bad authentication token - workspace id",
			args: args{
				workspaceID: "ws-ABCDEFG12345678",
				options: &Options{
					TFCloudToken:    "TOKEN",
					TFCloudEndpoint: "https://app.terraform.io/api/v2",
				},
			},
			mock: func(_ *mocks.Workspaces, StateVersions *mocks.StateVersions) {
				retDownloadURL := "https://archivist.terraform.io/v1/object/test"
				StateVersions.On("Current", mock.Anything, "ws-ABCDEFG12345678").Return(&tfe.StateVersion{DownloadURL: retDownloadURL}, errors.New("unauthorized"))
			},
			wantErr: errors.New("unable to read current state version: unauthorized"),
		},
		{
			name: "Should fail with bad authentication token - full path",
			args: args{
				workspaceID: "some-org/some-workspace",
				options: &Options{
					TFCloudToken:    "TOKEN",
					TFCloudEndpoint: "https://app.terraform.io/api/v2",
				},
			},
			mock: func(Workspaces *mocks.Workspaces, _ *mocks.StateVersions) {
				Workspaces.On("Read", mock.Anything, "some-org", "some-workspace").Return(&tfe.Workspace{ID: "ws-ABCDEFG12345678"}, errors.New("unauthorized"))
			},
			wantErr: errors.New("unable to read terraform workspace id: unauthorized"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewTFCloudReader(tt.args.workspaceID, tt.args.options)

			fakeWorkspaces := &mocks.Workspaces{}
			fakeStateVersions := &mocks.StateVersions{}
			tt.mock(fakeWorkspaces, fakeStateVersions)

			reader.client = &tfe.Client{
				Workspaces:    fakeWorkspaces,
				StateVersions: fakeStateVersions,
			}

			got := make([]byte, len(tt.expected))
			_, err := reader.Read(got)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			assert.NoError(t, err)

			fakeWorkspaces.AssertExpectations(t)
			fakeStateVersions.AssertExpectations(t)
			assert.NotNil(t, got)
			assert.Equal(t, tt.expected, string(got))
		})
	}
}
