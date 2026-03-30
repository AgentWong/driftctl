package aws

import (
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/terraform"
)

/**
 * Initialize remote (configure credentials, launch tf providers and start gRPC clients)
 * Required to use Scanner
 */

// Init configures the AWS provider and registers its enumerators with the remote library.
func Init(version string, _ alerter.Interface, providerLibrary *terraform.ProviderLibrary, remoteLibrary *common.RemoteLibrary, progress enumeration.ProgressCounter, factory resource.Factory, configDir string) error {

	provider, err := NewTerraformProvider(version, progress, configDir)
	if err != nil {
		return err
	}
	err = provider.CheckCredentialsExist()
	if err != nil {
		return err
	}
	err = provider.Init()
	if err != nil {
		return err
	}

	repositoryCache := cache.New(100)

	providerLibrary.AddProvider(terraform.AWS, provider)

	configRepo := repository.NewConfigRepository(provider.AwsCfg, repositoryCache)
	configEnumerator := NewConfigEnumerator(configRepo, factory)
	remoteLibrary.AddBulkEnumerator(configEnumerator)

	return nil
}
