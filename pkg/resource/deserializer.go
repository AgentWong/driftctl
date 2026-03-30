// Package resource provides resource factories, schemas, deserialization, and type metadata.
package resource

import (
	"encoding/json"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// Deserializer converts cty values into resource instances.
type Deserializer struct {
	factory resource.Factory
}

// NewDeserializer creates a Deserializer backed by the given factory.
func NewDeserializer(factory resource.Factory) *Deserializer {
	return &Deserializer{factory}
}

// Deserialize converts a list of cty values into resource instances of the given type.
func (s *Deserializer) Deserialize(ty string, rawList []cty.Value) ([]*resource.Resource, error) {
	resources := make([]*resource.Resource, 0)
	for _, rawRes := range rawList {
		rawResource := rawRes
		res, err := s.DeserializeOne(ty, rawResource)
		if err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

// DeserializeOne converts a single cty value into a resource.
func (s *Deserializer) DeserializeOne(ty string, value cty.Value) (*resource.Resource, error) {
	if value.IsNull() {
		return nil, nil
	}

	// Marked values cannot be deserialized to JSON.
	// For example, this ensures we can deserialize sensitive values too.
	unmarkedVal, _ := value.UnmarkDeep()

	var attrs resource.Attributes
	bytes, _ := ctyjson.Marshal(unmarkedVal, unmarkedVal.Type())
	err := json.Unmarshal(bytes, &attrs)
	if err != nil {
		return nil, err
	}

	return s.factory.CreateAbstractResource(ty, value.GetAttr("id").AsString(), attrs), nil
}
