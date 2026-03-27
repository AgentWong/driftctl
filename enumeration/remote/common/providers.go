package common

import (
	tf "github.com/snyk/driftctl/enumeration/terraform"
	"github.com/snyk/driftctl/enumeration/terraform/lock"
)

type RemoteParameter string

const (
	RemoteAWSTerraform = "aws+tf"
)

var remoteParameterMapping = map[RemoteParameter]string{
	RemoteAWSTerraform: tf.AWS,
}

func (p RemoteParameter) GetProviderAddress() *lock.ProviderAddress {
	return &lock.ProviderAddress{
		Hostname:  "registry.terraform.io",
		Namespace: "hashicorp",
		Type:      remoteParameterMapping[p],
	}
}
