// Package output provides scan result output formatters (JSON, HTML, console).
package output

import "fmt"

// Config holds the key and path configuration for a scan output destination.
type Config struct {
	Key  string
	Path string
}

func (o *Config) String() string {
	return fmt.Sprintf("%s://%s", o.Key, o.Path)
}
