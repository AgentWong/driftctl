// Package envproxy provides helpers for proxying environment variables between prefixes.
package envproxy

import (
	"os"
	"strings"
)

// EnvProxy proxies environment variables by renaming them from one prefix to another.
type EnvProxy struct {
	fromPrefix string
	toPrefix   string
	defaultEnv map[string]string
}

// NewEnvProxy creates a new EnvProxy that maps environment variables from fromPrefix to toPrefix.
func NewEnvProxy(fromPrefix, toPrefix string) *EnvProxy {
	envMap := map[string]string{}
	for _, variable := range os.Environ() {
		tmp := strings.SplitN(variable, "=", 2)
		envMap[tmp[0]] = tmp[1]
	}
	return &EnvProxy{
		fromPrefix: fromPrefix,
		toPrefix:   toPrefix,
		defaultEnv: envMap,
	}
}

// Apply sets environment variables by renaming those matching fromPrefix to use toPrefix.
func (s *EnvProxy) Apply() {
	if s.fromPrefix == "" || s.toPrefix == "" {
		return
	}
	for key, value := range s.defaultEnv {
		if strings.HasPrefix(key, s.fromPrefix) {
			key = strings.Replace(key, s.fromPrefix, s.toPrefix, 1)
			_ = os.Setenv(key, value)
		}
	}
}

// Restore resets environment variables to their original values before Apply was called.
func (s *EnvProxy) Restore() {
	if s.fromPrefix == "" || s.toPrefix == "" {
		return
	}
	for key, value := range s.defaultEnv {
		if strings.HasPrefix(key, s.fromPrefix) {
			key = strings.Replace(key, s.fromPrefix, s.toPrefix, 1)
			value = s.defaultEnv[key]
		}
		_ = os.Setenv(key, value)
	}
}
