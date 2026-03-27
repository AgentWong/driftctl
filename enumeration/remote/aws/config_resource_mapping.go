package aws

import "github.com/snyk/driftctl/enumeration/resource"

// ConfigToTerraformMapping maps AWS Config resource type strings to Terraform resource type strings.
// Only types that AWS Config can discover are included; sub-resources (e.g. security group rules,
// route table associations) are not tracked by Config and must still use individual enumerators.
var ConfigToTerraformMapping = map[string]string{
	// EC2
	"AWS::EC2::Instance":              "aws_instance",
	"AWS::EC2::SecurityGroup":         "aws_security_group",
	"AWS::EC2::VPC":                   "aws_vpc",
	"AWS::EC2::Subnet":                "aws_subnet",
	"AWS::EC2::InternetGateway":       "aws_internet_gateway",
	"AWS::EC2::NATGateway":            "aws_nat_gateway",
	"AWS::EC2::RouteTable":            "aws_route_table",
	"AWS::EC2::NetworkInterface":      "aws_network_interface",
	"AWS::EC2::EIP":                   "aws_eip",
	"AWS::EC2::Volume":                "aws_ebs_volume",
	"AWS::EC2::NetworkAcl":            "aws_network_acl",
	"AWS::EC2::LaunchTemplate":        "aws_launch_template",
	"AWS::EC2::FlowLog":               "aws_flow_log",
	"AWS::EC2::VPNGateway":            "aws_vpn_gateway",
	"AWS::EC2::VPNConnection":         "aws_vpn_connection",
	"AWS::EC2::CustomerGateway":       "aws_customer_gateway",
	"AWS::EC2::TransitGateway":        "aws_ec2_transit_gateway",
	"AWS::EC2::TransitGatewayAttachment":         "aws_ec2_transit_gateway_vpc_attachment",
	"AWS::EC2::TransitGatewayRouteTable":         "aws_ec2_transit_gateway_route_table",
	"AWS::EC2::VPCEndpoint":                      "aws_vpc_endpoint",
	"AWS::EC2::VPCEndpointService":               "aws_vpc_endpoint_service",
	"AWS::EC2::VPCPeeringConnection":             "aws_vpc_peering_connection",
	"AWS::EC2::RegisteredHAInstance":              "aws_instance",
	"AWS::EC2::Host":                              "aws_ec2_host",

	// S3
	"AWS::S3::Bucket":           "aws_s3_bucket",
	"AWS::S3::AccountPublicAccessBlock": "aws_s3_account_public_access_block",

	// IAM
	"AWS::IAM::User":   "aws_iam_user",
	"AWS::IAM::Role":   "aws_iam_role",
	"AWS::IAM::Policy": "aws_iam_policy",
	"AWS::IAM::Group":  "aws_iam_group",

	// Lambda
	"AWS::Lambda::Function": "aws_lambda_function",

	// RDS
	"AWS::RDS::DBInstance":     "aws_db_instance",
	"AWS::RDS::DBCluster":      "aws_rds_cluster",
	"AWS::RDS::DBSubnetGroup":  "aws_db_subnet_group",
	"AWS::RDS::DBSnapshot":     "aws_db_snapshot",
	"AWS::RDS::DBClusterSnapshot":     "aws_db_cluster_snapshot",
	"AWS::RDS::EventSubscription":     "aws_db_event_subscription",
	"AWS::RDS::DBSecurityGroup":       "aws_db_security_group",

	// DynamoDB
	"AWS::DynamoDB::Table": "aws_dynamodb_table",

	// SNS
	"AWS::SNS::Topic":        "aws_sns_topic",
	"AWS::SNS::Subscription": "aws_sns_topic_subscription",

	// SQS
	"AWS::SQS::Queue": "aws_sqs_queue",

	// CloudFormation
	"AWS::CloudFormation::Stack": "aws_cloudformation_stack",

	// ELB / ALB / NLB
	"AWS::ElasticLoadBalancing::LoadBalancer":       "aws_elb",
	"AWS::ElasticLoadBalancingV2::LoadBalancer":     "aws_lb",
	"AWS::ElasticLoadBalancingV2::TargetGroup":      "aws_lb_target_group",
	"AWS::ElasticLoadBalancingV2::Listener":         "aws_lb_listener",

	// CloudFront
	"AWS::CloudFront::Distribution":       "aws_cloudfront_distribution",
	"AWS::CloudFront::StreamingDistribution": "aws_cloudfront_distribution",

	// ECR / ECS / EKS
	"AWS::ECR::Repository":            "aws_ecr_repository",
	"AWS::ECS::Cluster":               "aws_ecs_cluster",
	"AWS::ECS::Service":               "aws_ecs_service",
	"AWS::ECS::TaskDefinition":        "aws_ecs_task_definition",
	"AWS::EKS::Cluster":               "aws_eks_cluster",

	// KMS
	"AWS::KMS::Key":   "aws_kms_key",
	"AWS::KMS::Alias": "aws_kms_alias",

	// Route53
	"AWS::Route53::HostedZone": "aws_route53_zone",

	// CloudTrail
	"AWS::CloudTrail::Trail": "aws_cloudtrail",

	// CloudWatch
	"AWS::CloudWatch::Alarm":    "aws_cloudwatch_metric_alarm",
	"AWS::Logs::LogGroup":       "aws_cloudwatch_log_group",

	// Auto Scaling
	"AWS::AutoScaling::AutoScalingGroup":    "aws_autoscaling_group",
	"AWS::AutoScaling::LaunchConfiguration": "aws_launch_configuration",
	"AWS::AutoScaling::ScalingPolicy":       "aws_autoscaling_policy",
	"AWS::AutoScaling::ScheduledAction":     "aws_autoscaling_schedule",

	// ElastiCache
	"AWS::ElastiCache::CacheCluster":      "aws_elasticache_cluster",
	"AWS::ElastiCache::ReplicationGroup":   "aws_elasticache_replication_group",
	"AWS::ElastiCache::CacheSubnetGroup":   "aws_elasticache_subnet_group",

	// Elasticsearch / OpenSearch
	"AWS::Elasticsearch::Domain": "aws_elasticsearch_domain",

	// API Gateway
	"AWS::ApiGateway::RestApi":   "aws_api_gateway_rest_api",
	"AWS::ApiGateway::Stage":     "aws_api_gateway_stage",
	"AWS::ApiGatewayV2::Api":     "aws_apigatewayv2_api",
	"AWS::ApiGatewayV2::Stage":   "aws_apigatewayv2_stage",
	"AWS::ApiGatewayV2::DomainName": "aws_apigatewayv2_domain_name",

	// Redshift
	"AWS::Redshift::Cluster":             "aws_redshift_cluster",
	"AWS::Redshift::ClusterSubnetGroup":  "aws_redshift_subnet_group",
	"AWS::Redshift::ClusterParameterGroup": "aws_redshift_parameter_group",

	// Kinesis
	"AWS::Kinesis::Stream":           "aws_kinesis_stream",
	"AWS::KinesisFirehose::DeliveryStream": "aws_kinesis_firehose_delivery_stream",

	// ACM
	"AWS::ACM::Certificate": "aws_acm_certificate",

	// WAF
	"AWS::WAFv2::WebACL":         "aws_wafv2_web_acl",
	"AWS::WAF::WebACL":           "aws_waf_web_acl",
	"AWS::WAFRegional::WebACL":   "aws_wafregional_web_acl",
	"AWS::WAF::RateBasedRule":    "aws_waf_rate_based_rule",
	"AWS::WAF::Rule":             "aws_waf_rule",
	"AWS::WAF::RuleGroup":        "aws_waf_rule_group",

	// Secrets Manager
	"AWS::SecretsManager::Secret": "aws_secretsmanager_secret",

	// SSM
	"AWS::SSM::ManagedInstanceInventory": "aws_ssm_managed_instance",
	"AWS::SSM::AssociationCompliance":    "aws_ssm_association",
	"AWS::SSM::PatchCompliance":          "aws_ssm_patch_baseline",
	"AWS::SSM::FileData":                 "aws_ssm_document",

	// Config
	"AWS::Config::ResourceCompliance":                  "aws_config_config_rule",
	"AWS::Config::ConformancePackCompliance":            "aws_config_conformance_pack",

	// CodeBuild / CodePipeline
	"AWS::CodeBuild::Project":     "aws_codebuild_project",
	"AWS::CodePipeline::Pipeline": "aws_codepipeline",

	// Step Functions
	"AWS::StepFunctions::StateMachine": "aws_sfn_state_machine",
	"AWS::StepFunctions::Activity":     "aws_sfn_activity",

	// Glue
	"AWS::Glue::Job":      "aws_glue_job",
	"AWS::Glue::Crawler":  "aws_glue_crawler",
	"AWS::Glue::Classifier": "aws_glue_classifier",

	// SageMaker
	"AWS::SageMaker::NotebookInstance": "aws_sagemaker_notebook_instance",
	"AWS::SageMaker::Model":            "aws_sagemaker_model",
	"AWS::SageMaker::EndpointConfig":   "aws_sagemaker_endpoint_configuration",

	// CloudWatch Events / EventBridge
	"AWS::Events::Rule":      "aws_cloudwatch_event_rule",
	"AWS::Events::EventBus":  "aws_cloudwatch_event_bus",

	// Shield / GuardDuty
	"AWS::Shield::Protection":        "aws_shield_protection",
	"AWS::GuardDuty::Detector":       "aws_guardduty_detector",

	// EMR
	"AWS::EMR::SecurityConfiguration": "aws_emr_security_configuration",

	// Elastic Beanstalk
	"AWS::ElasticBeanstalk::Application":   "aws_elastic_beanstalk_application",
	"AWS::ElasticBeanstalk::Environment":   "aws_elastic_beanstalk_environment",

	// DMS
	"AWS::DMS::ReplicationInstance":   "aws_dms_replication_instance",
	"AWS::DMS::ReplicationTask":       "aws_dms_replication_task",
	"AWS::DMS::Certificate":           "aws_dms_certificate",
	"AWS::DMS::EventSubscription":     "aws_dms_event_subscription",

	// EFS
	"AWS::EFS::FileSystem":     "aws_efs_file_system",
	"AWS::EFS::AccessPoint":    "aws_efs_access_point",

	// MSK
	"AWS::MSK::Cluster": "aws_msk_cluster",

	// Backup
	"AWS::Backup::BackupPlan":      "aws_backup_plan",
	"AWS::Backup::BackupSelection": "aws_backup_selection",
	"AWS::Backup::BackupVault":     "aws_backup_vault",

	// XRay
	"AWS::XRay::EncryptionConfig": "aws_xray_encryption_config",

	// Service Catalog / AppStream
	"AWS::ServiceCatalog::CloudFormationProduct":          "aws_servicecatalog_product",
	"AWS::ServiceCatalog::CloudFormationProvisionedProduct": "aws_servicecatalog_provisioned_product",
	"AWS::ServiceCatalog::Portfolio":                       "aws_servicecatalog_portfolio",

	// WorkSpaces
	"AWS::WorkSpaces::Workspace":           "aws_workspaces_workspace",
	"AWS::WorkSpaces::ConnectionAlias":     "aws_workspaces_directory",

	// ECR Public
	"AWS::ECR::PublicRepository": "aws_ecrpublic_repository",

	// Network Firewall
	"AWS::NetworkFirewall::Firewall":       "aws_networkfirewall_firewall",
	"AWS::NetworkFirewall::FirewallPolicy": "aws_networkfirewall_firewall_policy",
	"AWS::NetworkFirewall::RuleGroup":      "aws_networkfirewall_rule_group",

	// Global Accelerator
	"AWS::GlobalAccelerator::Accelerator": "aws_globalaccelerator_accelerator",
	"AWS::GlobalAccelerator::Listener":    "aws_globalaccelerator_listener",
	"AWS::GlobalAccelerator::EndpointGroup": "aws_globalaccelerator_endpoint_group",
}

