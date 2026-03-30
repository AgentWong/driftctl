package output

import (
	"encoding/json"
	"os"

	"github.com/snyk/driftctl/pkg/analyser"
)

// JSONOutputType is the key identifying JSON output.
const JSONOutputType = "json"

// JSONOutputExample shows the expected JSON output flag format.
const JSONOutputExample = "json://PATH/TO/FILE.json"

// JSON writes an analysis report as JSON.
type JSON struct {
	path string
}

// NewJSON creates a JSON output writer for the given path.
func NewJSON(path string) *JSON {
	return &JSON{path}
}

func (c *JSON) Write(analysis *analyser.Analysis) error {
	file := os.Stdout
	if !isStdOut(c.path) {
		f, err := os.OpenFile(c.path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		file = f
	}

	json, err := json.MarshalIndent(analysis, "", "\t")
	if err != nil {
		return err
	}
	if _, err := file.Write(json); err != nil {
		return err
	}
	return nil
}
