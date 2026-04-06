// Package goldenfile provides helpers for reading and writing golden test files.
package goldenfile

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	// GoldenFilePath is the base directory for golden test files.
	GoldenFilePath = "test"
	// ResultsFilename is the default name for golden result files.
	ResultsFilename = "results.golden.json"
)

var (
	_, b, _, _ = runtime.Caller(0)

	// Root folder of this project
	Root = filepath.Join(filepath.Dir(b), "../..")
)

// Update is a flag that specifies the name of a test whose golden file should be updated.
var Update = flag.String("update", "", "name of test to update")

// ReadRootFile reads a golden file from the project root test directory.
func ReadRootFile(p string, name string) []byte {
	p = path.Join(path.Join(Root, GoldenFilePath), p)

	_, err := os.Stat(path.Join(p, ResultsFilename))
	if os.IsNotExist(err) && name == ResultsFilename {
		return []byte("[]")
	}

	content, err := os.ReadFile(fmt.Sprintf("%s%c%s", p, os.PathSeparator, sanitizeName(name)))
	if err != nil {
		panic(err)
	}
	return content
}

// WriteRootFile writes content to a golden file in the project root test directory.
func WriteRootFile(p string, content []byte, name string) {
	output := content

	// Avoid creating golden files for empty results
	if name == ResultsFilename && string(output) == "[]" {
		return
	}

	p = path.Join(path.Join(Root, GoldenFilePath), p)
	if err := os.MkdirAll(p, 0750); err != nil {
		panic(err)
	}
	var indentBuffer bytes.Buffer
	err := json.Indent(&indentBuffer, output, "", " ")
	if err == nil {
		output = indentBuffer.Bytes()
	}
	if err != nil {
		logrus.Error(err)
	}

	if err := os.WriteFile(fmt.Sprintf("%s%c%s", p, os.PathSeparator, sanitizeName(name)), output, 0600); err != nil {
		panic(err)
	}
}

// ReadFile reads a golden file from the relative test directory.
func ReadFile(p string, name string) []byte {
	p = path.Join(GoldenFilePath, p)

	_, err := os.Stat(path.Join(p, ResultsFilename))
	if os.IsNotExist(err) && name == ResultsFilename {
		return []byte("[]")
	}

	content, err := os.ReadFile(fmt.Sprintf("%s%c%s", p, os.PathSeparator, sanitizeName(name)))
	if err != nil {
		panic(err)
	}
	return content
}

// WriteFile writes content to a golden file in the relative test directory.
func WriteFile(p string, content []byte, name string) {
	output := content

	// Avoid creating golden files for empty results
	if name == ResultsFilename && string(output) == "[]" {
		return
	}

	p = path.Join(GoldenFilePath, p)
	if err := os.MkdirAll(p, 0750); err != nil {
		panic(err)
	}
	var indentBuffer bytes.Buffer
	err := json.Indent(&indentBuffer, output, "", " ")
	if err == nil {
		output = indentBuffer.Bytes()
	}
	if err != nil {
		logrus.Error(err)
	}

	if err := os.WriteFile(fmt.Sprintf("%s%c%s", p, os.PathSeparator, sanitizeName(name)), output, 0600); err != nil {
		panic(err)
	}
}

// Remove forbidden characters like / in file name
func sanitizeName(name string) string {
	substitution := "_"
	replacer := strings.NewReplacer(
		"/", substitution,
		"\\", substitution,
		"<", substitution,
		">", substitution,
		":", substitution,
		"\"", substitution,
		"|", substitution,
		"?", substitution,
		"*", substitution,
	)
	return replacer.Replace(name)
}

// FileExists reports whether a golden file exists in the given directory.
func FileExists(dirname, f string) bool {
	fileName := path.Join(GoldenFilePath, dirname, sanitizeName(f))
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
