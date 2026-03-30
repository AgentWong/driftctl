package categorizer

import "github.com/snyk/driftctl/enumeration/resource"

// UnsupportedCategorizer detects resources whose Terraform type has no
// corresponding AWS Config resource type, meaning Config-based scanning
// cannot discover them.
type UnsupportedCategorizer struct {
	configSupportedTypes map[string]bool
}

// NewUnsupportedCategorizer creates a new UnsupportedCategorizer with the given set of Config-supported resource types.
func NewUnsupportedCategorizer(configSupportedTypes map[string]bool) *UnsupportedCategorizer {
	return &UnsupportedCategorizer{configSupportedTypes: configSupportedTypes}
}

// Categorize returns CategoryUnsupported if the resource type is not supported by AWS Config.
func (c *UnsupportedCategorizer) Categorize(r *resource.Resource) (Category, bool) {
	if !c.configSupportedTypes[r.ResourceType()] {
		return CategoryUnsupported, true
	}
	return "", false
}
