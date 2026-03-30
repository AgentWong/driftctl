// Package alerts defines alert types raised during remote resource scanning.
package alerts

import (
	"fmt"
	"strings"

	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"

	"github.com/sirupsen/logrus"
)

// ScanningPhase identifies the phase of scanning where an error occurred.
type ScanningPhase int

// EnumerationPhase is the list phase; DetailsFetchingPhase is the read phase.
const (
	EnumerationPhase ScanningPhase = iota
	DetailsFetchingPhase
)

// RemoteAccessDeniedAlert is raised when AWS returns an access denied error.
type RemoteAccessDeniedAlert struct {
	message       string
	provider      string
	scanningPhase ScanningPhase
	resource      *resource.Resource
}

// NewRemoteAccessDeniedAlert creates a RemoteAccessDeniedAlert.
func NewRemoteAccessDeniedAlert(provider string, scanErr *remoteerror.ResourceScanningError, scanningPhase ScanningPhase) *RemoteAccessDeniedAlert {
	var message string
	switch scanningPhase {
	case EnumerationPhase:
		message = fmt.Sprintf(
			"An error occurred listing %s: listing %s is forbidden: %s",
			scanErr.Resource(),
			scanErr.ListedTypeError(),
			scanErr.RootCause().Error(),
		)
	case DetailsFetchingPhase:
		message = fmt.Sprintf(
			"An error occurred listing %s: reading details of %s is forbidden: %s",
			scanErr.Resource(),
			scanErr.ListedTypeError(),
			scanErr.RootCause().Error(),
		)
	default:
		message = fmt.Sprintf(
			"An error occurred listing %s: %s",
			scanErr.Resource(),
			scanErr.RootCause().Error(),
		)
	}

	var relatedResource *resource.Resource
	resourceFQDNSSplit := strings.SplitN(scanErr.Resource(), ".", 2)
	if len(resourceFQDNSSplit) == 2 {
		relatedResource = &resource.Resource{
			ID:   resourceFQDNSSplit[1],
			Type: resourceFQDNSSplit[0],
		}
	}

	return &RemoteAccessDeniedAlert{message, provider, scanningPhase, relatedResource}
}

// Message returns the human-readable alert message.
func (e *RemoteAccessDeniedAlert) Message() string {
	return e.message
}

// ShouldIgnoreResource reports whether the alerting resource should be skipped.
func (e *RemoteAccessDeniedAlert) ShouldIgnoreResource() bool {
	return true
}

// Resource returns the resource associated with the alert, or nil.
func (e *RemoteAccessDeniedAlert) Resource() *resource.Resource {
	return e.resource
}

// GetProviderMessage returns a provider-specific help message.
func (e *RemoteAccessDeniedAlert) GetProviderMessage() string {
	var message string
	if e.scanningPhase == DetailsFetchingPhase {
		message = "It seems that we got access denied exceptions while reading details of resources.\n"
	}
	if e.scanningPhase == EnumerationPhase {
		message = "It seems that we got access denied exceptions while listing resources.\n"
	}

	switch e.provider {
	case common.RemoteAWSTerraform:
		message += "The latest minimal read-only IAM policy for driftctl is always available here, please update yours: https://docs.driftctl.com/aws/policy"
	default:
		return ""
	}
	return message
}

func sendRemoteAccessDeniedAlert(provider string, alerter alerter.Interface, listError *remoteerror.ResourceScanningError, p ScanningPhase) {
	logrus.WithFields(logrus.Fields{
		"resource":    listError.Resource(),
		"listed_type": listError.ListedTypeError(),
	}).Debugf("Got an access denied error: %+v", listError.Error())
	alerter.SendAlert(listError.Resource(), NewRemoteAccessDeniedAlert(provider, listError, p))
}

// SendEnumerationAlert sends an access denied alert for the enumeration phase.
func SendEnumerationAlert(provider string, alerter alerter.Interface, listError *remoteerror.ResourceScanningError) {
	sendRemoteAccessDeniedAlert(provider, alerter, listError, EnumerationPhase)
}
