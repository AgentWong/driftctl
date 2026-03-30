package resource

// Type represents a Terraform resource type string.
type Type string

var supportedTypes = map[string]TypeMeta{
	"aws_ami":                     {},
	"aws_cloudfront_distribution": {},
	"aws_db_instance":             {},
	"aws_db_subnet_group":         {},
	"aws_default_network_acl": {children: []Type{
		"aws_network_acl_rule",
	}},
	"aws_default_route_table": {children: []Type{
		"aws_route",
	}},
	"aws_default_security_group": {children: []Type{
		"aws_security_group_rule",
	}},
	"aws_default_subnet": {},
	"aws_default_vpc": {children: []Type{
		// VPC are used by aws_internet_gateway to determine if internet gateway is the default one in middleware
		"aws_internet_gateway",
	}},
	"aws_dynamodb_table": {},
	"aws_ebs_snapshot":   {},
	"aws_ebs_volume":     {},
	"aws_alb": {children: []Type{
		"aws_lb",
	}},
	"aws_lb":          {},
	"aws_lb_listener": {},
	"aws_alb_listener": {children: []Type{
		"aws_lb_listener",
	}},
	"aws_ebs_encryption_by_default": {},
	"aws_ecr_repository":            {},
	"aws_ecr_repository_policy":     {},
	"aws_eip": {children: []Type{
		"aws_eip_association",
	}},
	"aws_eip_association":       {},
	"aws_iam_access_key":        {},
	"aws_iam_policy":            {},
	"aws_iam_policy_attachment": {},
	"aws_iam_role": {children: []Type{
		"aws_iam_role_policy",
		"aws_iam_policy_attachment",
	}},
	"aws_iam_role_policy": {children: []Type{
		"aws_iam_role_policy_attachment",
	}},
	"aws_iam_role_policy_attachment": {children: []Type{
		"aws_iam_policy_attachment",
	}},
	"aws_iam_group_policy_attachment": {children: []Type{
		"aws_iam_policy_attachment",
	}},
	"aws_iam_user": {children: []Type{
		"aws_iam_user_policy",
	}},
	"aws_iam_user_policy": {children: []Type{
		"aws_iam_user_policy_attachment",
	}},
	"aws_iam_user_policy_attachment": {children: []Type{
		"aws_iam_policy_attachment",
	}},
	"aws_iam_group_policy": {},
	"aws_iam_group":        {},
	"aws_instance": {children: []Type{
		"aws_ebs_volume",
	}},
	"aws_internet_gateway": {children: []Type{
		// This is used to determine internet gateway default rule
		"aws_route",
	}},
	"aws_key_pair":                    {},
	"aws_kms_alias":                   {},
	"aws_kms_key":                     {},
	"aws_lambda_event_source_mapping": {},
	"aws_lambda_function":             {},
	"aws_nat_gateway":                 {},
	"aws_network_acl": {children: []Type{
		"aws_network_acl_rule",
	}},
	"aws_network_acl_rule":     {},
	"aws_route":                {},
	"aws_route53_health_check": {},
	"aws_route53_record":       {},
	"aws_route53_zone":         {},
	"aws_route_table": {children: []Type{
		"aws_route",
	}},
	"aws_route_table_association": {},
	"aws_s3_bucket": {children: []Type{
		"aws_s3_bucket_policy",
	}},
	"aws_s3_bucket_analytics_configuration": {},
	"aws_s3_bucket_inventory":               {},
	"aws_s3_bucket_metric":                  {},
	"aws_s3_bucket_notification":            {},
	"aws_s3_bucket_policy":                  {},
	"aws_s3_bucket_public_access_block":     {},
	"aws_security_group": {children: []Type{
		"aws_security_group_rule",
	}},
	"aws_s3_account_public_access_block": {},
	"aws_security_group_rule":            {},
	"aws_sns_topic": {children: []Type{
		"aws_sns_topic_policy",
	}},
	"aws_sns_topic_policy":       {},
	"aws_sns_topic_subscription": {},
	"aws_sqs_queue": {children: []Type{
		"aws_sqs_queue_policy",
	}},
	"aws_sqs_queue_policy":     {},
	"aws_subnet":               {},
	"aws_vpc":                  {},
	"aws_rds_cluster":          {},
	"aws_cloudformation_stack": {},
	"aws_api_gateway_rest_api": {children: []Type{
		"aws_api_gateway_resource",
		"aws_api_gateway_rest_api_policy",
		"aws_api_gateway_gateway_response",
	}},
	"aws_api_gateway_account":    {},
	"aws_api_gateway_api_key":    {},
	"aws_api_gateway_authorizer": {},
	"aws_api_gateway_deployment": {children: []Type{
		"aws_api_gateway_stage",
	}},
	"aws_api_gateway_stage": {},
	"aws_api_gateway_resource": {children: []Type{
		"aws_api_gateway_method",
		"aws_api_gateway_integration",
	}},
	"aws_api_gateway_domain_name":       {},
	"aws_api_gateway_vpc_link":          {},
	"aws_api_gateway_request_validator": {},
	"aws_api_gateway_rest_api_policy":   {},
	"aws_api_gateway_base_path_mapping": {},
	"aws_api_gateway_model":             {},
	"aws_api_gateway_method": {children: []Type{
		"aws_api_gateway_method_response",
	}},
	"aws_api_gateway_method_response":  {},
	"aws_api_gateway_gateway_response": {},
	"aws_api_gateway_method_settings":  {},
	"aws_api_gateway_integration": {children: []Type{
		"aws_api_gateway_integration_response",
	}},
	"aws_api_gateway_integration_response": {},
	"aws_appautoscaling_target":            {},
	"aws_rds_cluster_instance": {children: []Type{
		"aws_db_instance",
	}},
	"aws_appautoscaling_policy":           {},
	"aws_appautoscaling_scheduled_action": {},
	"aws_apigatewayv2_api": {children: []Type{
		"aws_apigatewayv2_route",
		"aws_apigatewayv2_integration",
	}},
	"aws_apigatewayv2_model":                {},
	"aws_apigatewayv2_stage":                {},
	"aws_apigatewayv2_route_response":       {},
	"aws_apigatewayv2_deployment":           {},
	"aws_apigatewayv2_domain_name":          {},
	"aws_apigatewayv2_api_mapping":          {},
	"aws_apigatewayv2_route":                {},
	"aws_apigatewayv2_vpc_link":             {},
	"aws_apigatewayv2_authorizer":           {},
	"aws_apigatewayv2_integration":          {},
	"aws_apigatewayv2_integration_response": {},
	"aws_launch_template":                   {},
	"aws_launch_configuration":              {},
	"aws_elb":                               {},
	"aws_elasticache_cluster":               {},
	"aws_cloudtrail":                        {},
}

// IsResourceTypeSupported reports whether the given type is in the supported set.
func IsResourceTypeSupported(ty string) bool {
	_, exist := supportedTypes[ty]
	return exist
}

func (ty Type) String() string {
	return string(ty)
}

// GetMeta returns the metadata for the given resource type.
func GetMeta(ty Type) TypeMeta {
	return supportedTypes[ty.String()]
}

// TypeMeta holds metadata about a resource type.
type TypeMeta struct {
	children []Type
}

// GetChildrenTypes returns the child resource types.
func (ty TypeMeta) GetChildrenTypes() []Type {
	return ty.children
}
