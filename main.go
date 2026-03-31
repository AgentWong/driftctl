// Package main is the entry point for the driftctl CLI.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/build"
	"github.com/snyk/driftctl/logger"
	"github.com/snyk/driftctl/pkg/cmd"
	cmderrors "github.com/snyk/driftctl/pkg/cmd/errors"
	"github.com/snyk/driftctl/pkg/cmd/scan"
	"github.com/snyk/driftctl/pkg/config"
	"github.com/snyk/driftctl/pkg/version"
)

func init() {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load() // The Original .env
}

func main() {
	os.Exit(run())
}

func run() int {
	config.Init()
	logger.Init()
	build := build.Build{}
	// Check whether driftCTL is run under Snyk CLI
	isSnyk := config.IsSnyk()
	logrus.WithFields(logrus.Fields{
		"isRelease":               fmt.Sprintf("%t", build.IsRelease()),
		"isUsageReportingEnabled": fmt.Sprintf("%t", build.IsUsageReportingEnabled()),
		"version":                 version.Current(),
		"isSnyk":                  fmt.Sprintf("%t", isSnyk),
	}).Debug("Build info")

	// Enable colorization when driftctl is launched under snyk cli (piped)
	if isSnyk {
		color.NoColor = false
	}

	driftctlCmd := cmd.NewDriftctlCmd(build)

	checkVersion := driftctlCmd.ShouldCheckVersion()
	latestVersionChan := make(chan string)
	if checkVersion {
		go func() {
			latestVersion := version.CheckLatest()
			latestVersionChan <- latestVersion
		}()
	}

	if _, err := driftctlCmd.ExecuteC(); err != nil {
		var notInSync cmderrors.InfrastructureNotInSync
		if errors.As(err, &notInSync) {
			return scan.ExitNotInSync
		}
		_, _ = fmt.Fprintln(os.Stderr, color.RedString("%s", err))
		return scan.ExitError
	}

	if checkVersion {
		newVersion := <-latestVersionChan
		if newVersion != "" {
			_, _ = fmt.Fprintln(os.Stderr, "\n\nYour version of driftctl is outdated, please upgrade!")
			_, _ = fmt.Fprintf(os.Stderr, "Current: %s; Latest: %s\n", version.Current(), newVersion)
		}
	}

	return scan.ExitInSync
}
