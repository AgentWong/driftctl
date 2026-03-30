package backend

import (
	"os"
)

// BackendKeyFile is the backend key for local file state.
const BackendKeyFile = ""

// NewFileReader opens a local file as a Backend.
func NewFileReader(path string) (Backend, error) {
	return os.Open(path) //nolint:gosec // G304: path comes from user-specified state file config
}
