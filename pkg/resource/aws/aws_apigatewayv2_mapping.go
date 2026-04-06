package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

// AwsAPIGatewayV2MappingResourceType is the Terraform resource type for aws_apigatewayv2_api_mapping.
const AwsAPIGatewayV2MappingResourceType = "aws_apigatewayv2_api_mapping"

func initAwsAPIGatewayV2MappingMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(
		AwsAPIGatewayV2MappingResourceType,
		func(res *resource.Resource) map[string]string {
			attrs := make(map[string]string)

			if v := res.Attributes().GetString("api_id"); v != nil {
				attrs["Api"] = *v
			}
			if v := res.Attributes().GetString("stage"); v != nil {
				attrs["Stage"] = *v
			}

			return attrs
		},
	)
}
