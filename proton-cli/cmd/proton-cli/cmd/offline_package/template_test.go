package offline_package

import (
	"testing"

	"sigs.k8s.io/yaml"
)

func TestRenderManifestTemplateArchitectureAliases(t *testing.T) {
	tests := []struct {
		name string
		arch string
		want string
	}{
		{name: "amd64", arch: "amd64", want: "architecture: amd64"},
		{name: "x86_64", arch: "x86_64", want: "architecture: amd64"},
		{name: "arm64", arch: "arm64", want: "architecture: arm64"},
		{name: "aarch64", arch: "aarch64", want: "architecture: arm64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := renderManifestTemplate(tt.arch)
			if err != nil {
				t.Fatalf("renderManifestTemplate returned error: %v", err)
			}

			var m Manifest
			if err := yaml.Unmarshal(out, &m); err != nil {
				t.Fatalf("yaml.Unmarshal returned error: %v", err)
			}
			if got := "architecture: " + m.Spec.Architecture; got != tt.want {
				t.Fatalf("unexpected architecture %q", got)
			}
		})
	}
}

func TestRenderManifestTemplateInvalidArchitecture(t *testing.T) {
	if _, err := renderManifestTemplate("ppc64le"); err == nil {
		t.Fatal("expected error for unsupported architecture")
	}
}
