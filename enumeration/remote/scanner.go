package remote

import (
	"context"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/parallel"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/resource"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Scanner struct {
	enumeratorRunner *parallel.ParallelRunner
	remoteLibrary    *common.RemoteLibrary
	alerter          alerter.AlerterInterface
	filter           enumeration.Filter
}

func NewScanner(remoteLibrary *common.RemoteLibrary, alerter alerter.AlerterInterface, filter enumeration.Filter) *Scanner {
	return &Scanner{
		enumeratorRunner: parallel.NewParallelRunner(context.TODO(), 10),
		remoteLibrary:    remoteLibrary,
		alerter:          alerter,
		filter:           filter,
	}
}

func (s *Scanner) retrieveRunnerResults(runner *parallel.ParallelRunner) ([]*resource.Resource, error) {
	results := make([]*resource.Resource, 0)
loop:
	for {
		select {
		case resources, ok := <-runner.Read():
			if !ok || resources == nil {
				break loop
			}

			for _, res := range resources.([]*resource.Resource) {
				if res != nil {
					results = append(results, res)
				}
			}
		case <-runner.DoneChan():
			break loop
		}
	}
	return results, runner.Err()
}

func (s *Scanner) scan() ([]*resource.Resource, error) {
	var allResources []*resource.Resource
	coveredTypes := make(map[resource.ResourceType]bool)

	// BulkEnumerators each cover many resource types in a single API call,
	// so we run them first and track which types they handle.
	for _, be := range s.remoteLibrary.GetBulkEnumerators() {
		resources, err := be.Enumerate(s.filter)
		if err != nil {
			err = HandleResourceEnumerationError(err, s.alerter)
			if err != nil {
				return nil, err
			}
			// error was handled (e.g. access denied alert sent); skip this bulk enumerator
			for _, t := range be.SupportedTypes() {
				coveredTypes[t] = true
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
		for _, t := range be.SupportedTypes() {
			coveredTypes[t] = true
		}
	}

	for _, enum := range s.remoteLibrary.Enumerators() {
		// skip types already discovered by a BulkEnumerator
		if coveredTypes[enum.SupportedType()] {
			continue
		}
		if s.filter.IsTypeIgnored(enum.SupportedType()) {
			logrus.WithFields(logrus.Fields{
				"type": enum.SupportedType(),
			}).Debug("Ignored enumeration of resources since it is ignored in filter")
			continue
		}
		enumerator := enum
		s.enumeratorRunner.Run(func() (interface{}, error) {
			resources, err := enumerator.Enumerate()
			if err != nil {
				err := HandleResourceEnumerationError(err, s.alerter)
				if err == nil {
					return []*resource.Resource{}, nil
				}
				return nil, err
			}
			for _, res := range resources {
				if res == nil {
					continue
				}
				logrus.WithFields(logrus.Fields{
					"id":   res.ResourceId(),
					"type": res.ResourceType(),
				}).Debug("Found cloud resource")
			}
			return resources, nil
		})
	}

	enumerationResult, err := s.retrieveRunnerResults(s.enumeratorRunner)
	if err != nil {
		return nil, err
	}

	allResources = append(allResources, enumerationResult...)
	return allResources, nil
}

func (s *Scanner) Resources() ([]*resource.Resource, error) {
	resources, err := s.scan()
	if err != nil {
		return nil, err
	}
	return resources, err
}

func (s *Scanner) Stop() {
	logrus.Debug("Stopping scanner")
	s.enumeratorRunner.Stop(errors.New("interrupted"))
}
