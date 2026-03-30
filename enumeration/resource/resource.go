package resource

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Source describes where a resource was loaded from.
type Source interface {
	Source() string
	Namespace() string
	InternalName() string
}

// SerializableSource is the JSON representation of a Source.
type SerializableSource struct {
	S    string `json:"source"`
	Ns   string `json:"namespace"`
	Name string `json:"internal_name"`
}

// TerraformStateSource identifies a resource loaded from a Terraform state file.
type TerraformStateSource struct {
	State  string
	Module string
	Name   string
}

// NewTerraformStateSource creates a TerraformStateSource.
func NewTerraformStateSource(state, module, name string) *TerraformStateSource {
	return &TerraformStateSource{state, module, name}
}

// Source returns the state file path.
func (s *TerraformStateSource) Source() string {
	return s.State
}

// Namespace returns the module path.
func (s *TerraformStateSource) Namespace() string {
	return s.Module
}

// InternalName returns the resource name within the module.
func (s *TerraformStateSource) InternalName() string {
	return s.Name
}

// Resource represents a single cloud or IaC resource.
type Resource struct {
	Id     string
	Type   string
	Attrs  *Attributes
	Sch    *Schema `json:"-" diff:"-"`
	Source Source  `json:"-"`
}

// Schema returns the resource's schema.
func (r *Resource) Schema() *Schema {
	return r.Sch
}

// ResourceId returns the resource's unique identifier.
func (r *Resource) ResourceId() string {
	return r.Id
}

// ResourceType returns the Terraform type string.
func (r *Resource) ResourceType() string {
	return r.Type
}

// Attributes returns the resource's attribute map.
func (r *Resource) Attributes() *Attributes {
	return r.Attrs
}

// Src returns the resource's source.
func (r *Resource) Src() Source {
	return r.Source
}

// SourceString returns a human-readable source location.
func (r *Resource) SourceString() string {
	if r.Source == nil {
		return ""
	}
	if r.Source.Namespace() == "" {
		return fmt.Sprintf("%s.%s", r.ResourceType(), r.Source.InternalName())
	}
	return fmt.Sprintf("%s.%s.%s", r.Source.Namespace(), r.ResourceType(), r.Source.InternalName())
}

// DisplayName returns a human-readable name for the resource, falling back to
// config_name attribute, then the Name tag, then empty string.
func (r *Resource) DisplayName() string {
	if r.Attrs == nil {
		return ""
	}
	attrs := *r.Attrs
	if name, ok := attrs["config_name"].(string); ok && name != "" {
		return name
	}
	if tags, ok := attrs["tags"].(map[string]interface{}); ok {
		if name, ok := tags["Name"].(string); ok && name != "" {
			return name
		}
	}
	return ""
}

// Equal reports whether two resources are logically the same.
func (r *Resource) Equal(res *Resource) bool {
	if r.ResourceId() != res.ResourceId() || r.ResourceType() != res.ResourceType() {
		return false
	}

	if r.Schema() != nil && r.Schema().DiscriminantFunc != nil {
		return r.Schema().DiscriminantFunc(r, res)
	}

	return true
}

// Factory creates abstract resources from raw data.
type Factory interface {
	CreateAbstractResource(ty, id string, data map[string]interface{}) *Resource
}

// SerializableResource is the JSON representation of a Resource.
type SerializableResource struct {
	Id                 string              `json:"id"`
	Type               string              `json:"type"`
	Name               string              `json:"name,omitempty"`
	ReadableAttributes map[string]string   `json:"human_readable_attributes,omitempty"`
	Source             *SerializableSource `json:"source,omitempty"`
	Category           string              `json:"category,omitempty"`
}

// NewSerializableResource creates a SerializableResource from a Resource.
func NewSerializableResource(res *Resource) *SerializableResource {
	var src *SerializableSource
	if res.Src() != nil {
		src = &SerializableSource{
			S:    res.Src().Source(),
			Ns:   res.Src().Namespace(),
			Name: res.Src().InternalName(),
		}
	}
	sr := &SerializableResource{
		Id:                 res.ResourceId(),
		Type:               res.ResourceType(),
		Name:               extractResourceName(res),
		ReadableAttributes: formatReadableAttributes(res),
		Source:             src,
	}
	return sr
}

// extractResourceName derives a human-friendly name from Config metadata or tags.
func extractResourceName(res *Resource) string {
	if res.Attrs == nil {
		return ""
	}
	attrs := *res.Attrs
	// prefer the Config-supplied resourceName
	if name, ok := attrs["config_name"].(string); ok && name != "" {
		return name
	}
	// fall back to the "Name" tag
	if tags, ok := attrs["tags"].(map[string]interface{}); ok {
		if name, ok := tags["Name"].(string); ok && name != "" {
			return name
		}
	}
	return ""
}

func formatReadableAttributes(res *Resource) map[string]string {
	if res.Schema() == nil || res.Schema().HumanReadableAttributesFunc == nil {
		return map[string]string{}
	}
	return res.Schema().HumanReadableAttributesFunc(res)
}

// NormalizedResource can normalize itself for state or provider comparison.
type NormalizedResource interface {
	NormalizeForState() (Resource, error)
	NormalizeForProvider() (Resource, error)
}

// Sort orders resources by type then by ID.
func Sort(res []*Resource) []*Resource {
	sort.SliceStable(res, func(i, j int) bool {
		if res[i].ResourceType() != res[j].ResourceType() {
			return res[i].ResourceType() < res[j].ResourceType()
		}
		return res[i].ResourceId() < res[j].ResourceId()
	})
	return res
}

