package terraform

import (
	"fmt"
	"runtime"
)

// ProviderConfig holds the key, version, and config directory for a Terraform provider.
type ProviderConfig struct {
	Key       string
	Version   string
	ConfigDir string
}

// GetDownloadURL returns the release download URL for this provider.
func (c *ProviderConfig) GetDownloadURL() string {
	return fmt.Sprintf(
		"https://releases.hashicorp.com/terraform-provider-%s/%s/terraform-provider-%s_%s_%s_%s.zip",
		c.Key,
		c.Version,
		c.Key,
		c.Version,
		runtime.GOOS,
		runtime.GOARCH,
	)
}

// GetBinaryName returns the expected binary name for this provider.
func (c *ProviderConfig) GetBinaryName() string {
	return fmt.Sprintf("terraform-provider-%s_v%s", c.Key, c.Version)
}
