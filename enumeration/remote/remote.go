// Package remote provides remote cloud resource scanning and enumeration.
package remote

import (
	awscfg "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/terraform"
)

var supportedRemotes = []string{
	common.RemoteAWSTerraform,
}

// IsSupported reports whether the given remote identifier is supported.
func IsSupported(remote string) bool {
	for _, r := range supportedRemotes {
		if r == remote {
			return true
		}
	}
	return false
}

// Activate initializes the given remote provider and returns the AWS config
// so callers can create additional AWS service clients (e.g. CloudFormation).
func Activate(remote, version string, alerter alerter.Interface, providerLibrary *terraform.ProviderLibrary, remoteLibrary *common.RemoteLibrary, progress enumeration.ProgressCounter, factory resource.Factory, configDir string) (awscfg.Config, error) {
	switch remote {
	case common.RemoteAWSTerraform:
		return aws.Init(version, alerter, providerLibrary, remoteLibrary, progress, factory, configDir)
	default:
		return awscfg.Config{}, errors.Errorf("unsupported remote '%s'", remote)
	}
}

// GetSupportedRemotes returns the list of supported remote identifiers.
func GetSupportedRemotes() []string {
	return supportedRemotes
}
