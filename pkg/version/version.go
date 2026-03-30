// Package version provides build version information and update checking.
package version

import (
	"encoding/json"
	"io"
	"net/http"
	"runtime"

	"github.com/snyk/driftctl/build"

	goversion "github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
)

// version is the current software version, injected at build time.
var version = "dev"

// Current returns the current version string, with a "-dev" suffix for non-release builds.
func Current() string {
	currentVersion := version
	build := build.Build{}
	if !build.IsRelease() {
		currentVersion += "-dev"
	}
	return currentVersion
}

// CheckLatest returns the latest version string if a newer version is available, or "" if current is up to date.
func CheckLatest() string {
	client := &http.Client{}

	req, _ := http.NewRequest(http.MethodGet, "https://telemetry.driftctl.com/version", nil)
	req.Header.Set("driftctl-version", Current())
	req.Header.Set("driftctl-os", runtime.GOOS)
	req.Header.Set("driftctl-arch", runtime.GOARCH)

	res, err := client.Do(req)
	if err != nil {
		logrus.Debugf("Unable to check for a newer version: %+v", err)
		return ""
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		logrus.Debugf("Unable to check for a newer version: %s", res.Status)
		return ""
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Debug("Unable to read response")
		logrus.Debug(err)
		return ""
	}

	responseBody := map[string]string{}
	err = json.Unmarshal(bodyBytes, &responseBody)
	if err != nil {
		logrus.Debug("Unable to decode version check response")
		logrus.Debug(err)
		return ""
	}

	currentVersion, err := goversion.NewVersion(version)
	if err != nil {
		logrus.Debugf("Unable to parse current version: %s", version)
		return ""
	}

	lastVersion, err := goversion.NewVersion(responseBody["latest"])
	if err != nil {
		logrus.Debugf("Unable to parse latest version: %s", responseBody["latest"])
		return ""
	}

	if currentVersion.LessThan(lastVersion) {
		return lastVersion.String()
	}

	return ""
}
