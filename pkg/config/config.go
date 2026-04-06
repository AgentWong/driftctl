// Package config provides application-level configuration initialisation and feature flags.
package config

import "github.com/spf13/viper"

// Init initialises application configuration by binding environment variables via viper.
func Init() {
	_ = viper.BindEnv("log_level")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("dctl")
}

// IsSnyk reports whether the binary is running in Snyk-branded mode.
func IsSnyk() bool {
	return viper.GetBool("IS_SNYK")
}
