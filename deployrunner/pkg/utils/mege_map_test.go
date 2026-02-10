package utils

import "testing"

func TestMergeMaps(t *testing.T) {
	maps := []map[string]interface{}{
		{
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{
					"cpu":    "1m",
					"memory": "1Mi",
				},
			},
		},
		{
			"deps": map[string]interface{}{
				"test1": "test1",
			},
		},
		{
			"namespace": "test",
		},
		{
			"deps": map[string]interface{}{
				"test": "test",
			},
		},
		{
			"namespace": "test1",
		},
	}
	out := MergeMaps(maps...)
	if _, ok := out["namespace"]; !ok {
		t.Fatalf("%#v", out)
	}
	if _, ok := out["resources"]; !ok {
		t.Fatalf("%#v", out)
	}
	if out["namespace"] != "test1" {
		t.Fatalf("%#v", out)
	}
}
