package backend

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	pkghttp "github.com/snyk/driftctl/pkg/http"

	"io"
	"net/http"
)

// BackendKeyHTTP is the backend key for HTTP state.
const BackendKeyHTTP = "http"

// BackendKeyHTTPS is the backend key for HTTPS state.
const BackendKeyHTTPS = "https"

// HTTPBackend reads Terraform state over HTTP(S).
type HTTPBackend struct {
	request *http.Request
	client  pkghttp.HTTPClient
	reader  io.ReadCloser
}

// NewHTTPReader creates an HTTPBackend that will fetch state from rawURL.
func NewHTTPReader(client pkghttp.HTTPClient, rawURL string, opts *Options) (*HTTPBackend, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range opts.Headers {
		req.Header.Add(key, value)
	}

	return &HTTPBackend{req, client, nil}, nil
}

func (h *HTTPBackend) Read(p []byte) (n int, err error) {
	if h.reader == nil {
		res, err := h.client.Do(h.request) //nolint:bodyclose // body lifecycle is managed by the struct via Close()
		if err != nil {
			return 0, err
		}
		h.reader = res.Body

		if res.StatusCode < 200 || res.StatusCode >= 400 {
			body, _ := io.ReadAll(h.reader)
			logrus.WithFields(logrus.Fields{"body": string(body)}).Trace("HTTP(s) backend response")

			return 0, errors.Errorf("error requesting HTTP(s) backend state: status code: %d", res.StatusCode)
		}
	}
	return h.reader.Read(p)
}

// Close releases the underlying HTTP response body.
func (h *HTTPBackend) Close() error {
	if h.reader != nil {
		return h.reader.Close()
	}
	return errors.New("Unable to close reader as nothing was opened")
}
