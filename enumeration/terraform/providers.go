package terraform

import (
	"github.com/sirupsen/logrus"
)

const (
	// AWS is the provider key for Amazon Web Services.
	AWS string = "aws"
)

// ProviderLibrary stores registered Terraform providers by name.
type ProviderLibrary struct {
	providers map[string]Provider
}

// NewProviderLibrary creates an empty ProviderLibrary.
func NewProviderLibrary() *ProviderLibrary {
	logrus.Debug("New provider library created")
	return &ProviderLibrary{
		make(map[string]Provider),
	}
}

// AddProvider registers a provider under the given name.
func (p *ProviderLibrary) AddProvider(name string, provider Provider) {
	p.providers[name] = provider
}

// Provider returns the provider registered under the given name.
func (p *ProviderLibrary) Provider(name string) Provider {
	return p.providers[name]
}

// Cleanup shuts down all registered providers.
func (p *ProviderLibrary) Cleanup() {
	logrus.Debug("Closing providers")
	for providerKey, provider := range p.providers {
		logrus.WithFields(logrus.Fields{
			"key": providerKey,
		}).Debug("Closing provider")
		provider.Cleanup()
	}
}
