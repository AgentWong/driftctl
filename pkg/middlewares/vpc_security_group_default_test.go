package middlewares

import (
	"testing"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestDefaultVPCSecurityGroupShouldBeIgnored(t *testing.T) {
	middleware := NewVPCDefaultSecurityGroupSanitizer()
	remoteResources := []*resource.Resource{
		{
			ID:   "sg-test",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "test",
			},
		},
		{
			ID:   "sg-foo",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "foo",
			},
		},
		{
			ID:   "sg-default",
			Type: aws.AwsDefaultSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "default",
			},
		},
	}
	stateResources := []*resource.Resource{
		{
			ID:   "sg-bar",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "bar",
			},
		},
	}
	err := middleware.Execute(&remoteResources, &stateResources)
	if err != nil {
		t.Error(err)
	}
	if len(remoteResources) != 2 {
		t.Error("Default security group was not ignored")
	}
}

func TestDefaultVPCSecurityGroupShouldNotBeIgnoredWhenManaged(t *testing.T) {
	middleware := NewVPCDefaultSecurityGroupSanitizer()
	remoteResources := []*resource.Resource{
		{
			ID:   "sg-test",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "test",
			},
		},
		{
			ID:   "sg-foo",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "foo",
			},
		},
		{
			ID:   "sg-default",
			Type: aws.AwsDefaultSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "default",
			},
		},
	}
	stateResources := []*resource.Resource{
		{
			ID:   "sg-default",
			Type: aws.AwsDefaultSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "default",
			},
		},
	}
	err := middleware.Execute(&remoteResources, &stateResources)
	if err != nil {
		t.Error(err)
	}
	if len(remoteResources) != 3 {
		t.Error("Default security group was ignored")
	}
	managedDefaultSecurityGroup := remoteResources[2]
	if *managedDefaultSecurityGroup.Attrs.GetString("name") != "default" {
		t.Error("Default security group is ignored when it should not be")
	}
}
