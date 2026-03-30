package aws_test

import (
	"testing"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
	testresource "github.com/snyk/driftctl/test/resource"
	"github.com/stretchr/testify/assert"
)

func TestAWS_Metadata_Flags(t *testing.T) {
	testcases := map[string][]resource.Flags{
		aws.AwsAmiResourceType:                             {},
		aws.AwsAPIGatewayAccountResourceType:               {},
		aws.AwsAPIGatewayAPIKeyResourceType:                {},
		aws.AwsAPIGatewayAuthorizerResourceType:            {},
		aws.AwsAPIGatewayBasePathMappingResourceType:       {},
		aws.AwsAPIGatewayDeploymentResourceType:            {},
		aws.AwsAPIGatewayDomainNameResourceType:            {},
		aws.AwsAPIGatewayGatewayResponseResourceType:       {},
		aws.AwsAPIGatewayIntegrationResourceType:           {},
		aws.AwsAPIGatewayIntegrationResponseResourceType:   {},
		aws.AwsAPIGatewayMethodResourceType:                {},
		aws.AwsAPIGatewayMethodResponseResourceType:        {},
		aws.AwsAPIGatewayMethodSettingsResourceType:        {},
		aws.AwsAPIGatewayModelResourceType:                 {},
		aws.AwsAPIGatewayRequestValidatorResourceType:      {},
		aws.AwsAPIGatewayResourceResourceType:              {},
		aws.AwsAPIGatewayRestAPIResourceType:               {},
		aws.AwsAPIGatewayRestAPIPolicyResourceType:         {},
		aws.AwsAPIGatewayStageResourceType:                 {},
		aws.AwsAPIGatewayVpcLinkResourceType:               {},
		aws.AwsAPIGatewayV2ApiResourceType:                 {},
		aws.AwsAPIGatewayV2RouteResourceType:               {},
		aws.AwsAPIGatewayV2DeploymentResourceType:          {},
		aws.AwsAPIGatewayV2VpcLinkResourceType:             {},
		aws.AwsAPIGatewayV2AuthorizerResourceType:          {},
		aws.AwsAPIGatewayV2RouteResponseResourceType:       {},
		aws.AwsAPIGatewayV2DomainNameResourceType:          {},
		aws.AwsAPIGatewayV2ModelResourceType:               {},
		aws.AwsAPIGatewayV2StageResourceType:               {},
		aws.AwsAPIGatewayV2MappingResourceType:             {},
		aws.AwsAPIGatewayV2IntegrationResourceType:         {},
		aws.AwsAPIGatewayV2IntegrationResponseResourceType: {},
		aws.AwsAppAutoscalingPolicyResourceType:            {},
		aws.AwsAppAutoscalingScheduledActionResourceType:   {},
		aws.AwsAppAutoscalingTargetResourceType:            {},
		aws.AwsCloudformationStackResourceType:             {},
		aws.AwsCloudfrontDistributionResourceType:          {},
		aws.AwsDbInstanceResourceType:                      {},
		aws.AwsDbSubnetGroupResourceType:                   {},
		aws.AwsDefaultNetworkACLResourceType:               {},
		aws.AwsDefaultRouteTableResourceType:               {},
		aws.AwsDefaultSecurityGroupResourceType:            {},
		aws.AwsDefaultSubnetResourceType:                   {},
		aws.AwsDefaultVpcResourceType:                      {},
		aws.AwsDynamodbTableResourceType:                   {},
		aws.AwsEbsEncryptionByDefaultResourceType:          {},
		aws.AwsEbsSnapshotResourceType:                     {},
		aws.AwsEbsVolumeResourceType:                       {},
		aws.AwsEcrRepositoryResourceType:                   {},
		aws.AwsEipResourceType:                             {},
		aws.AwsEipAssociationResourceType:                  {},
		aws.AwsElastiCacheClusterResourceType:              {},
		aws.AwsIamAccessKeyResourceType:                    {},
		aws.AwsIamPolicyResourceType:                       {},
		aws.AwsIamPolicyAttachmentResourceType:             {},
		aws.AwsIamRoleResourceType:                         {},
		aws.AwsIamRolePolicyResourceType:                   {},
		aws.AwsIamRolePolicyAttachmentResourceType:         {},
		aws.AwsIamUserResourceType:                         {},
		aws.AwsIamUserPolicyResourceType:                   {},
		aws.AwsIamUserPolicyAttachmentResourceType:         {},
		aws.AwsIamGroupPolicyResourceType:                  {},
		aws.AwsIamGroupPolicyAttachmentResourceType:        {},
		aws.AwsInstanceResourceType:                        {},
		aws.AwsInternetGatewayResourceType:                 {},
		aws.AwsKeyPairResourceType:                         {},
		aws.AwsKmsAliasResourceType:                        {},
		aws.AwsKmsKeyResourceType:                          {},
		aws.AwsLambdaEventSourceMappingResourceType:        {},
		aws.AwsLambdaFunctionResourceType:                  {},
		aws.AwsNatGatewayResourceType:                      {},
		aws.AwsNetworkACLResourceType:                      {},
		aws.AwsRDSClusterResourceType:                      {},
		aws.AwsRDSClusterInstanceResourceType:              {},
		aws.AwsRouteResourceType:                           {},
		aws.AwsRoute53HealthCheckResourceType:              {},
		aws.AwsRoute53RecordResourceType:                   {},
		aws.AwsRoute53ZoneResourceType:                     {},
		aws.AwsRouteTableResourceType:                      {},
		aws.AwsRouteTableAssociationResourceType:           {},
		aws.AwsS3BucketResourceType:                        {},
		aws.AwsS3BucketAnalyticsConfigurationResourceType:  {},
		aws.AwsS3BucketInventoryResourceType:               {},
		aws.AwsS3BucketMetricResourceType:                  {},
		aws.AwsS3BucketNotificationResourceType:            {},
		aws.AwsS3BucketPolicyResourceType:                  {},
		aws.AwsS3BucketPublicAccessBlockResourceType:       {},
		aws.AwsS3AccountPublicAccessBlockResourceType:      {},
		aws.AwsSecurityGroupResourceType:                   {},
		aws.AwsSnsTopicResourceType:                        {},
		aws.AwsSnsTopicPolicyResourceType:                  {},
		aws.AwsSnsTopicSubscriptionResourceType:            {},
		aws.AwsSqsQueueResourceType:                        {},
		aws.AwsSqsQueuePolicyResourceType:                  {},
		aws.AwsSubnetResourceType:                          {},
		aws.AwsVpcResourceType:                             {},
		aws.AwsSecurityGroupRuleResourceType:               {},
		aws.AwsNetworkACLRuleResourceType:                  {},
		aws.AwsLaunchTemplateResourceType:                  {},
		aws.AwsLaunchConfigurationResourceType:             {},
		aws.AwsLoadBalancerResourceType:                    {},
		aws.AwsApplicationLoadBalancerResourceType:         {},
		aws.AwsClassicLoadBalancerResourceType:             {},
		aws.AwsLoadBalancerListenerResourceType:            {},
		aws.AwsApplicationLoadBalancerListenerResourceType: {},
		aws.AwsIamGroupResourceType:                        {},
		aws.AwsEcrRepositoryPolicyResourceType:             {},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	aws.InitResourcesMetadata(schemaRepository)

	for ty, flags := range testcases {
		t.Run(ty, func(tt *testing.T) {
			sch, exist := schemaRepository.GetSchema(ty)
			assert.True(tt, exist)

			if len(flags) == 0 {
				assert.Equal(tt, resource.Flags(0x0), sch.Flags, "should not have any flag")
				return
			}

			for _, flag := range flags {
				assert.Truef(tt, sch.Flags.HasFlag(flag), "should have given flag %d", flag)
			}
		})
	}
}
