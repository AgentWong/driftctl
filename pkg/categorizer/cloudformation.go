package categorizer

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

// CloudFormationCategorizer detects resources managed by CloudFormation stacks
// by checking whether the resource's ID appears in the set of physical resource
// IDs returned by the CloudFormation API (ListStackResources).
type CloudFormationCategorizer struct {
	managedIDs map[string]bool
}

// NewCloudFormationCategorizer creates a new CloudFormationCategorizer.
// managedIDs is the set of physical resource IDs from all active CloudFormation stacks.
func NewCloudFormationCategorizer(managedIDs map[string]bool) *CloudFormationCategorizer {
	return &CloudFormationCategorizer{managedIDs: managedIDs}
}

// Categorize returns CategoryCloudFormationManaged if the resource's ID is
// present in the CloudFormation-managed physical resource ID set.
func (c *CloudFormationCategorizer) Categorize(r *resource.Resource) (Category, bool) {
	if len(c.managedIDs) == 0 {
		return "", false
	}
	if c.managedIDs[r.ResourceID()] {
		return CategoryCloudFormationManaged, true
	}
	return "", false
}
