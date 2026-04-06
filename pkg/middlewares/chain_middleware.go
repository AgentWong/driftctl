package middlewares

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
)

// Chain is a sequence of Middleware functions executed in order.
type Chain []Middleware

// NewChain creates a Chain.
func NewChain(middlewares ...Middleware) Chain {
	return middlewares
}

// Execute applies the Chain middleware.
func (c Chain) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	for _, middleware := range c {
		logrus.WithFields(logrus.Fields{
			"middleware": fmt.Sprintf("%T", middleware),
		}).Debug("Starting middleware")
		err := middleware.Execute(remoteResources, resourcesFromState)
		if err != nil {
			return err
		}
	}
	return nil
}
