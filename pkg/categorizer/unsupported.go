package categorizer

import "github.com/snyk/driftctl/enumeration/resource"

// UnsupportedCategorizer detects resources whose Terraform type has no
// corresponding AWS Config resource type, meaning Config-based scanning
// cannot discover them.
type UnsupportedCategorizer struct {
	configSupportedTypes map[string]bool
}

func NewUnsupportedCategorizer(configSupportedTypes map[string]bool) *UnsupportedCategorizer {
	return &UnsupportedCategorizer{configSupportedTypes: configSupportedTypes}
}

func (c *UnsupportedCategorizer) Categorize(r *resource.Resource) (Category, bool) {
	if !c.configSupportedTypes[r.ResourceType()] {
		return CategoryUnsupported, true
	}
	return "", false
}
