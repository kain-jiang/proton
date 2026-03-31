package offline_package

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sigs.k8s.io/yaml"
)

func TestNormalizeAppPlatform(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "", want: "linux/amd64"},
		{in: "linux/amd64", want: "linux/amd64"},
		{in: "amd64", want: "linux/amd64"},
		{in: "x86_64", want: "linux/amd64"},
		{in: "linux/arm64", want: "linux/arm64"},
		{in: "arm64", want: "linux/arm64"},
		{in: "aarch64", want: "linux/arm64"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := normalizeAppPlatform(tt.in)
			if err != nil {
				t.Fatalf("normalizeAppPlatform returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeAppPlatformInvalid(t *testing.T) {
	if _, err := normalizeAppPlatform("ppc64le"); err == nil {
		t.Fatal("expected error for invalid platform")
	}
}

func TestNormalizeAppPlatforms(t *testing.T) {
	got, err := normalizeAppPlatforms("linux/amd64,arm64,x86_64")
	if err != nil {
		t.Fatalf("normalizeAppPlatforms returned error: %v", err)
	}
	if len(got) != 2 || got[0] != "linux/amd64" || got[1] != "linux/arm64" {
		t.Fatalf("unexpected platforms: %#v", got)
	}
}

func TestValidateAppManifest(t *testing.T) {
	manifest := &appManifest{
		Kind:    "VersionSet",
		Product: "test",
		Version: "1.0.0",
		Source: appManifestSource{
			HelmRepoURL: "https://example.com/charts",
		},
		Releases: map[string]appManifestRelease{
			"web": {Chart: "web", Version: "1.0.0"},
		},
	}

	if err := validateAppManifest(manifest); err != nil {
		t.Fatalf("validateAppManifest returned error: %v", err)
	}
}

func TestExtractAppImagesFromValues(t *testing.T) {
	values := map[string]any{
		"image": map[string]any{
			"registry": "docker.io",
			"controller": map[string]any{
				"repository": "bitnami/nginx",
				"tag":        "1.0.0",
			},
			"sidecar": map[string]any{
				"repository": "bitnami/os-shell",
				"tag":        "1.0.0",
			},
		},
	}

	refs, ok := extractAppImagesFromValues(values)
	if !ok {
		t.Fatal("expected extractAppImagesFromValues to find images")
	}
	if len(refs) != 2 {
		t.Fatalf("got %d refs, want 2", len(refs))
	}
	if refs[0].Repository == "" || refs[0].Tag == "" {
		t.Fatalf("unexpected first ref: %+v", refs[0])
	}
}

func TestBuildExportedManifest(t *testing.T) {
	originalTimeNow := timeNow
	timeNow = func() time.Time {
		return time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	}
	defer func() { timeNow = originalTimeNow }()

	out, err := buildExportedManifest([]byte("kind: VersionSet\nproduct: demo\nversion: 1.0.0\nsource:\n  helmRepoUrl: https://example.com\nreleases:\n  web:\n    chart: web\n    version: 1.0.0\n"), []string{"linux/amd64", "linux/arm64"}, []appPackageImage{{
		Source:             "docker.io/library/nginx:1.27.0",
		Repository:         "library/nginx",
		Tag:                "1.27.0",
		LocalRef:           "library/nginx:1.27.0",
		RequestedPlatforms: []string{"linux/amd64", "linux/arm64"},
		ExportedPlatforms:  []string{"linux/amd64"},
		Exported:           true,
	}}, []string{"image docker.io/library/busybox:1.0 skipped: EOF"})
	if err != nil {
		t.Fatalf("buildExportedManifest returned error: %v", err)
	}

	var doc map[string]any
	if err := yaml.Unmarshal(out, &doc); err != nil {
		t.Fatalf("yaml.Unmarshal returned error: %v", err)
	}

	offlinePackage, ok := doc["offlinePackage"].(map[string]any)
	if !ok {
		t.Fatalf("missing offlinePackage block: %+v", doc)
	}
	if offlinePackage["platform"] != nil {
		t.Fatalf("unexpected single platform field: %+v", offlinePackage["platform"])
	}
	platforms, ok := offlinePackage["platforms"].([]any)
	if !ok || len(platforms) != 2 {
		t.Fatalf("unexpected platforms block: %+v", offlinePackage["platforms"])
	}
	images, ok := offlinePackage["images"].([]any)
	if !ok || len(images) != 1 {
		t.Fatalf("unexpected images block: %+v", offlinePackage["images"])
	}
	imageErrors, ok := offlinePackage["imageErrors"].([]any)
	if !ok || len(imageErrors) != 1 {
		t.Fatalf("unexpected imageErrors block: %+v", offlinePackage["imageErrors"])
	}
}

func TestValidateAppPackageLayout(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{
		filepath.Join(root, "manifest.yaml"),
		filepath.Join(root, "charts"),
		filepath.Join(root, "images"),
	} {
		if strings.HasSuffix(path, ".yaml") {
			if err := os.WriteFile(path, []byte("kind: VersionSet"), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	if _, _, _, err := validateAppPackageLayout(root); err != nil {
		t.Fatalf("validateAppPackageLayout returned error: %v", err)
	}
}

func TestCountExportedAppImages(t *testing.T) {
	got := countExportedAppImages([]appPackageImage{
		{Exported: true},
		{Exported: false},
		{Exported: true},
	})
	if got != 2 {
		t.Fatalf("got %d, want 2", got)
	}
}
