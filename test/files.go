package test

import (
	"os"
	"path"
	"runtime"
)

// WriteTestFile writes content to a file at path p relative to the calling test file.
func WriteTestFile(p string, content []byte) error {
	_, filename, _, _ := runtime.Caller(1)
	return os.WriteFile(path.Join(path.Dir(filename), p), content, 0600)
}

// ReadTestFile reads and returns the content of a file at path p relative to the calling test file.
func ReadTestFile(p string) ([]byte, error) {
	_, filename, _, _ := runtime.Caller(1)
	return os.ReadFile(path.Join(path.Dir(filename), p)) //nolint:gosec // G304: path from test helper callers
}
