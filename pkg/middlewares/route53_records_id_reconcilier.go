package middlewares

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// Route53RecordIDReconcilier Since AWS returns the FQDN as the name of the remote record, we must change the Id of the
// state record to be equivalent (ZoneId_FQDN_Type_SetIdentifier)
// For a TXT record toto for zone example.com with Id 1234
// From AWS provider, we retrieve: 1234_toto.example.com_TXT
// Route53RecordIDReconcilier from Terraform state, we retrieve: 1234_toto_TXT
type Route53RecordIDReconcilier struct{}

// NewRoute53RecordIDReconcilier creates a Route53RecordIDReconcilier.
func NewRoute53RecordIDReconcilier() Route53RecordIDReconcilier {
	return Route53RecordIDReconcilier{}
}

// Execute applies the Route53RecordIDReconcilier middleware.
func (m Route53RecordIDReconcilier) Execute(_, resourcesFromState *[]*resource.Resource) error {

	for _, stateResource := range *resourcesFromState {

		if stateResource.ResourceType() != aws.AwsRoute53RecordResourceType {
			continue
		}

		vars := []string{
			(*stateResource.Attrs)["zone_id"].(string),
			(*stateResource.Attrs)["fqdn"].(string),
			(*stateResource.Attrs)["type"].(string),
		}
		newID := strings.Join(vars, "_")
		if newID != stateResource.Id {
			stateResource.Id = newID
			_ = stateResource.Attrs.SafeSet([]string{"id"}, newID)
			logrus.WithFields(logrus.Fields{
				"old_id": stateResource.ResourceId(),
				"new_id": newID,
			}).Debug("Normalized route53 record ID")
		}
	}

	return nil
}
