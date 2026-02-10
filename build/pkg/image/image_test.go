package image

import "testing"

func TestYAML(t *testing.T) {
	for _, r := range protonPackageReferences {
		t.Log(r)
	}
}
