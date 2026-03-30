package middlewares

import (
	"testing"

	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/stretchr/testify/assert"
)

func TestAwsRDSClusterInstanceExpander_Execute(t *testing.T) {
	tests := []struct {
		name                    string
		remoteResources         []*resource.Resource
		stateResources          []*resource.Resource
		expectedRemoteResources []*resource.Resource
		expectedStateResources  []*resource.Resource
		mock                    func(factory *dctlresource.MockResourceFactory)
	}{
		{
			name: "should not map any rds cluster instance into db instances",
			remoteResources: []*resource.Resource{
				{
					ID:    "db-0",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "db-1",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			stateResources: []*resource.Resource{},
			expectedRemoteResources: []*resource.Resource{
				{
					ID:    "db-0",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "db-1",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expectedStateResources: []*resource.Resource{},
		},
		{
			name: "should import db instances in state",
			remoteResources: []*resource.Resource{
				{
					ID:    "bucket89713",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "bucket01",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:   "aurora-cluster-demo-0",
					Type: aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{
						"field": "test",
					},
				},
				{
					ID:   "aurora-cluster-demo-1",
					Type: aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{
						"field": "test",
					},
				},
			},
			stateResources: []*resource.Resource{
				{
					ID:    "aurora-cluster-demo-0",
					Type:  aws.AwsRDSClusterInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "aurora-cluster-demo-1",
					Type:  aws.AwsRDSClusterInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expectedRemoteResources: []*resource.Resource{
				{
					ID:    "bucket89713",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "bucket01",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:   "aurora-cluster-demo-0",
					Type: aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{
						"field": "test",
					},
				},
				{
					ID:   "aurora-cluster-demo-1",
					Type: aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{
						"field": "test",
					},
				},
			},
			expectedStateResources: []*resource.Resource{
				{
					ID:   "aurora-cluster-demo-0",
					Type: aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{
						"field": "test",
					},
				},
				{
					ID:   "aurora-cluster-demo-1",
					Type: aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{
						"field": "test",
					},
				},
			},
			mock: func(factory *dctlresource.MockResourceFactory) {
				factory.On("CreateAbstractResource", aws.AwsDbInstanceResourceType, "aurora-cluster-demo-0", map[string]interface{}{"field": "test"}).
					Return(&resource.Resource{
						ID:    "aurora-cluster-demo-0",
						Type:  aws.AwsDbInstanceResourceType,
						Attrs: &resource.Attributes{"field": "test"},
					}).
					Once()

				factory.On("CreateAbstractResource", aws.AwsDbInstanceResourceType, "aurora-cluster-demo-1", map[string]interface{}{"field": "test"}).
					Return(&resource.Resource{
						ID:    "aurora-cluster-demo-1",
						Type:  aws.AwsDbInstanceResourceType,
						Attrs: &resource.Attributes{"field": "test"},
					}).
					Once()
			},
		},
		{
			name: "should find only one db instances in remote",
			remoteResources: []*resource.Resource{
				{
					ID:    "bucket89713",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "bucket01",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "aurora-cluster-demo-0",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			stateResources: []*resource.Resource{
				{
					ID:    "bucket01",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "aurora-cluster-demo-0",
					Type:  aws.AwsRDSClusterInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "aurora-cluster-demo-1",
					Type:  aws.AwsRDSClusterInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expectedRemoteResources: []*resource.Resource{
				{
					ID:    "bucket89713",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "bucket01",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "aurora-cluster-demo-0",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expectedStateResources: []*resource.Resource{
				{
					ID:    "bucket01",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "aurora-cluster-demo-0",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					ID:    "aurora-cluster-demo-1",
					Type:  aws.AwsRDSClusterInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			mock: func(factory *dctlresource.MockResourceFactory) {
				factory.On("CreateAbstractResource", aws.AwsDbInstanceResourceType, "aurora-cluster-demo-0", map[string]interface{}{}).
					Return(&resource.Resource{
						ID:    "aurora-cluster-demo-0",
						Type:  aws.AwsDbInstanceResourceType,
						Attrs: &resource.Attributes{},
					}).
					Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &dctlresource.MockResourceFactory{}
			if tt.mock != nil {
				tt.mock(factory)
			}

			m := NewRDSClusterInstanceExpander(factory)
			err := m.Execute(&tt.remoteResources, &tt.stateResources)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.expectedRemoteResources, tt.remoteResources, "Unexpected remote resources")
			assert.Equal(t, tt.expectedStateResources, tt.stateResources, "Unexpected state resources")
		})
	}
}
