package addons

import (
	"context"
	"strings"
	"testing"
	"unsafe"

	helmtesting "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3/testing"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/chart"
)

func TestReconcileReturnsErrorWhenAddonChartMissing(t *testing.T) {
	pkg := &servicepackage.ServicePackage{}

	pkgCharts := servicepackage.Charts{
		{
			Path: "charts/ingress-nginx-4.15.0.tgz",
			Metadata: chart.Metadata{
				Name:    "ingress-nginx",
				Version: "4.15.0",
			},
		},
	}

	*pkg = servicepackage.ServicePackage{}
	setServicePackageCharts(pkg, pkgCharts)

	h := helmtesting.New("resource", t.Logf)

	err := Reconcile(
		context.Background(),
		logrus.New(),
		h,
		pkg,
		"node1:5000",
		nil,
		configuration.CSAddonNameNodeExporter,
	)
	if err == nil {
		t.Fatal("expected error when addon chart is missing")
	}
	if !strings.Contains(err.Error(), "chart proton-node-exporter for addon node-exporter not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReconcileInstallsAddonWhenReleaseDoesNotExist(t *testing.T) {
	pkg := &servicepackage.ServicePackage{}
	setServicePackageCharts(pkg, servicepackage.Charts{
		{
			Path: "charts/proton-node-exporter-0.0.0.tgz",
			Metadata: chart.Metadata{
				Name:    ChartNameNodeExporter,
				Version: "0.0.0",
			},
		},
	})

	h := helmtesting.New("resource", t.Logf)

	err := Reconcile(
		context.Background(),
		logrus.New(),
		h,
		pkg,
		"node1:5000",
		nil,
		configuration.CSAddonNameNodeExporter,
	)
	if err != nil {
		t.Fatalf("expected install to succeed when release is absent, got: %v", err)
	}
	if !h.CallUpgrade {
		t.Fatal("expected addon reconcile to install via upgrade --install when release is absent")
	}
}

func TestHelmValuesForIngressNginxDefaultsPorts(t *testing.T) {
	values, ok := helmValuesForAddon(configuration.CSAddonNameIngressNginx, "node1:5000", nil).(ValuesIngressNginx)
	if !ok {
		t.Fatalf("expected ingress-nginx values type, got %T", values)
	}

	if values.Controller.ContainerPort.HTTP == nil || *values.Controller.ContainerPort.HTTP != 80 {
		t.Fatalf("expected default http port 80, got %#v", values.Controller.ContainerPort.HTTP)
	}
	if values.Controller.ContainerPort.HTTPS == nil || *values.Controller.ContainerPort.HTTPS != 443 {
		t.Fatalf("expected default https port 443, got %#v", values.Controller.ContainerPort.HTTPS)
	}
}

func TestHelmValuesForIngressNginxUsesConfiguredPorts(t *testing.T) {
	values, ok := helmValuesForAddon(
		configuration.CSAddonNameIngressNginx,
		"node1:5000",
		&configuration.CSAddonsConfig{
			IngressNginx: &configuration.CSAddonIngressNginxConfig{
				HTTPPort:  8080,
				HTTPSPort: 8443,
			},
		},
	).(ValuesIngressNginx)
	if !ok {
		t.Fatalf("expected ingress-nginx values type, got %T", values)
	}

	if values.Controller.ContainerPort.HTTP == nil || *values.Controller.ContainerPort.HTTP != 8080 {
		t.Fatalf("expected configured http port 8080, got %#v", values.Controller.ContainerPort.HTTP)
	}
	if values.Controller.ContainerPort.HTTPS == nil || *values.Controller.ContainerPort.HTTPS != 8443 {
		t.Fatalf("expected configured https port 8443, got %#v", values.Controller.ContainerPort.HTTPS)
	}
}

func setServicePackageCharts(pkg *servicepackage.ServicePackage, charts servicepackage.Charts) {
	type servicePackageAlias struct {
		charts   []servicepackage.Chart
		images   []string
		basePath string
	}

	alias := (*servicePackageAlias)(unsafe.Pointer(pkg))
	alias.charts = charts
	alias.basePath = "/tmp/service-package"
}

var _ helm3.Client = (*helmtesting.FakeHelm3)(nil)
