package stringslicestring

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/go-test/deep"
)

func TestFromString(t *testing.T) {
	i := FromString("76")
	if i.Type != String || i.StringValue != "76" {
		t.Errorf("Expected StringValue=\"76\", got %+v", i)
	}
}

func TestFromSliceString(t *testing.T) {
	i := FromSliceString([]string{"a", "b", "c"})
	if i.Type != SliceString || deep.Equal(i.SliceStringValue, []string{"a", "b", "c"}) != nil {
		t.Errorf("Expected IntVal=93, got %+v", i)
	}
}

type StringOrSliceStringHolder struct {
	SOrSS StringOrSliceString `json:"val"`
}

func TestStringOrSliceStringUnmarshalJSON(t *testing.T) {
	cases := []struct {
		input  string
		result StringOrSliceString
	}{
		{"{\"val\": [\"a\",\"b\",\"c\"]}", FromSliceString([]string{"a", "b", "c"})},
		{"{\"val\": \"123\"}", FromString("123")},
	}

	for _, c := range cases {
		var result StringOrSliceStringHolder
		if err := json.Unmarshal([]byte(c.input), &result); err != nil {
			t.Errorf("Failed to unmarshal input '%v': %v", c.input, err)
		}
		if diff := deep.Equal(result.SOrSS, c.result); diff != nil {
			for _, d := range diff {
				t.Errorf("Failed to unmarshal input '%v': %v", c.input, d)
			}
		}
	}
}

func TestStringOrSliceStringMarshalJSON(t *testing.T) {
	cases := []struct {
		input  StringOrSliceString
		result string
		err    error
	}{
		{FromSliceString([]string{"a", "b", "c"}), `{"val":["a","b","c"]}`, nil},
		{FromString("abc"), `{"val":"abc"}`, nil},
		{StringOrSliceString{Type: 3}, "", ErrImposibleType},
	}

	for _, c := range cases {
		input := StringOrSliceStringHolder{c.input}
		result, err := json.Marshal(&input)
		if !errors.Is(err, c.err) {
			t.Errorf("Failed to marshal input '%v': want %v, got %v", input, c.err, err)
		}
		if string(result) != c.result {
			t.Errorf("Failed to marshal input '%v': expected: %v, got %v", input, c.result, string(result))
		}
	}
}
