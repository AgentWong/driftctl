package categorizer

import "github.com/snyk/driftctl/enumeration/resource"

// CloudFormationCategorizer detects resources managed by CloudFormation stacks
// by looking for the aws:cloudformation:stack-name tag.
type CloudFormationCategorizer struct{}

func NewCloudFormationCategorizer() *CloudFormationCategorizer {
	return &CloudFormationCategorizer{}
}

func (c *CloudFormationCategorizer) Categorize(r *resource.Resource) (Category, bool) {
	attrs := r.Attributes()
	if attrs == nil {
		return "", false
	}

	tags, ok := (*attrs)["tags"]
	if !ok {
		return "", false
	}

	switch t := tags.(type) {
	case map[string]interface{}:
		if _, hasStackTag := t["aws:cloudformation:stack-name"]; hasStackTag {
			return CategoryCloudFormationManaged, true
		}
	case map[string]string:
		if _, hasStackTag := t["aws:cloudformation:stack-name"]; hasStackTag {
			return CategoryCloudFormationManaged, true
		}
	}

	return "", false
}
