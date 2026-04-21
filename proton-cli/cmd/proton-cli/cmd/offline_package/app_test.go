package offline_package

import (
	"archive/tar"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
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

func TestResolveAppManifestLocationLocalRelative(t *testing.T) {
	base := filepath.Join("/tmp", "manifests", "root.yaml")
	got, err := resolveAppManifestLocation(base, "./deps/child.yaml")
	if err != nil {
		t.Fatalf("resolveAppManifestLocation returned error: %v", err)
	}

	want := filepath.Join("/tmp", "manifests", "deps", "child.yaml")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolveAppManifestLocationURLRelative(t *testing.T) {
	got, err := resolveAppManifestLocation("https://example.com/manifests/root.yaml", "./deps/child.yaml")
	if err != nil {
		t.Fatalf("resolveAppManifestLocation returned error: %v", err)
	}
	if got != "https://example.com/manifests/deps/child.yaml" {
		t.Fatalf("got %q", got)
	}
}

func TestLoadAppManifestTreeLocalDependencies(t *testing.T) {
	root := t.TempDir()
	depsDir := filepath.Join(root, "deps")
	if err := os.MkdirAll(depsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeManifest := func(path, body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeManifest(filepath.Join(root, "root.yaml"), `
kind: VersionSet
product: root
version: 1.0.0
source:
  helmRepoUrl: https://charts.example.com/root
dependencies:
  - product: dep
    version: 1.0.0
    manifest: ./deps/dep.yaml
releases:
  root:
    chart: root-chart
    version: 1.0.0
`)
	writeManifest(filepath.Join(depsDir, "dep.yaml"), `
kind: VersionSet
product: dep
version: 1.0.0
source:
  helmRepoUrl: https://charts.example.com/dep
dependencies:
  - product: nested
    version: 1.0.0
    manifest: ./nested.yaml
releases:
  dep:
    chart: dep-chart
    version: 2.0.0
`)
	writeManifest(filepath.Join(depsDir, "nested.yaml"), `
kind: VersionSet
product: nested
version: 1.0.0
source:
  helmRepoUrl: https://charts.example.com/nested
releases:
  nested:
    chart: nested-chart
    version: 3.0.0
`)

	_, _, documents, err := loadAppManifestTree(context.Background(), filepath.Join(root, "root.yaml"), true)
	if err != nil {
		t.Fatalf("loadAppManifestTree returned error: %v", err)
	}
	if len(documents) != 3 {
		t.Fatalf("got %d documents, want 3", len(documents))
	}

	charts := collectAppChartRequests(documents)
	if len(charts) != 3 {
		t.Fatalf("got %d charts, want 3", len(charts))
	}
}

func TestLoadAppManifestTreeURLRelativeDependencies(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/manifests/root.yaml", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`
kind: VersionSet
product: root
version: 1.0.0
source:
  helmRepoUrl: https://charts.example.com/root
dependencies:
  - product: dep
    version: 1.0.0
    manifest: ./deps/dep.yaml
releases:
  root:
    chart: root-chart
    version: 1.0.0
`))
	})
	mux.HandleFunc("/manifests/deps/dep.yaml", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`
kind: VersionSet
product: dep
version: 1.0.0
source:
  helmRepoUrl: https://charts.example.com/dep
releases:
  dep:
    chart: dep-chart
    version: 2.0.0
`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	_, _, documents, err := loadAppManifestTree(context.Background(), server.URL+"/manifests/root.yaml", true)
	if err != nil {
		t.Fatalf("loadAppManifestTree returned error: %v", err)
	}
	if len(documents) != 2 {
		t.Fatalf("got %d documents, want 2", len(documents))
	}
	if documents[1].Location != server.URL+"/manifests/deps/dep.yaml" {
		t.Fatalf("unexpected dependency location: %q", documents[1].Location)
	}
}

func TestLoadAppManifestTreeDisableDependencies(t *testing.T) {
	root := t.TempDir()
	depsDir := filepath.Join(root, "deps")
	if err := os.MkdirAll(depsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeManifest := func(path, body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	rootPath := filepath.Join(root, "root.yaml")
	writeManifest(rootPath, `
kind: VersionSet
product: root
version: 1.0.0
source:
  helmRepoUrl: https://charts.example.com/root
dependencies:
  - product: dep
    version: 1.0.0
    manifest: ./deps/dep.yaml
releases:
  root:
    chart: root-chart
    version: 1.0.0
`)
	writeManifest(filepath.Join(depsDir, "dep.yaml"), `
kind: VersionSet
product: dep
version: 1.0.0
source:
  helmRepoUrl: https://charts.example.com/dep
releases:
  dep:
    chart: dep-chart
    version: 2.0.0
`)

	_, _, documents, err := loadAppManifestTree(context.Background(), rootPath, false)
	if err != nil {
		t.Fatalf("loadAppManifestTree returned error: %v", err)
	}
	if len(documents) != 1 {
		t.Fatalf("got %d documents, want 1", len(documents))
	}

	charts := collectAppChartRequests(documents)
	if len(charts) != 1 {
		t.Fatalf("got %d charts, want 1", len(charts))
	}
	if charts[0].Name != "root-chart" {
		t.Fatalf("unexpected chart: %+v", charts[0])
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

func TestExtractAppImagesFromValuesRegistryWithPathPrefix(t *testing.T) {
	values := map[string]any{
		"image": map[string]any{
			"registry":   "swr.cn-east-3.myhuaweicloud.com/kweaver-ai",
			"repository": "dip/agent-backend",
			"tag":        "0.5.1",
		},
	}

	refs, ok := extractAppImagesFromValues(values)
	if !ok || len(refs) != 1 {
		t.Fatalf("unexpected refs: %#v", refs)
	}
	if refs[0].Source != "swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip/agent-backend:0.5.1" {
		t.Fatalf("unexpected source: %q", refs[0].Source)
	}
	if refs[0].Repository != "dip/agent-backend" {
		t.Fatalf("unexpected repository: %q", refs[0].Repository)
	}
	if refs[0].LocalRef() != "dip/agent-backend:0.5.1" {
		t.Fatalf("unexpected local ref: %q", refs[0].LocalRef())
	}
}

func TestExtractAppImagesFromValuesWithoutRegistry(t *testing.T) {
	values := map[string]any{
		"image": map[string]any{
			"repository": "as/audit-log",
			"tag":        "0.5.0",
		},
	}

	refs, ok := extractAppImagesFromValues(values)
	if !ok || len(refs) != 1 {
		t.Fatalf("unexpected refs: %#v", refs)
	}
	if refs[0].Registry != "" {
		t.Fatalf("unexpected registry: %q", refs[0].Registry)
	}
	if refs[0].Source != "as/audit-log:0.5.0" {
		t.Fatalf("unexpected source: %q", refs[0].Source)
	}
	if refs[0].LocalRef() != "as/audit-log:0.5.0" {
		t.Fatalf("unexpected local ref: %q", refs[0].LocalRef())
	}
}

func TestBuildExportedManifest(t *testing.T) {
	originalTimeNow := timeNow
	timeNow = func() time.Time {
		return time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	}
	defer func() { timeNow = originalTimeNow }()

	out, err := buildExportedManifest([]byte("kind: VersionSet\nproduct: demo\nversion: 1.0.0\nsource:\n  helmRepoUrl: https://example.com\nreleases:\n  web:\n    chart: web\n    version: 1.0.0\n"), []string{"linux/amd64", "linux/arm64"}, "mirror.example.com", []appPackageImage{{
		Source:             "docker.io/library/nginx:1.27.0",
		PullSource:         "mirror.example.com/library/nginx:1.27.0",
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
	if offlinePackage["overrideRegistry"] != "mirror.example.com" {
		t.Fatalf("unexpected overrideRegistry: %+v", offlinePackage["overrideRegistry"])
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

func TestOverrideAppImageSource(t *testing.T) {
	image := appImageRef{
		Registry:   "docker.io",
		Source:     "docker.io/library/nginx:1.27.0",
		Repository: "library/nginx",
		Tag:        "1.27.0",
	}

	got, err := overrideAppImageSource(image, "mirror.example.com")
	if err != nil {
		t.Fatalf("overrideAppImageSource returned error: %v", err)
	}
	if got != "mirror.example.com/library/nginx:1.27.0" {
		t.Fatalf("got %q", got)
	}
}

func TestOverrideAppImageSourceWithRepositoryPrefix(t *testing.T) {
	tests := []struct {
		name     string
		image    appImageRef
		override string
		want     string
	}{
		{
			name: "prepend prefix",
			image: appImageRef{
				Registry:   "acr.aishu.cn",
				Source:     "acr.aishu.cn/dip/agent-backend:0.5.1",
				Repository: "dip/agent-backend",
				Tag:        "0.5.1",
			},
			override: "swr.cn-east-3.myhuaweicloud.com/kweaver-ai",
			want:     "swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip/agent-backend:0.5.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := overrideAppImageSource(tt.image, tt.override)
			if err != nil {
				t.Fatalf("overrideAppImageSource returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOverrideAppImageSourceFillsMissingRegistry(t *testing.T) {
	image := appImageRef{
		Source:     "as/audit-log:0.5.0",
		Repository: "as/audit-log",
		Tag:        "0.5.0",
	}

	got, err := overrideAppImageSource(image, "acr.example.com")
	if err != nil {
		t.Fatalf("overrideAppImageSource returned error: %v", err)
	}
	if got != "acr.example.com/as/audit-log:0.5.0" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveAppExportImagesDirDefault(t *testing.T) {
	workdir := filepath.Join("/tmp", "offline-app-export")

	got, err := resolveAppExportImagesDir(workdir, "")
	if err != nil {
		t.Fatalf("resolveAppExportImagesDir returned error: %v", err)
	}

	want := filepath.Join(workdir, "images")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolveAppExportImagesDirWithCacheDir(t *testing.T) {
	workdir := filepath.Join("/tmp", "offline-app-export")
	cacheDir := filepath.Join("/var", "cache", "proton-cli")

	got, err := resolveAppExportImagesDir(workdir, cacheDir)
	if err != nil {
		t.Fatalf("resolveAppExportImagesDir returned error: %v", err)
	}

	want := filepath.Join(cacheDir, "images")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNewAppExportCommandIncludesCacheDirFlag(t *testing.T) {
	cmd := newAppExportCommand()
	flag := cmd.Flags().Lookup("cache-dir")
	if flag == nil {
		t.Fatal("expected cache-dir flag")
	}
}

func TestTarAppPackageIncludesExternalImagesDir(t *testing.T) {
	workdir := t.TempDir()
	imagesDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(workdir, "manifest.yaml"), []byte("kind: VersionSet\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(workdir, "charts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workdir, "charts", "demo.tgz"), []byte("chart"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(imagesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imagesDir, "index.json"), []byte(`{"manifests":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imagesDir, "oci-layout"), []byte(`{"imageLayoutVersion":"1.0.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(t.TempDir(), "offline-app-package.tar")
	if err := tarAppPackage(workdir, imagesDir, output); err != nil {
		t.Fatalf("tarAppPackage returned error: %v", err)
	}

	f, err := os.Open(output)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tr := tar.NewReader(f)
	var entries []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read tar entry: %v", err)
		}
		entries = append(entries, hdr.Name)
	}

	for _, required := range []string{"manifest.yaml", "charts/demo.tgz", "images/index.json", "images/oci-layout"} {
		found := slices.Contains(entries, required)
		if !found {
			t.Fatalf("missing tar entry %q in %#v", required, entries)
		}
	}
}

func TestHydrateAppImportOptionsRequiresTargetsWithoutAuto(t *testing.T) {
	opts := &appImportOptions{}
	err := hydrateAppImportOptions(context.Background(), opts)
	if err == nil || !strings.Contains(err.Error(), "target registry is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHydrateAppImportOptionsUsesAutoTargets(t *testing.T) {
	opts := &appImportOptions{auto: true}

	original := loadAppImportAutoTargetsFunc
	loadAppImportAutoTargetsFunc = func(_ context.Context) (*appImportAutoTargets, error) {
		return &appImportAutoTargets{
			registries:          []string{"registry.example.com"},
			registryUsername:    "user",
			registryPassword:    "pass",
			registryPlainHTTP:   true,
			chartmuseumURLs:     []string{"http://chartmuseum.example.com"},
			chartmuseumUsername: "chart-user",
			chartmuseumPassword: "chart-pass",
		}, nil
	}
	defer func() {
		loadAppImportAutoTargetsFunc = original
	}()

	if err := hydrateAppImportOptions(context.Background(), opts); err != nil {
		t.Fatalf("hydrateAppImportOptions returned error: %v", err)
	}
	if opts.registry != "registry.example.com" || opts.chartmuseumURL != "http://chartmuseum.example.com" {
		t.Fatalf("unexpected hydrated options: %+v", opts)
	}
	if !opts.registryPlainHTTP {
		t.Fatalf("expected registryPlainHTTP to be true")
	}
}

func TestHydrateAppImportOptionsManualOverridesAuto(t *testing.T) {
	opts := &appImportOptions{
		auto:              true,
		registry:          "manual.registry",
		chartmuseumURL:    "http://manual.chartmuseum",
		registryPlainHTTP: false,
	}

	original := loadAppImportAutoTargetsFunc
	loadAppImportAutoTargetsFunc = func(_ context.Context) (*appImportAutoTargets, error) {
		return &appImportAutoTargets{
			registries:        []string{"auto.registry"},
			registryPlainHTTP: true,
			chartmuseumURLs:   []string{"http://auto.chartmuseum"},
		}, nil
	}
	defer func() {
		loadAppImportAutoTargetsFunc = original
	}()

	if err := hydrateAppImportOptions(context.Background(), opts); err != nil {
		t.Fatalf("hydrateAppImportOptions returned error: %v", err)
	}
	if opts.registry != "manual.registry" {
		t.Fatalf("expected manual registry to win, got %q", opts.registry)
	}
	if opts.chartmuseumURL != "http://manual.chartmuseum" {
		t.Fatalf("expected manual chartmuseum to win, got %q", opts.chartmuseumURL)
	}
	if !opts.registryPlainHTTP {
		t.Fatalf("expected auto plainHTTP to fill missing true value")
	}
}

func TestAppImportAutoTargetsFromExternalOCI(t *testing.T) {
	cfg := &configuration.ClusterConfig{
		Cr: &configuration.Cr{
			External: &configuration.ExternalCR{
				ImageRepo: configuration.RepoOCI,
				ChartRepo: configuration.RepoChartmuseum,
				OCI: &configuration.OCI{
					Registry:  "registry.example.com",
					PlainHTTP: true,
					Username:  "oci-user",
					Password:  "oci-pass",
				},
				Chartmuseum: &configuration.Chartmuseum{
					Host:     "http://chartmuseum.example.com",
					Username: "chart-user",
					Password: "chart-pass",
				},
			},
		},
	}

	targets, err := appImportAutoTargetsFromConfig(cfg)
	if err != nil {
		t.Fatalf("appImportAutoTargetsFromConfig returned error: %v", err)
	}
	if len(targets.registries) != 1 || targets.registries[0] != "registry.example.com" || !targets.registryPlainHTTP {
		t.Fatalf("unexpected registry targets: %+v", targets)
	}
	if len(targets.chartmuseumURLs) != 1 || targets.chartmuseumURLs[0] != "http://chartmuseum.example.com" {
		t.Fatalf("unexpected chartmuseum url: %+v", targets)
	}
}

func TestAppImportAutoTargetsRejectsNonChartmuseum(t *testing.T) {
	cfg := &configuration.ClusterConfig{
		Cr: &configuration.Cr{
			External: &configuration.ExternalCR{
				ImageRepo: configuration.RepoOCI,
				ChartRepo: configuration.RepoOCI,
				OCI: &configuration.OCI{
					Registry: "registry.example.com",
				},
			},
		},
	}

	_, err := appImportAutoTargetsFromConfig(cfg)
	if err == nil || !strings.Contains(err.Error(), "not chartmuseum") {
		t.Fatalf("unexpected error: %v", err)
	}
}
