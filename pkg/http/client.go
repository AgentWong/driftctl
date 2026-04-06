// Package http provides HTTP client interfaces used across the application.
package http

import "net/http"

// HTTPClient is an interface for http.Client type.
type HTTPClient interface { //nolint:revive
	Do(req *http.Request) (*http.Response, error)
}
