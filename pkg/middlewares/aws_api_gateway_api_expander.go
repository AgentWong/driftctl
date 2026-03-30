package middlewares

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsAPIGatewayAPIExpander Explodes the body attribute of api gateway apis v1|v2 to dedicated resources as per Terraform documentations
// (https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/api_gateway_rest_api)
// AwsAPIGatewayAPIExpander (https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/apigatewayv2_api)
type AwsAPIGatewayAPIExpander struct {
	resourceFactory resource.Factory
}

// OpenAPIAwsExtensions is a middleware.
type OpenAPIAwsExtensions struct {
	GatewayResponses map[string]interface{} `json:"x-amazon-apigateway-gateway-responses"`
}

// OpenAPIAwsMethodExtensions is a middleware.
type OpenAPIAwsMethodExtensions struct {
	Integration map[string]interface{} `json:"x-amazon-apigateway-integration"`
}

// NewAwsAPIGatewayAPIExpander creates a AwsAPIGatewayAPIExpander.
func NewAwsAPIGatewayAPIExpander(resourceFactory resource.Factory) AwsAPIGatewayAPIExpander {
	return AwsAPIGatewayAPIExpander{
		resourceFactory: resourceFactory,
	}
}

// Execute applies the AwsAPIGatewayAPIExpander middleware.
func (m AwsAPIGatewayAPIExpander) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newStateResources := make([]*resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than aws_api_gateway_rest_api && aws_apigatewayv2_api
		if res.ResourceType() != aws.AwsAPIGatewayRestAPIResourceType &&
			res.ResourceType() != aws.AwsAPIGatewayV2ApiResourceType {
			newStateResources = append(newStateResources, res)
			continue
		}

		newStateResources = append(newStateResources, res)

		err := m.handleBody(res, &newStateResources, remoteResources)
		if err != nil {
			return err
		}
	}
	*resourcesFromState = newStateResources
	return nil
}

func (m *AwsAPIGatewayAPIExpander) handleBody(api *resource.Resource, results, remoteResources *[]*resource.Resource) error {
	body := api.Attrs.GetString("body")
	if body == nil || *body == "" {
		return nil
	}

	docV3 := &openapi3.T{}
	if err := json.Unmarshal([]byte(*body), &docV3); err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			err = yaml.Unmarshal([]byte(*body), &docV3)
		}
		if err != nil {
			return err
		}
	}
	// It's an OpenAPI v3 document
	if docV3.OpenAPI != "" {
		return m.handleBodyOpenAPIv3(api, docV3, results, remoteResources)
	}

	docV2 := &openapi2.T{}
	if err := json.Unmarshal([]byte(*body), &docV2); err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			err = yaml.Unmarshal([]byte(*body), &docV2)
		}
		if err != nil {
			return err
		}
	}
	// It's an OpenAPI v2 document
	if docV2.Swagger != "" {
		return m.handleBodyOpenAPIv2(api, docV2, results, remoteResources)
	}

	return nil
}

func (m *AwsAPIGatewayAPIExpander) handleBodyOpenAPIv3(api *resource.Resource, doc *openapi3.T, results, remoteResources *[]*resource.Resource) error {
	if api.ResourceType() == aws.AwsAPIGatewayV2ApiResourceType {
		return m.handleBodyOpenAPIv3GatewayV2(api, doc, results, remoteResources)
	}

	apiID := api.ResourceID()
	for path, pathItem := range doc.Paths {
		if res := m.createAPIGatewayResource(apiID, path, results, remoteResources); res != nil {
			ops := pathItem.Operations()
			for httpMethod, method := range ops {
				m.createAPIGatewayMethod(apiID, res.ResourceID(), httpMethod, results)
				for statusCode := range method.Responses {
					m.createAPIGatewayMethodResponse(apiID, res.ResourceID(), httpMethod, statusCode, results)
				}
				m.createAPIGatewayIntegration(apiID, res.ResourceID(), httpMethod, results)
				if err := m.createMethodExtensionsResources(apiID, res.ResourceID(), httpMethod, method.Extensions, results); err != nil {
					return err
				}
			}
		}
	}
	if err := m.createExtensionsResources(apiID, doc.Extensions, results); err != nil {
		return err
	}
	return nil
}

