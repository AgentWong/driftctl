package remote

import (
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/resource"

	"github.com/sirupsen/logrus"
)

type Scanner struct {
	remoteLibrary *common.RemoteLibrary
	alerter       alerter.AlerterInterface
	filter        enumeration.Filter
}

func NewScanner(remoteLibrary *common.RemoteLibrary, alerter alerter.AlerterInterface, filter enumeration.Filter) *Scanner {
	return &Scanner{
		remoteLibrary: remoteLibrary,
		alerter:       alerter,
		filter:        filter,
	}
}

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
