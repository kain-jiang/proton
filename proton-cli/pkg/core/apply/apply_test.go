package apply

import (
	"os"
	"path/filepath"
	"testing"

	registryfake "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/registry/fake"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

func TestAppendOptionalModulesDoesNotIncludeECeph(t *testing.T) {
	clusterConf := &configuration.ClusterConfig{
		Deploy: &configuration.Deploy{},
		Cr: &configuration.Cr{
			Local: &configuration.LocalCR{},
		},
		ComponentManage: &configuration.ComponentManagement{},
		ECeph: &configuration.ECeph{
			Hosts: []string{"node1"},
		},
	}

	pkg := loadMinimalServicePackageForTest(t)
	modules := appendOptionalModules(
		nil,
		nil,
		nil,
		nil,
		nil,
		clusterConf,
		nil,
		registryfake.New("registry.example.com", nil),
		pkg,
		nil,
	)

	if len(modules) != 1 {
		t.Fatalf("expected only component_manage module, got %d", len(modules))
	}
	if modules[0].name != "component_manage" {
		t.Fatalf("expected component_manage module, got %q", modules[0].name)
	}
	for _, module := range modules {
		if module.name == "eceph" {
			t.Fatalf("unexpected eceph module in apply optional modules")
		}
	}
}

func loadMinimalServicePackageForTest(t *testing.T) *servicepackage.ServicePackage {
	t.Helper()

	baseDir := t.TempDir()
	chartsDir := filepath.Join(baseDir, "charts", "component-manage")
	imagesDir := filepath.Join(baseDir, "images")

	if err := os.MkdirAll(chartsDir, 0o755); err != nil {
		t.Fatalf("create charts dir: %v", err)
	}
	if err := os.MkdirAll(imagesDir, 0o755); err != nil {
		t.Fatalf("create images dir: %v", err)
	}

	chartYAML := "apiVersion: v2\nname: component-manage\nversion: 1.0.0\n"
	if err := os.WriteFile(filepath.Join(chartsDir, "Chart.yaml"), []byte(chartYAML), 0o644); err != nil {
		t.Fatalf("write chart yaml: %v", err)
	}
	indexJSON := "{\"schemaVersion\":2,\"manifests\":[]}"
	if err := os.WriteFile(filepath.Join(imagesDir, "index.json"), []byte(indexJSON), 0o644); err != nil {
		t.Fatalf("write index.json: %v", err)
	}

	pkg := &servicepackage.ServicePackage{}
	if err := pkg.Load(baseDir); err != nil {
		t.Fatalf("load service package: %v", err)
	}
	return pkg
}