func (m *AwsAPIGatewayAPIExpander) handleBodyOpenAPIv3GatewayV2(api *resource.Resource, doc *openapi3.T, results, remoteResources *[]*resource.Resource) error {
	for path, pathValue := range doc.Paths {
		for method := range doc.Paths[path].Operations() {
			openAPIDerivedRoute := findMatchingOpenAPIDerivedRoute(api.ResourceID(), path, method, remoteResources)
			if openAPIDerivedRoute != nil {
				dummy := m.resourceFactory.CreateAbstractResource(
					aws.AwsAPIGatewayV2RouteResourceType,
					openAPIDerivedRoute.ResourceID(),
					map[string]interface{}{},
				)
				*results = append(*results, dummy)
			}

			for _, operation := range pathValue.Operations() {
				integ, err := decodeMethodExtensions(operation.Extensions)
				if err != nil {
					continue
				}

				openAPIDerivedIntegration := findMatchingOpenAPIDerivedIntegration(api.ResourceID(),
					integ,
					remoteResources)
				if openAPIDerivedIntegration != nil {
					dummy := m.resourceFactory.CreateAbstractResource(
						aws.AwsAPIGatewayV2IntegrationResourceType,
						openAPIDerivedIntegration.ResourceID(),
						map[string]interface{}{},
					)
					*results = append(*results, dummy)
				}
			}
		}
	}
	return nil
}

// The types are similar structurally between the openapi2 and openapi3
// libraries, but without generics we can't really de-dup this witout code
// handleBodyOpenAPIv2GatewayV2 generation, which isn't worth it for this short function.
func (m *AwsAPIGatewayAPIExpander) handleBodyOpenAPIv2GatewayV2(api *resource.Resource, doc *openapi2.T, results, remoteResources *[]*resource.Resource) error {
	for path, pathValue := range doc.Paths {
		for method := range doc.Paths[path].Operations() {
			openAPIDerivedRoute := findMatchingOpenAPIDerivedRoute(api.ResourceID(), path, method, remoteResources)
			if openAPIDerivedRoute != nil {
				dummy := m.resourceFactory.CreateAbstractResource(
					aws.AwsAPIGatewayV2RouteResourceType,
					openAPIDerivedRoute.ResourceID(),
					map[string]interface{}{},
				)
				*results = append(*results, dummy)
			}

			for _, operation := range pathValue.Operations() {
				integ, err := decodeMethodExtensions(operation.Extensions)
				if err != nil {
					continue
				}

				openAPIDerivedIntegration := findMatchingOpenAPIDerivedIntegration(api.ResourceID(),
					integ,
					remoteResources)
				if openAPIDerivedIntegration != nil {
					dummy := m.resourceFactory.CreateAbstractResource(
						aws.AwsAPIGatewayV2IntegrationResourceType,
						openAPIDerivedIntegration.ResourceID(),
						map[string]interface{}{},
					)
					*results = append(*results, dummy)
				}
			}
		}
	}
	return nil
}

func findMatchingOpenAPIDerivedRoute(desiredAPIID, desiredPath, desiredMethod string, remoteResources *[]*resource.Resource) *resource.Resource {
	desiredRouteKey := fmt.Sprintf("%s %s", desiredMethod, desiredPath)
	for _, remoteResource := range *remoteResources {
		if remoteResource.ResourceType() != aws.AwsAPIGatewayV2RouteResourceType {
			continue
		}
		routeKey := *remoteResource.Attributes().GetString("route_key")
		apiID := *remoteResource.Attributes().GetString("api_id")
		if desiredAPIID == apiID && routeKey == desiredRouteKey {
			return remoteResource
		}
	}
	return nil
}

func findMatchingOpenAPIDerivedIntegration(desiredAPIID string, desiredIntegration *OpenAPIAwsMethodExtensions, remoteResources *[]*resource.Resource) *resource.Resource {
	desiredType := desiredIntegration.Integration["type"]
	desiredMethod := desiredIntegration.Integration["httpMethod"]

	if desiredType == nil || desiredMethod == nil {
		return nil
	}

	for _, remoteResource := range *remoteResources {
		if remoteResource.ResourceType() != aws.AwsAPIGatewayV2IntegrationResourceType {
			continue
		}
		apiID := *remoteResource.Attributes().GetString("api_id")
		integrationType := *remoteResource.Attributes().GetString("integration_type")
		if remoteResource.Attributes().GetString("integration_method") == nil {
			// This is nilable in MOCK type only, and they cannot be embedded
			continue
		}
		integrationMethod := *remoteResource.Attributes().GetString("integration_method")
		if desiredAPIID == apiID && integrationType == desiredType && integrationMethod == desiredMethod {
			return remoteResource
		}
	}
	return nil
}

