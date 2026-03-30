// Package categorizer classifies AWS resources into categories such as managed, unmanaged, or service-linked.
package categorizer

import "github.com/snyk/driftctl/enumeration/resource"

// Category represents the classification of a resource.
type Category string

const (
	// CategoryManaged indicates a resource that is managed by Terraform or CloudFormation.
	CategoryManaged Category = "managed"
	// CategoryUnmanaged indicates a resource not tracked by any IaC tool.
	CategoryUnmanaged Category = "unmanaged"
	// CategoryCloudFormationManaged indicates a resource managed by CloudFormation.
	CategoryCloudFormationManaged Category = "cloudformation_managed"
	// CategoryServiceLinked indicates a service-linked resource.
	CategoryServiceLinked Category = "service_linked"
	// CategoryDefaultResource indicates an AWS default resource.
	CategoryDefaultResource Category = "default_resource"
	// CategoryUnsupported indicates a resource type not supported for drift detection.
	CategoryUnsupported Category = "unsupported"
)

// Categorizer classifies a resource into a category.
// Returns (category, matched). If matched=false, the next categorizer in the chain is tried.
type Categorizer interface {
	Categorize(r *resource.Resource) (Category, bool)
}

// Chain applies categorizers in order, returning the first match.
type Chain struct {
	categorizers []Categorizer
}

// NewChain creates a Chain of categorizers that are tried in order.
func NewChain(categorizers ...Categorizer) *Chain {
	return &Chain{categorizers: categorizers}
}

// Categorize runs the resource through each categorizer in the chain and returns the first matching category.
func (c *Chain) Categorize(r *resource.Resource) Category {
	for _, cat := range c.categorizers {
		if category, matched := cat.Categorize(r); matched {
			return category
		}
	}
	return CategoryUnmanaged
}

// CategorizeAll applies the chain to each resource and groups them by category.
func CategorizeAll(chain *Chain, resources []*resource.Resource) map[Category][]*resource.Resource {
	result := make(map[Category][]*resource.Resource)
	for _, r := range resources {
		cat := chain.Categorize(r)
		result[cat] = append(result[cat], r)
	}
	return result
}
