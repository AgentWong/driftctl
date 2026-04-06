package resource

import (
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform/configs/configschema"
)

// AttributeSchema wraps a Terraform config schema attribute with additional metadata.
type AttributeSchema struct {
	ConfigSchema configschema.Attribute
	JSONString   bool
}

// Flags is a bitfield for resource schema flags.
type Flags uint32

// HasFlag reports whether the given flag is set.
func (f Flags) HasFlag(flag Flags) bool {
	return f&flag != 0
}

// AddFlag sets the given flag.
func (f *Flags) AddFlag(flag Flags) {
	*f |= flag
}

// Schema holds the metadata for a single resource type.
type Schema struct {
	ProviderVersion             *version.Version
	Flags                       Flags
	SchemaVersion               int64
	Attributes                  map[string]AttributeSchema
	NormalizeFunc               func(res *Resource)
	HumanReadableAttributesFunc func(res *Resource) map[string]string
	DiscriminantFunc            func(*Resource, *Resource) bool
}

// IsComputedField reports whether the attribute at the given path is computed.
func (s *Schema) IsComputedField(path []string) bool {
	metadata, exist := s.Attributes[strings.Join(path, ".")]
	if !exist {
		return false
	}
	return metadata.ConfigSchema.Computed
}

// IsJSONStringField reports whether the attribute at the given path is a JSON string.
func (s *Schema) IsJSONStringField(path []string) bool {
	metadata, exist := s.Attributes[strings.Join(path, ".")]
	if !exist {
		return false
	}
	return metadata.JSONString
}
