package remote

import (
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/resource"

	"github.com/sirupsen/logrus"
)

// Scanner discovers remote resources using the registered enumerators.
type Scanner struct {
	remoteLibrary *common.RemoteLibrary
	alerter       alerter.Interface
	filter        enumeration.Filter
}

// NewScanner creates a Scanner.
func NewScanner(remoteLibrary *common.RemoteLibrary, alerter alerter.Interface, filter enumeration.Filter) *Scanner {
	return &Scanner{
		remoteLibrary: remoteLibrary,
		alerter:       alerter,
		filter:        filter,
	}
}

// Resources returns all remote resources discovered by the registered enumerators.
func (s *Scanner) Resources() ([]*resource.Resource, error) {
	var allResources []*resource.Resource

	for _, be := range s.remoteLibrary.GetBulkEnumerators() {
		resources, err := be.Enumerate(s.filter)
		if err != nil {
			err = HandleResourceEnumerationError(err, s.alerter)
			if err != nil {
				return nil, err
			}
			continue
		}
		for _, res := range resources {
			if res != nil {
				logrus.WithFields(logrus.Fields{
					"id":   res.ResourceId(),
					"type": res.ResourceType(),
				}).Debug("Found cloud resource")
				allResources = append(allResources, res)
			}
		}
	}

	return allResources, nil
}
