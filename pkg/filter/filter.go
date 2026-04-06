package filter

import "github.com/snyk/driftctl/enumeration/resource"

// Filter determines whether resources or types should be ignored.
type Filter interface {
	IsTypeIgnored(ty resource.Type) bool
	IsResourceIgnored(res *resource.Resource) bool
}
