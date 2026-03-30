// Package terraform provides a gRPC-based Terraform provider wrapper used for reading remote resources.
package terraform

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/plugin/discovery"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	progress2 "github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/parallel"
	tf "github.com/snyk/driftctl/enumeration/terraform"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// ExitError is the exit code used when the provider fails to configure.
const ExitError = 3

// ProviderConfig holds the configuration for a Terraform provider.
// Aliases namespace gRPC clients (e.g. one per AWS region).
type ProviderConfig struct {
	Name              string
	DefaultAlias      string
	GetProviderConfig func(alias string) interface{}
}

// Provider wraps a set of gRPC plugin.GRPCProvider instances for reading resources.
type Provider struct {
	lock              sync.Mutex
	providerInstaller *tf.ProviderInstaller
	grpcProviders     map[string]*plugin.GRPCProvider
	schemas           map[string]providers.Schema
	Config            ProviderConfig
	runner            *parallel.Runner
	progress          progress2.ProgressCounter
}

// NewProvider creates a new Provider.
func NewProvider(installer *tf.ProviderInstaller, config ProviderConfig, progress progress2.ProgressCounter) (*Provider, error) {
	p := Provider{
		providerInstaller: installer,
		runner:            parallel.NewRunner(context.TODO(), 10),
		grpcProviders:     make(map[string]*plugin.GRPCProvider),
		Config:            config,
		progress:          progress,
	}
	return &p, nil
}

// Init configures the default alias and sets up signal handling for cleanup.
func (p *Provider) Init() error {
	stopCh := make(chan bool)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-c:
			logrus.Warn("Detected interrupt during terraform provider configuration, cleanup ...")
			p.Cleanup()
			os.Exit(ExitError)
		case <-stopCh:
			return
		}
	}()
	defer func() {
		stopCh <- true
	}()
	err := p.configure(p.Config.DefaultAlias)
	if err != nil {
		return err
	}
	return nil
}

// Schema returns the cached provider schemas.
func (p *Provider) Schema() map[string]providers.Schema {
	return p.schemas
}

// Runner returns the parallel runner for concurrent resource reads.
func (p *Provider) Runner() *parallel.Runner {
	return p.runner
}

func (p *Provider) configure(alias string) error {
	providerPath, err := p.providerInstaller.Install()
	if err != nil {
		return err
	}

	if p.grpcProviders[alias] == nil {
		logrus.WithFields(logrus.Fields{
			"alias": alias,
		}).Debug("Starting gRPC client")
		GRPCProvider, err := tf.NewGRPCProvider(discovery.PluginMeta{
			Path: providerPath,
		})

		if err != nil {
			return err
		}
		p.grpcProviders[alias] = GRPCProvider
	}

	schema := p.grpcProviders[alias].GetSchema()
	if p.schemas == nil {
		p.schemas = schema.ResourceTypes
	}

	// This value is optional. It'll be overridden by the provider config.
	config := cty.NullVal(cty.DynamicPseudoType)

	if p.Config.GetProviderConfig != nil {
		configType := schema.Provider.Block.ImpliedType()
		config, err = gocty.ToCtyValue(p.Config.GetProviderConfig(alias), configType)
		if err != nil {
			return err
		}
	}

	resp := p.grpcProviders[alias].Configure(providers.ConfigureRequest{
		Config: config,
	})
	if resp.Diagnostics.HasErrors() {
		return resp.Diagnostics.Err()
	}

	logrus.WithFields(logrus.Fields{
		"alias": alias,
	}).Debug("New gRPC client started")

	logrus.WithFields(logrus.Fields{
		"name":  p.Config.Name,
		"alias": alias,
	}).Debug("Terraform provider initialized")

	return nil
}

// ReadResource reads a single resource from the Terraform provider.
func (p *Provider) ReadResource(args tf.ReadResourceArgs) (*cty.Value, error) {
	logrus.WithFields(logrus.Fields{
		"id":    args.ID,
		"type":  args.Ty,
		"attrs": args.Attributes,
	}).Debugf("Reading cloud resource")

	typ := string(args.Ty)
	state := &terraform.InstanceState{
		ID:         args.ID,
		Attributes: map[string]string{},
	}

	alias := p.Config.DefaultAlias
	if args.Attributes["alias"] != "" {
		alias = args.Attributes["alias"]
		delete(args.Attributes, "alias")
	}

	p.lock.Lock()
	if p.grpcProviders[alias] == nil {
		err := p.configure(alias)
		if err != nil {
			return nil, err
		}
	}
	p.lock.Unlock()

	if len(args.Attributes) > 0 {
		// call to the provider sometimes add and delete field to their attribute this may broke caller so we deep copy attributes
		state.Attributes = make(map[string]string, len(args.Attributes))
		for k, v := range args.Attributes {
			state.Attributes[k] = v
		}
	}

	impliedType := p.schemas[typ].Block.ImpliedType()

	priorState, err := state.AttrsAsObjectValue(impliedType)
	if err != nil {
		return nil, err
	}

	var newState cty.Value
	r := retrier.New(retrier.ConstantBackoff(3, 100*time.Millisecond), nil)

	err = r.Run(func() error {
		resp := p.grpcProviders[alias].ReadResource(providers.ReadResourceRequest{
			TypeName:     typ,
			PriorState:   priorState,
			Private:      []byte{},
			ProviderMeta: cty.NullVal(cty.DynamicPseudoType),
		})
		if resp.Diagnostics.HasErrors() {
			return resp.Diagnostics.Err()
		}
		nonFatalErr := resp.Diagnostics.NonFatalErr()
		if resp.NewState.IsNull() && nonFatalErr != nil {
			return errors.Errorf("state returned by ReadResource is nil: %+v", nonFatalErr)
		}
		newState = resp.NewState
		return nil
	})

	if err != nil {
		return nil, err
	}
	p.progress.Inc()
	return &newState, nil
}

// Cleanup shuts down all gRPC provider clients.
func (p *Provider) Cleanup() {
	for alias, client := range p.grpcProviders {
		logrus.WithFields(logrus.Fields{
			"alias": alias,
		}).Debug("Closing gRPC client")
		_ = client.Close()
	}
}
