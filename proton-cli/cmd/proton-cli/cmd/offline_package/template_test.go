package offline_package

import (
	"strings"
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
			out, err := renderManifestTemplate(tt.arch, "", "")
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
	if _, err := renderManifestTemplate("ppc64le", "", ""); err == nil {
		t.Fatal("expected error for unsupported architecture")
	}
}

func TestRenderManifestTemplateProtonCLIWithVersion(t *testing.T) {
	out, err := renderManifestTemplate("amd64", "", "0.1.0")
	if err != nil {
		t.Fatalf("renderManifestTemplate returned error: %v", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(out, &m); err != nil {
		t.Fatalf("yaml.Unmarshal returned error: %v", err)
	}

	// Check that proton-cli is in the binaries list
	found := false
	for _, a := range m.Spec.Binaries {
		if a.Name == "proton-cli" {
			found = true
			if a.HTTP == nil {
				t.Fatal("expected proton-cli to have HTTP source")
			}
			expectedURL := "https://github.com/kweaver-ai/proton/releases/download/v0.1.0/proton-cli-linux-amd64.tar.gz"
			if a.HTTP.URL != expectedURL {
				t.Fatalf("unexpected proton-cli URL: got %q, want %q", a.HTTP.URL, expectedURL)
			}
			if a.HTTP.Format != "tar+gzip" {
				t.Fatalf("unexpected proton-cli format: got %q, want %q", a.HTTP.Format, "tar+gzip")
			}
			if a.HTTP.Path != "proton-cli/bin/proton-cli" {
				t.Fatalf("unexpected proton-cli path: got %q, want %q", a.HTTP.Path, "proton-cli/bin/proton-cli")
			}
			break
		}
	}
	if !found {
		t.Fatal("proton-cli not found in binaries list")
	}
}

func TestRenderManifestTemplateProtonCLIWithPath(t *testing.T) {
	testPath := "/usr/local/bin/proton-cli"
	out, err := renderManifestTemplate("arm64", testPath, "")
	if err != nil {
		t.Fatalf("renderManifestTemplate returned error: %v", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(out, &m); err != nil {
		t.Fatalf("yaml.Unmarshal returned error: %v", err)
	}

	// Check that proton-cli is in the binaries list
	found := false
	for _, a := range m.Spec.Binaries {
		if a.Name == "proton-cli" {
			found = true
			if a.File == nil {
				t.Fatal("expected proton-cli to have File source")
			}
			if a.File.Path != testPath {
				t.Fatalf("unexpected proton-cli path: got %q, want %q", a.File.Path, testPath)
			}
			break
		}
	}
	if !found {
		t.Fatal("proton-cli not found in binaries list")
	}
}

func TestRenderManifestTemplateProtonCLIVersionTakesPrecedence(t *testing.T) {
	// When both version and path are set, version should take precedence
	out, err := renderManifestTemplate("amd64", "/some/path", "0.2.0")
	if err != nil {
		t.Fatalf("renderManifestTemplate returned error: %v", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(out, &m); err != nil {
		t.Fatalf("yaml.Unmarshal returned error: %v", err)
	}

	// Check that proton-cli uses HTTP source (version takes precedence)
	found := false
	for _, a := range m.Spec.Binaries {
		if a.Name == "proton-cli" {
			found = true
			if a.HTTP == nil {
				t.Fatal("expected proton-cli to have HTTP source when version is set")
			}
			if a.File != nil {
				t.Fatal("expected proton-cli to NOT have File source when version is set")
			}
			if !strings.Contains(a.HTTP.URL, "v0.2.0") {
				t.Fatalf("expected URL to contain v0.2.0, got %q", a.HTTP.URL)
			}
			break
		}
	}
	if !found {
		t.Fatal("proton-cli not found in binaries list")
	}
}

func TestRenderManifestTemplateNoProtonCLI(t *testing.T) {
	// When neither version nor path is set, proton-cli should not be in the binaries list
	out, err := renderManifestTemplate("amd64", "", "")
	if err != nil {
		t.Fatalf("renderManifestTemplate returned error: %v", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(out, &m); err != nil {
		t.Fatalf("yaml.Unmarshal returned error: %v", err)
	}

	// Check that proton-cli is NOT in the binaries list
	for _, a := range m.Spec.Binaries {
		if a.Name == "proton-cli" {
			t.Fatal("proton-cli should not be in binaries list when neither version nor path is set")
		}
	}
}
