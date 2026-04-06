// Package backend provides Terraform state backend implementations.
package backend

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/iac/config"
)

var supportedBackends = []string{
	BackendKeyFile,
	BackendKeyS3,
	BackendKeyHTTP,
	BackendKeyHTTPS,
	BackendKeyTFCloud,
}

// Backend represents a readable and closable state backend.
type Backend io.ReadCloser

// Options holds optional configuration for backend readers.
type Options struct {
	Headers         map[string]string
	TFCloudToken    string
	TFCloudEndpoint string
}

// IsSupported reports whether the given backend key is supported.
func IsSupported(backend string) bool {
	for _, b := range supportedBackends {
		if b == backend {
			return true
		}
	}

	return false
}

// GetBackend returns a Backend reader for the given supplier configuration.
func GetBackend(config config.SupplierConfig, opts *Options) (Backend, error) {
	backend := config.Backend

	if !IsSupported(backend) {
		return nil, errors.Errorf("Unsupported backend '%s'", backend)
	}

	switch backend {
	case BackendKeyFile:
		return NewFileReader(config.Path)
	case BackendKeyS3:
		return NewS3Reader(config.Path)
	case BackendKeyHTTP:
		fallthrough
	case BackendKeyHTTPS:
		return NewHTTPReader(&http.Client{}, fmt.Sprintf("%s://%s", config.Backend, config.Path), opts)
	case BackendKeyTFCloud:
		return NewTFCloudReader(config.Path, opts), nil
	default:
		return nil, errors.Errorf("Unsupported backend '%s'", backend)
	}
}

// GetSupportedBackends returns the list of supported backend keys (excluding the default file backend).
func GetSupportedBackends() []string {
	return supportedBackends[1:]
}
