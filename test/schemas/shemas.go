// Package schemas provides helpers for reading and writing Terraform provider schemas used in tests.
package schemas

import (
	"embed"
	gojson "encoding/json"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/terraform/providers"
)

//go:embed */*/schema.json
var fakeSchemaFS embed.FS

// WriteTestSchema writes a provider schema to the embedded filesystem for use in tests.
func WriteTestSchema(schema map[string]providers.Schema, provider, version string) error {
	_, relativeFilePath, _, _ := runtime.Caller(0)
	fileName := path.Join(path.Dir(relativeFilePath), provider, version, "schema.json")
	content, _ := gojson.Marshal(schema)
	err := os.MkdirAll(filepath.Dir(fileName), 0750)
	if err != nil {
		return err
	}
	err = os.WriteFile(fileName, content, 0600)
	if err != nil {
		return err
	}
	return nil
}

// ReadTestSchema reads a provider schema from the embedded filesystem.
func ReadTestSchema(provider, version string) (map[string]providers.Schema, error) {
	content, err := fakeSchemaFS.ReadFile(path.Join(provider, version, "schema.json"))
	if err != nil {
		return nil, err
	}
	var schema map[string]providers.Schema
	if err := gojson.Unmarshal(content, &schema); err != nil {
		return nil, err
	}
	return schema, nil
}
