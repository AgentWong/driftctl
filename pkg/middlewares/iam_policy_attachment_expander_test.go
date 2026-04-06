package middlewares

import (
	"fmt"
	"strings"
	"testing"

	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestIamPolicyAttachmentExpander_Execute(t *testing.T) {
	type resources struct {
		RemoteResources    *[]*resource.Resource
		ResourcesFromState *[]*resource.Resource
	}
	tests := []struct {
		name     string
		args     resources
		mocks    func(*dctlresource.MockResourceFactory)
		expected resources
		wantErr  bool
	}{
		{
			name: "Split users and ReId",
			mocks: func(factory *dctlresource.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"jean-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"users":      []interface{}{"jean"},
					},
				).Once().Return(&resource.Resource{
					ID:   "jean-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"paul-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"users":      []interface{}{"paul"},
					},
				).Once().Return(&resource.Resource{
					ID:   "paul-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"pierre-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"users":      []interface{}{"pierre"},
					},
				).Once().Return(&resource.Resource{
					ID:   "pierre-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"jean-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"users":      []interface{}{"jean"},
					},
				).Once().Return(&resource.Resource{
					ID:   "jean-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"paul-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"users":      []interface{}{"paul"},
					},
				).Once().Return(&resource.Resource{
					ID:   "paul-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"jacques-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"users":      []interface{}{"jacques"},
					},
				).Once().Return(&resource.Resource{
					ID:   "jacques-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"jean-fromstatearn",
					map[string]interface{}{
						"policy_arn": "fromstatearn",
						"users":      []interface{}{"jean"},
					},
				).Once().Return(&resource.Resource{
					ID:   "jean-fromstatearn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
			},
			args: struct {
				RemoteResources    *[]*resource.Resource
				ResourcesFromState *[]*resource.Resource
			}{
				RemoteResources: &[]*resource.Resource{
					{
						ID:   "wrongId",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "arn",
							"users":      []interface{}{"jean", "paul", "pierre"},
						},
					},
					{
						ID:   "wrongId2",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "thisisarn",
							"users":      []interface{}{"jean", "paul", "jacques"},
						},
					},
				},
				ResourcesFromState: &[]*resource.Resource{
					{
						ID:   "wrongId",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "fromstatearn",
							"users":      []interface{}{"jean"},
						},
					},
				},
			},
			expected: struct {
				RemoteResources    *[]*resource.Resource
				ResourcesFromState *[]*resource.Resource
			}{
				RemoteResources: &[]*resource.Resource{
					{
						ID:   "jean-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "paul-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "pierre-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "jean-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "paul-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "jacques-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
				},
				ResourcesFromState: &[]*resource.Resource{
					{
						ID:   "jean-fromstatearn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Split Roles and ReId",
			mocks: func(factory *dctlresource.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role1-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"roles":      []interface{}{"role1"},
					},
				).Once().Return(&resource.Resource{
					ID:   "role1-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role2-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"roles":      []interface{}{"role2"},
					},
				).Once().Return(&resource.Resource{
					ID:   "role2-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"pierre-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"roles":      []interface{}{"pierre"},
					},
				).Once().Return(&resource.Resource{
					ID:   "pierre-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role1-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"roles":      []interface{}{"role1"},
					},
				).Once().Return(&resource.Resource{
					ID:   "role1-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role2-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"roles":      []interface{}{"role2"},
					},
				).Once().Return(&resource.Resource{
					ID:   "role2-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role3-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"roles":      []interface{}{"role3"},
					},
				).Once().Return(&resource.Resource{
					ID:   "role3-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role1-fromstatearn",
					map[string]interface{}{
						"policy_arn": "fromstatearn",
						"roles":      []interface{}{"role1"},
					},
				).Once().Return(&resource.Resource{
					ID:   "role1-fromstatearn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
			},
			args: struct {
				RemoteResources    *[]*resource.Resource
				ResourcesFromState *[]*resource.Resource
			}{
				RemoteResources: &[]*resource.Resource{
					{
						ID:   "wrongId",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "arn",
							"roles":      []interface{}{"role1", "role2", "pierre"},
						},
					},
					{
						ID:   "wrongId2",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "thisisarn",
							"roles":      []interface{}{"role1", "role2", "role3"},
						},
					},
				},
				ResourcesFromState: &[]*resource.Resource{
					{
						ID:   "wrongId",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "fromstatearn",
							"roles":      []interface{}{"role1"},
						},
					},
				},
			},
			expected: struct {
				RemoteResources    *[]*resource.Resource
				ResourcesFromState *[]*resource.Resource
			}{
				RemoteResources: &[]*resource.Resource{
					{
						ID:   "role1-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "role2-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "pierre-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "role1-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "role2-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "role3-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
				},
				ResourcesFromState: &[]*resource.Resource{
					{
						ID:   "role1-fromstatearn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Split Groups and ReId",
			mocks: func(factory *dctlresource.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"group1-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"groups":     []interface{}{"group1"},
					},
				).Once().Return(&resource.Resource{
					ID:   "group1-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"group2-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"groups":     []interface{}{"group2"},
					},
				).Once().Return(&resource.Resource{
					ID:   "group2-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"foobar-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"groups":     []interface{}{"foobar"},
					},
				).Once().Return(&resource.Resource{
					ID:   "foobar-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"group1-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"groups":     []interface{}{"group1"},
					},
				).Once().Return(&resource.Resource{
					ID:   "group1-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"group2-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"groups":     []interface{}{"group2"},
					},
				).Once().Return(&resource.Resource{
					ID:   "group2-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"group3-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"groups":     []interface{}{"group3"},
					},
				).Once().Return(&resource.Resource{
					ID:   "group3-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"group1-fromstatearn",
					map[string]interface{}{
						"policy_arn": "fromstatearn",
						"groups":     []interface{}{"group1"},
					},
				).Once().Return(&resource.Resource{
					ID:   "group1-fromstatearn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
			},
			args: struct {
				RemoteResources    *[]*resource.Resource
				ResourcesFromState *[]*resource.Resource
			}{
				RemoteResources: &[]*resource.Resource{
					{
						ID:   "wrongId",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "arn",
							"groups":     []interface{}{"group1", "group2", "foobar"},
						},
					},
					{
						ID:   "wrongId2",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "thisisarn",
							"groups":     []interface{}{"group1", "group2", "group3"},
						},
					},
				},
				ResourcesFromState: &[]*resource.Resource{
					{
						ID:   "wrongId",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "fromstatearn",
							"groups":     []interface{}{"group1"},
						},
					},
				},
			},
			expected: struct {
				RemoteResources    *[]*resource.Resource
				ResourcesFromState *[]*resource.Resource
			}{
				RemoteResources: &[]*resource.Resource{
					{
						ID:   "group1-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "group2-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "foobar-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "group1-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "group2-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						ID:   "group3-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
				},
				ResourcesFromState: &[]*resource.Resource{
					{
						ID:   "group1-fromstatearn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &dctlresource.MockResourceFactory{}
			if tt.mocks != nil {
				tt.mocks(factory)
			}

			m := NewIamPolicyAttachmentExpander(factory)
			if err := m.Execute(tt.args.RemoteResources, tt.args.ResourcesFromState); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			changelog, err := diff.Diff(tt.args, tt.expected)
			if err != nil {
				t.Error(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), fmt.Sprintf("%v", change.From), fmt.Sprintf("%v", change.To))
				}
			}
		})
	}
}