func (m *AwsAPIGatewayAPIExpander) handleBodyOpenAPIv2(api *resource.Resource, doc *openapi2.T, results, remoteResources *[]*resource.Resource) error {
	if api.ResourceType() == aws.AwsAPIGatewayV2ApiResourceType {
		return m.handleBodyOpenAPIv2GatewayV2(api, doc, results, remoteResources)
	}

	apiID := api.ResourceID()
	for path, pathItem := range doc.Paths {
		if res := m.createAPIGatewayResource(apiID, path, results, remoteResources); res != nil {
			ops := pathItem.Operations()
			for httpMethod, method := range ops {
				m.createAPIGatewayMethod(apiID, res.ResourceID(), httpMethod, results)
				for statusCode := range method.Responses {
					m.createAPIGatewayMethodResponse(apiID, res.ResourceID(), httpMethod, statusCode, results)
				}
				m.createAPIGatewayIntegration(apiID, res.ResourceID(), httpMethod, results)
				if err := m.createMethodExtensionsResources(apiID, res.ResourceID(), httpMethod, method.Extensions, results); err != nil {
					return err
				}
			}
		}
	}
	if err := m.createExtensionsResources(apiID, doc.Extensions, results); err != nil {
		return err
	}
	return nil
}

// createExtensionsResources create resources based on our OpenAPIAwsExtensions struct
func (m *AwsAPIGatewayAPIExpander) createExtensionsResources(apiID string, extensions map[string]interface{}, results *[]*resource.Resource) error {
	ext, err := decodeExtensions(extensions)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":   apiID,
			"type": aws.AwsAPIGatewayRestAPIResourceType,
		}).Debug("Failed to decode extensions from the OpenAPI body attribute")
		return err
	}
	for gtwResponse := range ext.GatewayResponses {
		m.createAPIGatewayGatewayResponse(apiID, gtwResponse, results)
	}
	return nil
}

// createMethodExtensionsResources create resources based on our OpenAPIAwsMethodExtensions struct
func (m *AwsAPIGatewayAPIExpander) createMethodExtensionsResources(apiID, resourceID, httpMethod string, extensions map[string]interface{}, results *[]*resource.Resource) error {
	ext, err := decodeMethodExtensions(extensions)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":   apiID,
			"type": aws.AwsAPIGatewayRestAPIResourceType,
		}).Debug("Failed to decode method extensions from the OpenAPI body attribute")
		return err
	}
	if responses, exist := ext.Integration["responses"]; exist {
		for _, response := range responses.(map[string]interface{}) {
			if statusCode, ok := response.(map[string]interface{})["statusCode"]; ok {
				if s, isFloat64 := statusCode.(float64); isFloat64 {
					statusCode = strconv.FormatFloat(s, 'f', -1, 64)
				}
				m.createAPIGatewayIntegrationResponse(apiID, resourceID, httpMethod, statusCode.(string), results)
			}
		}
	}
	return nil
}

// createAPIGatewayResource create aws_api_gateway_resource resource
func (m *AwsAPIGatewayAPIExpander) createAPIGatewayResource(apiID, path string, results, remoteResources *[]*resource.Resource) *resource.Resource {
	if res := foundMatchingResource(apiID, path, remoteResources); res != nil {
		newResource := m.resourceFactory.CreateAbstractResource(aws.AwsAPIGatewayResourceResourceType, res.ResourceID(), map[string]interface{}{
			"rest_api_id": *res.Attributes().GetString("rest_api_id"),
			"path":        path,
		})
		*results = append(*results, newResource)
		return newResource
	}
	return nil
}

