// Package resource defines the core resource types and attribute helpers.
package resource

import (
	"encoding/json"

	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// Deserializer converts raw cty values into abstract Resources.
type Deserializer struct {
	factory Factory
}

// NewDeserializer creates a Deserializer with the given Factory.
func NewDeserializer(factory Factory) *Deserializer {
	return &Deserializer{factory}
}

// Deserialize converts a list of cty values into Resources.
func (s *Deserializer) Deserialize(ty string, rawList []cty.Value) ([]*Resource, error) {
	resources := make([]*Resource, 0)
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

// DeserializeOne converts a single cty value into a Resource.
func (s *Deserializer) DeserializeOne(ty string, value cty.Value) (*Resource, error) {
	if value.IsNull() {
		return nil, nil
	}

	// Marked values cannot be deserialized to JSON.
	// For example, this ensures we can deserialize sensitive values too.
	unmarkedVal, _ := value.UnmarkDeep()

	var attrs Attributes
	bytes, _ := ctyjson.Marshal(unmarkedVal, unmarkedVal.Type())
	err := json.Unmarshal(bytes, &attrs)
	if err != nil {
		return nil, err
	}

	return s.factory.CreateAbstractResource(ty, value.GetAttr("id").AsString(), attrs), nil
}
