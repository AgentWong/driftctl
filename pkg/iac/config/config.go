// Package config defines configuration types for IaC state suppliers.
package config

import "fmt"

// SupplierConfig holds the configuration for an IaC state supplier.
type SupplierConfig struct {
	Key     string
	Backend string
	Path    string
}

func (c *SupplierConfig) String() string {
	str := c.Key
	if c.Backend != "" {
		str += fmt.Sprintf("+%s", c.Backend)
	}
	if str != "" {
		str += "://"
	}
	if c.Path != "" {
		str += c.Path
	}
	return str
}
