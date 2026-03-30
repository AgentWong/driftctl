package mocks

import (
	gojson "encoding/json"
	"errors"
	"fmt"
	"sort"

	terraform2 "github.com/snyk/driftctl/enumeration/terraform"

	"github.com/snyk/driftctl/test/goldenfile"

	"github.com/hashicorp/terraform/providers"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// MockedGoldenTFProvider is a TerraformProvider implementation that reads from and writes to golden files.
type MockedGoldenTFProvider struct {
	name            string
	providerName    string
	providerVersion string
	realProvider    terraform2.Provider
	update          bool
}

// NewMockedGoldenTFProvider creates a MockedGoldenTFProvider that records/replays provider calls via golden files.
func NewMockedGoldenTFProvider(name, providerName, providerVersion string, realProvider terraform2.Provider, update bool) *MockedGoldenTFProvider {
	return &MockedGoldenTFProvider{
		name:            name,
		providerName:    providerName,
		providerVersion: providerVersion,
		realProvider:    realProvider,
		update:          update}
}

// Schema returns the provider schema, reading from or writing to the golden file depending on update mode.
func (m *MockedGoldenTFProvider) Schema() map[string]providers.Schema {
	if m.update {
		schema := m.realProvider.Schema()
		m.writeSchema(schema)
		return schema
	}
	return m.readSchema()
}

// ReadResource reads or records a resource value, using the golden file in replay mode.
func (m *MockedGoldenTFProvider) ReadResource(args terraform2.ReadResourceArgs) (*cty.Value, error) {
	if m.update {
		readResource, err := m.realProvider.ReadResource(args)
		m.writeReadResource(args, readResource, err)
		return readResource, err
	}

	return m.readReadResource(args)
}

func (m *MockedGoldenTFProvider) writeSchema(schema map[string]providers.Schema) {
	marshal, err := gojson.Marshal(schema)
	if err != nil {
		panic(err)
	}
	schemaPath := m.getSchemaPath()
	goldenfile.WriteRootFile(schemaPath, marshal, "schema.json")
}

func (m *MockedGoldenTFProvider) getSchemaPath() string {
	schemaPath := fmt.Sprintf("schemas/%s/%s", m.providerName, m.providerVersion)
	return schemaPath
}

func (m *MockedGoldenTFProvider) readSchema() map[string]providers.Schema {
	schemaPath := m.getSchemaPath()

	content := goldenfile.ReadRootFile(schemaPath, "schema.json")
	var schema map[string]providers.Schema
	if err := gojson.Unmarshal(content, &schema); err != nil {
		panic(err)
	}
	return schema
}

func (m *MockedGoldenTFProvider) writeReadResource(args terraform2.ReadResourceArgs, readResource *cty.Value, err error) {
	var readRes = ReadResource{
		Value: readResource,
		Err:   err,
	}

	marshalled, err := gojson.Marshal(&readRes)
	if err != nil {
		panic(err)
	}
	fileName := getFileName(args)
	goldenfile.WriteFile(m.name, marshalled, fileName)
}

func (m *MockedGoldenTFProvider) readReadResource(args terraform2.ReadResourceArgs) (*cty.Value, error) {
	fileName := getFileName(args)
	// TODO I'm putting this here for compatibility reason...
	if !goldenfile.FileExists(m.name, fileName) {
		fileName = fmt.Sprintf("%s-%s.res.golden.json", args.Ty, args.ID)
	}

	content := goldenfile.ReadFile(m.name, fileName)
	var readRes ReadResource
	if err := gojson.Unmarshal(content, &readRes); err != nil {
		panic(err)
	}
	return readRes.Value, readRes.Err
}

// ReadResource holds the result of a provider ReadResource call for golden file serialisation.
type ReadResource struct {
	Value *cty.Value
	Err   error
}

// UnmarshalJSON deserialises a ReadResource from JSON bytes.
func (m *ReadResource) UnmarshalJSON(bytes []byte) error {
	var unm struct {
		Typ []byte
		Val []byte
		Err *string
	}
	if err := gojson.Unmarshal(bytes, &unm); err != nil {
		return err
	}
	if unm.Typ != nil {
		unmarshalType, err := ctyjson.UnmarshalType(unm.Typ)
		if err != nil {
			return err
		}
		if unm.Val != nil {
			unmarshal, err := ctyjson.Unmarshal(unm.Val, unmarshalType)
			if err != nil {
				return err
			}
			m.Value = &unmarshal
		}
	}
	if unm.Err != nil {
		m.Err = errors.New(*unm.Err)
	}
	return nil
}

// MarshalJSON serialises a ReadResource to JSON bytes.
func (m *ReadResource) MarshalJSON() ([]byte, error) {
	var unm struct {
		Typ []byte
		Val []byte
		Err *string
	}
	if m.Value != nil {
		var err error
		unm.Typ, err = ctyjson.MarshalType(m.Value.Type())
		if err != nil {
			return nil, err
		}
		unm.Val, err = ctyjson.Marshal(*m.Value, m.Value.Type())
		if err != nil {
			return nil, err
		}
	}
	if m.Err != nil {
		e := m.Err.Error()
		unm.Err = &e
	}
	return gojson.Marshal(unm)
}

func getFileName(args terraform2.ReadResourceArgs) string {
	suffix := getFileNameSuffix(args)
	fileName := fmt.Sprintf("%s-%s%s.res.golden.json", args.Ty, args.ID, suffix)
	return fileName
}

func getFileNameSuffix(args terraform2.ReadResourceArgs) string {
	suffix := ""
	keys := make([]string, 0, len(args.Attributes))
	for k := range args.Attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		suffix = fmt.Sprintf("%s-%s", suffix, args.Attributes[k])
	}
	return suffix
}

// Cleanup is a no-op for the mock provider.
func (m MockedGoldenTFProvider) Cleanup() {}

// Name returns the underlying provider's name.
func (m *MockedGoldenTFProvider) Name() string {
	return m.realProvider.Name()
}

// Version returns the underlying provider's version.
func (m *MockedGoldenTFProvider) Version() string {
	return m.realProvider.Version()
}
