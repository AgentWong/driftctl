package categorizer

import (
	"strings"

	"github.com/snyk/driftctl/enumeration/resource"
)

// ServiceLinkedCategorizer detects AWS service-linked roles, which are
// automatically created by AWS services and show up as false-positive unmanaged resources.
type ServiceLinkedCategorizer struct{}

// NewServiceLinkedCategorizer creates a new ServiceLinkedCategorizer.
func NewServiceLinkedCategorizer() *ServiceLinkedCategorizer {
	return &ServiceLinkedCategorizer{}
}

// Categorize returns CategoryServiceLinked if the resource is an AWS service-linked role.
func (c *ServiceLinkedCategorizer) Categorize(r *resource.Resource) (Category, bool) {
	if r.ResourceType() != "aws_iam_role" {
		return "", false
	}

	if strings.Contains(r.ResourceId(), "/aws-service-role/") {
		return CategoryServiceLinked, true
	}

	attrs := r.Attributes()
	if attrs == nil {
		return "", false
	}

	if path, ok := (*attrs)["path"].(string); ok {
		if strings.HasPrefix(path, "/aws-service-role/") {
			return CategoryServiceLinked, true
		}
	}

	return "", false
}
