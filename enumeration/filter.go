package enumeration

import "github.com/snyk/driftctl/enumeration/resource"

// Filter decides which resource types and resources should be ignored during enumeration.
type Filter interface {
	IsTypeIgnored(ty resource.Type) bool
	IsResourceIgnored(res *resource.Resource) bool
}
