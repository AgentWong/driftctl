package middlewares

import "github.com/snyk/driftctl/enumeration/resource"

// Middleware transforms resources between scan phases.
type Middleware interface {
	Execute(remoteResources, resourcesFromState *[]*resource.Resource) error
}
