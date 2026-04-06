package terraform

import (
	"os/exec"

	tfplugin "github.com/hashicorp/terraform/plugin"
	"github.com/snyk/driftctl/logger"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform/plugin/discovery"
)

// ClientConfig returns the plugin client configuration for launching a Terraform provider process.
func ClientConfig(m discovery.PluginMeta) *plugin.ClientConfig {
	logger := logger.NewTerraformPluginLogger()
	return &plugin.ClientConfig{
		Cmd:              exec.Command(m.Path), //nolint:gosec // G204: m.Path is a trusted provider binary from plugin discovery
		HandshakeConfig:  tfplugin.Handshake,
		VersionedPlugins: tfplugin.VersionedPlugins,
		Managed:          true,
		Logger:           logger,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		AutoMTLS:         true,
	}
}

// Client returns a plugin client for the plugin described by the given metadata.
func Client(m discovery.PluginMeta) *plugin.Client {
	return plugin.NewClient(ClientConfig(m))
}
