package offline_package

import (
	"testing"

	"sigs.k8s.io/yaml"
)

func TestRenderManifestTemplateArm64(t *testing.T) {
	out, err := renderManifestTemplate("arm64")
	if err != nil {
		t.Fatalf("renderManifestTemplate returned error: %v", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(out, &m); err != nil {
		t.Fatalf("yaml.Unmarshal returned error: %v", err)
	}

	if m.Spec.Architecture != "arm64" {
		t.Fatalf("unexpected architecture %q", m.Spec.Architecture)
	}
	if got := m.Spec.Binaries[0].HTTP.URL; got != "https://github.com/projectcalico/calico/releases/download/v3.25.2/calicoctl-linux-arm64" {
		t.Fatalf("unexpected binary URL %q", got)
	}
	if got := m.Spec.Binaries[1].HTTP.Path; got != "linux-arm64/helm" {
		t.Fatalf("unexpected helm path %q", got)
	}
	if got := m.Spec.RPMs[0].Name; got != "containerd-1.7.25.aarch64.rpm" {
		t.Fatalf("unexpected rpm name %q", got)
	}
	if got := m.Spec.RPMs[1].HTTP.URL; got != "https://github.com/kweaver-ai/proton/releases/download/cri-tools-1.35.0/cri-tools-1.35.0-150500.1.1.aarch64.rpm" {
		t.Fatalf("unexpected rpm URL %q", got)
	}
}

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
