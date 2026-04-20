package app

import "testing"

func TestStripISFRdsDatabase_RemovesDatabase(t *testing.T) {
	in := map[string]interface{}{
		"depServices": map[string]interface{}{
			"rds": map[string]interface{}{
				"host":     "mariadb",
				"port":     3306,
				"database": "kweaver",
			},
			"redis": map[string]interface{}{
				"connectType": "sentinel",
			},
		},
		"namespace": "kweaver-ai",
	}

	out := StripISFRdsDatabase(in)

	rds := out["depServices"].(map[string]interface{})["rds"].(map[string]interface{})
	if _, exists := rds["database"]; exists {
		t.Fatalf("expected depServices.rds.database to be removed, got: %v", rds)
	}
	if rds["host"] != "mariadb" {
		t.Fatalf("expected rds.host to be preserved, got: %v", rds["host"])
	}

	// Original map must not be mutated.
	origRds := in["depServices"].(map[string]interface{})["rds"].(map[string]interface{})
	if _, exists := origRds["database"]; !exists {
		t.Fatalf("original values map was mutated; rds.database was removed")
	}

	// Sibling branches preserved.
	if _, ok := out["depServices"].(map[string]interface{})["redis"]; !ok {
		t.Fatalf("expected depServices.redis to be preserved")
	}
	if out["namespace"] != "kweaver-ai" {
		t.Fatalf("expected top-level keys to be preserved")
	}
}

func TestStripISFRdsDatabase_NoDepServices(t *testing.T) {
	in := map[string]interface{}{"namespace": "ns"}
	out := StripISFRdsDatabase(in)
	if out["namespace"] != "ns" {
		t.Fatalf("unexpected output: %v", out)
	}
}

func TestStripISFRdsDatabase_NoRds(t *testing.T) {
	in := map[string]interface{}{
		"depServices": map[string]interface{}{
			"redis": map[string]interface{}{"x": 1},
		},
	}
	out := StripISFRdsDatabase(in)
	if _, ok := out["depServices"].(map[string]interface{})["redis"]; !ok {
		t.Fatalf("expected redis preserved")
	}
}

func TestStripISFRdsDatabase_NilInput(t *testing.T) {
	if out := StripISFRdsDatabase(nil); out != nil {
		t.Fatalf("expected nil for nil input, got: %v", out)
	}
}
