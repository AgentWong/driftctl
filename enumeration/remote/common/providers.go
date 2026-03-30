package common

import (
	tf "github.com/snyk/driftctl/enumeration/terraform"
	"github.com/snyk/driftctl/enumeration/terraform/lock"
)

// RemoteParameter identifies a remote provider.
type RemoteParameter string

// RemoteAWSTerraform is the parameter for the AWS Terraform provider.
const (
	RemoteAWSTerraform = "aws+tf"
)

var remoteParameterMapping = map[RemoteParameter]string{
	RemoteAWSTerraform: tf.AWS,
}

// GetProviderAddress returns the registry address for this remote parameter.
func (p RemoteParameter) GetProviderAddress() *lock.ProviderAddress {
	return &lock.ProviderAddress{
		Hostname:  "registry.terraform.io",
		Namespace: "hashicorp",
		Type:      remoteParameterMapping[p],
	}
}
