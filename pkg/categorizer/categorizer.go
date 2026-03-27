package categorizer

import "github.com/snyk/driftctl/enumeration/resource"

type Category string

const (
	CategoryManaged              Category = "managed"
	CategoryUnmanaged            Category = "unmanaged"
	CategoryCloudFormationManaged Category = "cloudformation_managed"
	CategoryServiceLinked        Category = "service_linked"
	CategoryUnsupported          Category = "unsupported"
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

func NewChain(categorizers ...Categorizer) *Chain {
	return &Chain{categorizers: categorizers}
}

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