// createAPIGatewayGatewayResponse create aws_api_gateway_gateway_response resource
func (m *AwsAPIGatewayAPIExpander) createAPIGatewayGatewayResponse(apiID, gtwResponse string, results *[]*resource.Resource) {
	newResource := m.resourceFactory.CreateAbstractResource(
		aws.AwsAPIGatewayGatewayResponseResourceType,
		strings.Join([]string{"aggr", apiID, gtwResponse}, "-"),
		map[string]interface{}{},
	)
	*results = append(*results, newResource)
}

// foundMatchingResource returns the aws_api_gateway_resource resource that matches the path attribute
func foundMatchingResource(apiID, path string, remoteResources *[]*resource.Resource) *resource.Resource {
	for _, res := range *remoteResources {
		if res.ResourceType() == aws.AwsAPIGatewayResourceResourceType {
			p := res.Attributes().GetString("path")
			i := res.Attributes().GetString("rest_api_id")
			if p != nil && i != nil && *p == path && *i == apiID {
				return res
			}
		}
	}
	return nil
}

// createAPIGatewayMethod create aws_api_gateway_method resource
func (m *AwsAPIGatewayAPIExpander) createAPIGatewayMethod(apiID, resourceID, httpMethod string, results *[]*resource.Resource) {
	newResource := m.resourceFactory.CreateAbstractResource(
		aws.AwsAPIGatewayMethodResourceType,
		strings.Join([]string{"agm", apiID, resourceID, httpMethod}, "-"),
		map[string]interface{}{},
	)
	*results = append(*results, newResource)
}

// createAPIGatewayMethodResponse create aws_api_gateway_method_response resource
func (m *AwsAPIGatewayAPIExpander) createAPIGatewayMethodResponse(apiID, resourceID, httpMethod, statusCode string, results *[]*resource.Resource) {
	newResource := m.resourceFactory.CreateAbstractResource(
		aws.AwsAPIGatewayMethodResponseResourceType,
		strings.Join([]string{"agmr", apiID, resourceID, httpMethod, statusCode}, "-"),
		map[string]interface{}{},
	)
	*results = append(*results, newResource)
}

// Decode openapi.Extensions into our custom OpenAPIAwsExtensions struct that follows AWS
// decodeExtensions openAPI addons.
func decodeExtensions(extensions map[string]interface{}) (*OpenAPIAwsExtensions, error) {
	rawExtensions, err := json.Marshal(extensions)
	if err != nil {
		return nil, err
	}
	decodedExtensions := &OpenAPIAwsExtensions{}
	err = json.Unmarshal(rawExtensions, decodedExtensions)
	if err != nil {
		return nil, err
	}
	return decodedExtensions, nil
}

// createAPIGatewayIntegration create aws_api_gateway_integration resource
func (m *AwsAPIGatewayAPIExpander) createAPIGatewayIntegration(apiID, resourceID, httpMethod string, results *[]*resource.Resource) {
	newResource := m.resourceFactory.CreateAbstractResource(
		aws.AwsAPIGatewayIntegrationResourceType,
		strings.Join([]string{"agi", apiID, resourceID, httpMethod}, "-"),
		map[string]interface{}{},
	)
	*results = append(*results, newResource)
}

// createAPIGatewayIntegrationResponse create aws_api_gateway_integration resource
func (m *AwsAPIGatewayAPIExpander) createAPIGatewayIntegrationResponse(apiID, resourceID, httpMethod, statusCode string, results *[]*resource.Resource) {
	newResource := m.resourceFactory.CreateAbstractResource(
		aws.AwsAPIGatewayIntegrationResponseResourceType,
		strings.Join([]string{"agir", apiID, resourceID, httpMethod, statusCode}, "-"),
		map[string]interface{}{},
	)
	*results = append(*results, newResource)
}

// Decode openapi.Method.Extensions into our custom OpenAPIAwsMethodExtensions struct that follows AWS
// decodeMethodExtensions openAPI addons.
func decodeMethodExtensions(extensions map[string]interface{}) (*OpenAPIAwsMethodExtensions, error) {
	rawExtensions, err := json.Marshal(extensions)
	if err != nil {
		return nil, err
	}
	decodedExtensions := &OpenAPIAwsMethodExtensions{}
	err = json.Unmarshal(rawExtensions, decodedExtensions)
	if err != nil {
		return nil, err
	}
	return decodedExtensions, nil
}
