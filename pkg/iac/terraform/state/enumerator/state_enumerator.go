package enumerator

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend"
)

// StateEnumerator discovers state files from a particular backend.
type StateEnumerator interface {
	Origin() string
	Enumerate() ([]string, error)
}

// GetEnumerator returns the appropriate StateEnumerator for the given config.
func GetEnumerator(config config.SupplierConfig, _ *backend.Options) (StateEnumerator, error) {

	switch config.Backend {
	case backend.BackendKeyFile:
		return NewFileEnumerator(config), nil
	case backend.BackendKeyS3:
		return NewS3Enumerator(config), nil
	}

	logrus.WithFields(logrus.Fields{
		"backend": config.Backend,
	}).Debug("No enumerator for backend")

	return nil, nil
}
