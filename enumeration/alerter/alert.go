// Package alerter provides alert collection and delivery during resource enumeration.
package alerter

import (
	"encoding/json"

	"github.com/snyk/driftctl/enumeration/resource"
)

// Alerts is a map of resource keys to their associated alert slices.
type Alerts map[string][]Alert

// Alert is the interface implemented by all alert types produced during enumeration.
type Alert interface {
	Message() string
	ShouldIgnoreResource() bool
	Resource() *resource.Resource
}

// FakeAlert is a test implementation of Alert with configurable message and ignore behaviour.
type FakeAlert struct {
	Msg            string
	IgnoreResource bool
}

// Message returns the alert message string.
func (f *FakeAlert) Message() string {
	return f.Msg
}

// ShouldIgnoreResource reports whether the alerted resource should be ignored.
func (f *FakeAlert) ShouldIgnoreResource() bool {
	return f.IgnoreResource
}

// Resource returns the resource associated with this alert, or nil.
func (f *FakeAlert) Resource() *resource.Resource {
	return nil
}

// SerializableAlert wraps an Alert to provide JSON marshal/unmarshal support.
type SerializableAlert struct {
	Alert
}

// SerializedAlert is the JSON-serialisable representation of an Alert.
type SerializedAlert struct {
	Msg string `json:"message"`
}

// Message returns the serialized alert's message string.
func (u *SerializedAlert) Message() string {
	return u.Msg
}

// ShouldIgnoreResource always returns false for serialized alerts.
func (u *SerializedAlert) ShouldIgnoreResource() bool {
	return false
}

// Resource returns nil as serialized alerts are not associated with a specific resource.
func (u *SerializedAlert) Resource() *resource.Resource {
	return nil
}

// UnmarshalJSON deserialises a SerializableAlert from JSON bytes.
func (s *SerializableAlert) UnmarshalJSON(bytes []byte) error {
	var res SerializedAlert

	if err := json.Unmarshal(bytes, &res); err != nil {
		return err
	}
	s.Alert = &res
	return nil
}

// MarshalJSON serialises a SerializableAlert to JSON bytes.
func (s *SerializableAlert) MarshalJSON() ([]byte, error) {
	return json.Marshal(SerializedAlert{Msg: s.Message()})
}
