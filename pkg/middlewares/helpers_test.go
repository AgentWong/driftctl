package middlewares

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func readFile(t *testing.T, name string) string {
	content, err := os.ReadFile(name) //nolint:gosec // G304: test helper reading test fixtures
	require.Nil(t, err)
	return string(content)
}
