package categorizer

import (
	"strings"

	"github.com/snyk/driftctl/enumeration/resource"
)

// DefaultResourceCategorizer detects resources that AWS creates automatically
// in every account (default event buses, managed event rules, SSO-reserved
// roles, default KMS aliases, etc.).
type DefaultResourceCategorizer struct{}

func NewDefaultResourceCategorizer() *DefaultResourceCategorizer {
	return &DefaultResourceCategorizer{}
}

func (c *DefaultResourceCategorizer) Categorize(r *resource.Resource) (Category, bool) {
	resType := r.ResourceType()
	resID := r.ResourceId()
	name := resourceName(r)

	switch resType {
	case "aws_cloudwatch_event_bus":
		if isDefaultEventBus(name, resID) {
			return CategoryDefaultResource, true
		}

	case "aws_cloudwatch_event_rule":
		if isDefaultEventRule(name, resID) {
			return CategoryDefaultResource, true
		}

	case "aws_iam_role":
		// AWSReservedSSO_* roles are created by AWS SSO, not user-managed
		if strings.HasPrefix(name, "AWSReservedSSO_") || strings.HasPrefix(resID, "AWSReservedSSO_") {
			return CategoryDefaultResource, true
		}
		// AWSServiceRoleFor* roles are auto-created by AWS services
		if strings.HasPrefix(name, "AWSServiceRoleFor") || strings.HasPrefix(resID, "AWSServiceRoleFor") {
			return CategoryDefaultResource, true
		}

	case "aws_kms_alias":
		// alias/aws/* are AWS-managed default KMS aliases
		if strings.HasPrefix(resID, "alias/aws/") || strings.HasPrefix(name, "alias/aws/") {
			return CategoryDefaultResource, true
		}
	}

	return "", false
}

func resourceName(r *resource.Resource) string {
	attrs := r.Attributes()
	if attrs == nil {
		return ""
	}
	if name, ok := (*attrs)["name"].(string); ok {
		return name
	}
	// Config-enumerated resources store the AWS Config resource name separately
	if name, ok := (*attrs)["config_name"].(string); ok {
		return name
	}
	return ""
}

func isDefaultEventBus(name, id string) bool {
	// AWS creates "default" and regional "IS-LocalEvents-default" buses
	defaults := []string{"default", "IS-LocalEvents-default"}
	for _, d := range defaults {
		if name == d || id == d {
			return true
		}
	}
	return false
}

func isDefaultEventRule(name, id string) bool {
	defaults := []string{"AutoScalingManagedRule"}
	for _, d := range defaults {
		if name == d || id == d {
			return true
		}
	}
	// IS-Tagging-default-* rules are auto-created by AWS tag policies
	if strings.HasPrefix(name, "IS-Tagging-default-") || strings.HasPrefix(id, "IS-Tagging-default-") {
		return true
	}
	return false
}