// Attributes is a string-keyed attribute map for a resource.
type Attributes map[string]interface{}

// Copy returns a shallow copy of the attributes.
func (a *Attributes) Copy() *Attributes {
	res := Attributes{}

	for key, value := range *a {
		_ = res.SafeSet([]string{key}, value)
	}

	return &res
}

// Get returns the attribute value at the given path.
func (a *Attributes) Get(path string) (interface{}, bool) {
	val, exist := (*a)[path]
	return val, exist
}

// GetSlice returns the attribute as a slice.
func (a *Attributes) GetSlice(path string) []interface{} {
	val, exist := (*a)[path]
	if !exist {
		return nil
	}
	return val.([]interface{})
}

// GetString returns the attribute as a string pointer.
func (a *Attributes) GetString(path string) *string {
	val, exist := (*a)[path]
	if !exist {
		return nil
	}
	v := val.(string)
	return &v
}

// GetBool returns the attribute as a bool pointer.
func (a *Attributes) GetBool(path string) *bool {
	val, exist := (*a)[path]
	if !exist {
		return nil
	}
	v := val.(bool)
	return &v
}

// GetInt returns the attribute as an int pointer.
func (a *Attributes) GetInt(path string) *int {
	val, exist := (*a)[path]
	if !exist {
		return nil
	}
	if v, isInt := val.(int); isInt {
		return &v
	}
	floatVal := a.GetFloat64(path)
	if val == nil {
		return nil
	}
	v := int(*floatVal)
	return &v
}

// GetFloat64 returns the attribute as a float64 pointer.
func (a *Attributes) GetFloat64(path string) *float64 {
	val, exist := (*a)[path]
	if !exist {
		return nil
	}
	v := val.(float64)
	return &v
}

// GetMap returns the attribute as a map.
func (a *Attributes) GetMap(path string) map[string]interface{} {
	val, exist := (*a)[path]
	if !exist {
		return nil
	}
	return val.(map[string]interface{})
}

// SafeDelete removes the attribute at the given path.
func (a *Attributes) SafeDelete(path []string) {
	for i, key := range path {
		if i == len(path)-1 {
			delete(*a, key)
			return
		}

		v, exists := (*a)[key]
		if !exists {
			return
		}
		m, ok := v.(Attributes)
		if !ok {
			return
		}
		*a = m
	}
}

// SafeSet sets the attribute at the given path.
func (a *Attributes) SafeSet(path []string, value interface{}) error {
	for i, key := range path {
		if i == len(path)-1 {
			(*a)[key] = value
			return nil
		}

		v, exists := (*a)[key]
		if !exists {
			(*a)[key] = map[string]interface{}{}
			v = (*a)[key]
		}

		m, ok := v.(Attributes)
		if !ok {
			return errors.Errorf("Path %s cannot be set: %s is not a nested struct", strings.Join(path, "."), key)
		}
		*a = m
	}
	return errors.New("Error setting value") // should not happen ?
}

// DeleteIfDefault removes the attribute if it holds the zero value for its type.
func (a *Attributes) DeleteIfDefault(path string) {
	val, exist := a.Get(path)
	ty := reflect.TypeOf(val)
	if exist && val == reflect.Zero(ty).Interface() {
		a.SafeDelete([]string{path})
	}
}

func concatenatePath(path, next string) string {
	if path == "" {
		return next
	}
	return strings.Join([]string{path, next}, ".")
}

// SanitizeDefaults removes empty slices, maps, and zero-value fields.
func (a *Attributes) SanitizeDefaults() {
	original := reflect.ValueOf(*a)
	attributesCopy := reflect.New(original.Type()).Elem()
	a.sanitize("", original, attributesCopy)
	*a = attributesCopy.Interface().(Attributes)
}

func (a *Attributes) sanitize(path string, original, dst reflect.Value) bool {
	switch original.Kind() {
	case reflect.Ptr:
		originalValue := original.Elem()
		if !originalValue.IsValid() {
			return false
		}
		dst.Set(reflect.New(originalValue.Type()))
		a.sanitize(path, originalValue, dst.Elem())
	case reflect.Interface:
		// Get rid of the wrapping interface
		originalValue := original.Elem()
		if !originalValue.IsValid() {
			return false
		}
		if originalValue.Kind() == reflect.Slice || originalValue.Kind() == reflect.Map {
			if originalValue.Len() == 0 {
				return false
			}
		}
		// Create a new object. Now new gives us a pointer, but we want the value it
		// points to, so we have to call Elem() to unwrap it
		copyValue := reflect.New(originalValue.Type()).Elem()
		a.sanitize(path, originalValue, copyValue)
		dst.Set(copyValue)

	case reflect.Struct:
		for i := 0; i < original.NumField(); i++ {
			field := original.Field(i)
			a.sanitize(concatenatePath(path, field.String()), field, dst.Field(i))
		}
	case reflect.Slice:
		dst.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i++ {
			a.sanitize(concatenatePath(path, strconv.Itoa(i)), original.Index(i), dst.Index(i))
		}
	case reflect.Map:
		dst.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			created := a.sanitize(concatenatePath(path, key.String()), originalValue, copyValue)
			if created {
				dst.SetMapIndex(key, copyValue)
			}
		}
	default:
		dst.Set(original)
	}
	return true
}
