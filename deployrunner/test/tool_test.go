package test

import (
	"fmt"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestRerender(t *testing.T) {
	tt := TestingT{T: t}
	values := map[string]string{
		"test": "{{.Values.namespace}}",
	}
	bs, err := yaml.Marshal(values)
	tt.AssertNil(err)
	values = make(map[string]string)
	err = yaml.Unmarshal(bs, &values)
	tt.AssertNil(err)
	fmt.Println(string(bs), values)
}
