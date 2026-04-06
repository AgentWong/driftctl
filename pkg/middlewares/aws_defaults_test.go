package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsDefaults_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		assert             func(t *testing.T, remoteResources, _ []*resource.Resource)
	}{
		{
			"ignore default iam roles when they're not managed by IaC",
			[]*resource.Resource{
				{
					ID:   "AWSServiceRoleForSSO",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/sso.amazonaws.com",
					},
				},
				{
					ID:   "OrganizationAccountAccessRole",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/sso.amazonaws.com/",
					},
				},
				{
					ID:   "terraform-20210408093258091700000001",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/",
					},
				},
				{
					ID:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
			},
			[]*resource.Resource{
				{
					ID:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
				{
					ID:   "terraform-20210408093258091700000001",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/",
					},
				},
			},
			func(t *testing.T, remoteResources, _ []*resource.Resource) {
				assert.Len(t, remoteResources, 3)
				for _, remoteResource := range remoteResources {
					if remoteResource.ResourceID() == "AWSServiceRoleForSSO" {
						t.Fatal("AWSServiceRoleForSSO should have been ignored")
					}
				}
			},
		},
		{
			"ignore default iam roles when they're managed by IaC",
			[]*resource.Resource{
				{
					ID:   "AWSServiceRoleForSSO",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path":        "/aws-service-role/sso.amazonaws.com/",
						"description": "test",
					},
				},
				{
					ID:   "OrganizationAccountAccessRole",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/sso.amazonaws.com/",
					},
				},
				{
					ID:   "driftctl_assume_role:driftctl_policy.10",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/",
						"tags": map[string]string{
							"test": "value",
						},
					},
				},
			},
			[]*resource.Resource{
				{
					ID:   "AWSServiceRoleForSSO",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/sso.amazonaws.com/",
					},
				},
				{
					ID:   "OrganizationAccountAccessRole",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/sso.amazonaws.com/",
					},
				},
				{
					ID:   "driftctl_assume_role:driftctl_policy.10",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/",
						"tags": map[string]string{},
					},
				},
			},
			func(t *testing.T, remoteResources, resourcesFromState []*resource.Resource) {
				assert.Len(t, remoteResources, 2)
				assert.Len(t, resourcesFromState, 2)
			},
		},
		{
			"ignore default iam role policies when they're not managed by IaC",
			[]*resource.Resource{
				{
					ID:   "AWSServiceRoleForSSO",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/sso.amazonaws.com",
					},
				},
				{
					ID:   "OrganizationAccountAccessRole",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/sso.amazonaws.com",
					},
				},
				{
					ID:   "AWSServiceRoleForSSO",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role": "AWSServiceRoleForSSO",
					},
				},
				{
					ID:   "OrganizationAccountAccessRole",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role": "OrganizationAccountAccessRole",
					},
				},
				{
					ID:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
			},
			[]*resource.Resource{
				{
					ID:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
			},
			func(t *testing.T, remoteResources, _ []*resource.Resource) {
				assert.Len(t, remoteResources, 3)
				for _, remoteResource := range remoteResources {
					if remoteResource.ResourceID() == "AWSServiceRoleForSSO" &&
						remoteResource.ResourceType() == aws.AwsIamRoleResourceType {
						t.Fatal("AWSServiceRoleForSSO role should have been ignored")
					}
					if remoteResource.ResourceID() == "AWSServiceRoleForSSO" &&
						remoteResource.ResourceType() == aws.AwsIamRolePolicyResourceType {
						t.Fatal("AWSServiceRoleForSSO policy should have been ignored")
					}
				}
			},
		},
		{
			"ignore default iam role policies even when they're managed by IaC",
			[]*resource.Resource{
				{
					ID:   "custom-role",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/sso.amazonaws.com",
					},
				},
				{
					ID:   "OrganizationAccountAccessRole",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/sso.amazonaws.com",
					},
				},
				{
					ID:   "driftctl_assume_role:driftctl_policy.10",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role": "custom-role",
					},
				},
				{
					ID:   "OrganizationAccountAccessRole:AdministratorAccess",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role":        "OrganizationAccountAccessRole",
						"name_prefix": nil,
					},
				},
			},
			[]*resource.Resource{
				{
					ID:   "OrganizationAccountAccessRole:AdministratorAccess",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role":        "OrganizationAccountAccessRole",
						"name_prefix": "tf-",
					},
				},
			},
			func(t *testing.T, remoteResources, _ []*resource.Resource) {
				assert.Len(t, remoteResources, 2)
				for _, remoteResource := range remoteResources {
					if remoteResource.ResourceID() == "OrganizationAccountAccessRole" &&
						remoteResource.ResourceType() == aws.AwsIamRoleResourceType {
						t.Fatal("OrganizationAccountAccessRole role should have been ignored")
					}
					if remoteResource.ResourceID() == "OrganizationAccountAccessRole:AdministratorAccess" &&
						remoteResource.ResourceType() == aws.AwsIamRolePolicyResourceType {
						t.Fatal("OrganizationAccountAccessRole:AdministratorAccess policy should have been ignored")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &AwsDefaults{}
			err := m.Execute(&tt.remoteResources, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}
			tt.assert(t, tt.remoteResources, tt.resourcesFromState)
		})
	}
}
