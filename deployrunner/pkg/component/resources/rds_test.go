package resources

import (
	"testing"

	"taskrunner/test"
	"taskrunner/trait"
)

func TestRDSSChema(t *testing.T) {
	// TODO LOAD test case
	tt := test.TestingT{T: t}
	com := trait.ComponentMeta{RawConfigSchema: _RdsSchema}
	err := com.ValidateConfig(map[string]interface{}{
		"source_type": "Proton_MariaDB",
		"password":    "test",
		"user":        "test",
	})
	tt.AssertNil(err)

	err = com.ValidateConfig(map[string]interface{}{})
	if err == nil {
		t.Fatal("no input required config")
	}
}