// terraformToConfigMapping is the reverse mapping, lazily built from ConfigToTerraformMapping.
var terraformToConfigMapping map[string]string

func init() {
	terraformToConfigMapping = make(map[string]string, len(ConfigToTerraformMapping))
	for configType, tfType := range ConfigToTerraformMapping {
		// first-writer-wins: some Config types may map to the same Terraform type
		if _, exists := terraformToConfigMapping[tfType]; !exists {
			terraformToConfigMapping[tfType] = configType
		}
	}
}

func ConfigTypeToTerraformType(configType string) (string, bool) {
	tfType, ok := ConfigToTerraformMapping[configType]
	return tfType, ok
}

func TerraformTypeToConfigType(tfType string) (string, bool) {
	configType, ok := terraformToConfigMapping[tfType]
	return configType, ok
}

// UnsupportedByConfig returns Terraform resource types from the mapping that exist as
// supported driftctl types but have no corresponding AWS Config resource type.
// This helps identify gaps where individual enumerators are still needed.
func UnsupportedByConfig() []string {
	var unsupported []string
	for _, tfType := range ConfigToTerraformMapping {
		if !resource.IsResourceTypeSupported(tfType) {
			unsupported = append(unsupported, tfType)
		}
	}
	return unsupported
}

// ConfigSupportedTerraformTypes returns a set of Terraform resource types that have
// a corresponding AWS Config resource type mapping.
func ConfigSupportedTerraformTypes() map[string]bool {
	result := make(map[string]bool, len(terraformToConfigMapping))
	for tfType := range terraformToConfigMapping {
		result[tfType] = true
	}
	return result
}
