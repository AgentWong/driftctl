package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsAPIGatewayV2ModelResourceType is the Terraform resource type for aws_apigatewayv2_model.
const AwsAPIGatewayV2ModelResourceType = "aws_apigatewayv2_model"

func initAwsAPIGatewayV2ModelMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(
		AwsAPIGatewayV2ModelResourceType,
		func(res *resource.Resource) map[string]string {
			return map[string]string{
				"name": *res.Attributes().GetString("name"),
			}
		},
	)
}
