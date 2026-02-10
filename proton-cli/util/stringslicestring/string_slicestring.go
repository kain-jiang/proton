package stringslicestring

import (
	"encoding/json"
	"errors"
)

type StringOrSliceString struct {
	Type Type

	StringValue string

	SliceStringValue []string
}

// FromString creates a StringOrSliceString object with a string value.
func FromString(val string) StringOrSliceString {
	return StringOrSliceString{Type: String, StringValue: val}
}

// FromSliceString creates a StringOrSliceString object with a slice of strings.
func FromSliceString(val []string) StringOrSliceString {
	return StringOrSliceString{Type: SliceString, SliceStringValue: val}
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (sss *StringOrSliceString) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		sss.Type = String
		return json.Unmarshal(value, &sss.StringValue)
	}
	sss.Type = SliceString
	return json.Unmarshal(value, &sss.SliceStringValue)
}

var ErrImposibleType = errors.New("impossible StringOrSliceString.Type")

// MarshalJSON implements the json.Marshaller interface.
func (intstr StringOrSliceString) MarshalJSON() ([]byte, error) {
	switch intstr.Type {
	case String:
		return json.Marshal(intstr.StringValue)
	case SliceString:
		return json.Marshal(intstr.SliceStringValue)
	default:
		return []byte{}, ErrImposibleType
	}
}

// Type represents the stored type of StringOrSliceString.
type Type int64

const (
	String      Type = iota // The StringOrSliceString holds a string.
	SliceString             // The StringOrSliceString holds a slice of strings.
)
