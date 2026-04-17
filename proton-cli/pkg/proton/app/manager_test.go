package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	helm3testing "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3/testing"
)

func TestBuildInstallPlanPreservesStage(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
releases:
  pre-release:
    chart: pre-chart
    version: 1.0.0
    stage: pre
  normal-release:
    chart: normal-chart
    version: 1.0.0
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	plan, err := BuildInstallPlan(manifestPath)
	if err != nil {
		t.Fatalf("BuildInstallPlan() error = %v", err)
	}

	if len(plan.Steps) != 2 {
		t.Fatalf("len(plan.Steps) = %d, want 2", len(plan.Steps))
	}
	if len(plan.Steps[0]) != 1 || plan.Steps[0][0].Stage != "pre" {
		t.Fatalf("first step = %#v, want one pre-stage item", plan.Steps[0])
	}
	if len(plan.Steps[1]) != 1 || plan.Steps[1][0].Stage != "" {
		t.Fatalf("second step = %#v, want one normal-stage item", plan.Steps[1])
	}
}

func TestBuildInstallPlanHonorsDependsOn(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
releases:
  studio-web:
    chart: studio-web
    version: 1.0.0
  agent-web:
    chart: agent-web
    version: 1.0.0
    dependsOn:
      - studio-web
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	plan, err := BuildInstallPlan(manifestPath)
	if err != nil {
		t.Fatalf("BuildInstallPlan() error = %v", err)
	}

	if len(plan.Steps) != 2 {
		t.Fatalf("len(plan.Steps) = %d, want 2", len(plan.Steps))
	}
	if len(plan.Steps[0]) != 1 || plan.Steps[0][0].ReleaseName != "studio-web" {
		t.Fatalf("first step = %#v, want studio-web", plan.Steps[0])
	}
	if len(plan.Steps[1]) != 1 || plan.Steps[1][0].ReleaseName != "agent-web" {
		t.Fatalf("second step = %#v, want agent-web", plan.Steps[1])
	}
}

func TestBuildInstallPlanRejectsUnknownDependsOn(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
releases:
  agent-web:
    chart: agent-web
    version: 1.0.0
    dependsOn:
      - studio-web
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	_, err := BuildInstallPlan(manifestPath)
	if err == nil {
		t.Fatal("BuildInstallPlan() error = nil, want unknown dependsOn error")
	}
	if !strings.Contains(err.Error(), "unknown dependency release") {
		t.Fatalf("BuildInstallPlan() error = %v, want unknown dependency release", err)
	}
}

func TestBuildInstallPlanRejectsSelfDependsOn(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
releases:
  agent-web:
    chart: agent-web
    version: 1.0.0
    dependsOn:
      - agent-web
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	_, err := BuildInstallPlan(manifestPath)
	if err == nil {
		t.Fatal("BuildInstallPlan() error = nil, want self dependency error")
	}
	if !strings.Contains(err.Error(), "cannot depend on itself") {
		t.Fatalf("BuildInstallPlan() error = %v, want cannot depend on itself", err)
	}
}

func TestBuildInstallPlanRejectsDependencyCycle(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
releases:
  studio-web:
    chart: studio-web
    version: 1.0.0
    dependsOn:
      - agent-web
  agent-web:
    chart: agent-web
    version: 1.0.0
    dependsOn:
      - studio-web
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	_, err := BuildInstallPlan(manifestPath)
	if err == nil {
		t.Fatal("BuildInstallPlan() error = nil, want cycle error")
	}
	if !strings.Contains(err.Error(), "dependency cycle") {
		t.Fatalf("BuildInstallPlan() error = %v, want dependency cycle", err)
	}
}

func TestBuildInstallPlanCombinesStagePreWithDependsOn(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
releases:
  db-init:
    chart: db-init
    version: 1.0.0
    stage: pre
  studio-web:
    chart: studio-web
    version: 1.0.0
    stage: pre
  agent-web:
    chart: agent-web
    version: 1.0.0
    dependsOn:
      - studio-web
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	plan, err := BuildInstallPlan(manifestPath)
	if err != nil {
		t.Fatalf("BuildInstallPlan() error = %v", err)
	}

	if len(plan.Steps) != 2 {
		t.Fatalf("len(plan.Steps) = %d, want 2", len(plan.Steps))
	}
	if got := []string{plan.Steps[0][0].ReleaseName, plan.Steps[0][1].ReleaseName}; strings.Join(got, ",") != "db-init,studio-web" {
		t.Fatalf("first step releases = %v, want [db-init studio-web]", got)
	}
	if len(plan.Steps[1]) != 1 || plan.Steps[1][0].ReleaseName != "agent-web" {
		t.Fatalf("second step = %#v, want agent-web", plan.Steps[1])
	}
}

func TestBuildInstallPlanHonorsOptionalDependencyEnabledIf(t *testing.T) {
	dir := t.TempDir()
	dependencyPath := filepath.Join(dir, "isf.yaml")
	dependencyManifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: isf
version: 1.0.0
source:
  helmRepoName: local
releases:
  isf-server:
    chart: isf-server
    version: 1.0.0
`
	if err := os.WriteFile(dependencyPath, []byte(dependencyManifest), 0o600); err != nil {
		t.Fatalf("write dependency manifest: %v", err)
	}

	rootPath := filepath.Join(dir, "manifest.yaml")
	rootManifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
dependencies:
  - product: isf
    version: 1.0.0
    manifest: ./isf.yaml
    defaultEnabled: true
    enabledIf: auth.enabled
releases:
  demo-release:
    chart: demo-release
    version: 1.0.0
`
	if err := os.WriteFile(rootPath, []byte(rootManifest), 0o600); err != nil {
		t.Fatalf("write root manifest: %v", err)
	}

	plan, err := BuildInstallPlanWithValues(rootPath, map[string]interface{}{
		"auth": map[string]interface{}{
			"enabled": false,
		},
	})
	if err != nil {
		t.Fatalf("BuildInstallPlanWithValues() error = %v", err)
	}

	if len(plan.Steps) != 1 {
		t.Fatalf("len(plan.Steps) = %d, want 1", len(plan.Steps))
	}
	if len(plan.Steps[0]) != 1 || plan.Steps[0][0].ReleaseName != "demo-release" {
		t.Fatalf("plan.Steps = %#v, want only demo-release", plan.Steps)
	}

	plan, err = BuildInstallPlanWithValues(rootPath, map[string]interface{}{
		"auth": map[string]interface{}{
			"enabled": true,
		},
	})
	if err != nil {
		t.Fatalf("BuildInstallPlanWithValues() error = %v", err)
	}

	if len(plan.Steps) != 2 {
		t.Fatalf("len(plan.Steps) = %d, want 2 when auth.enabled=true", len(plan.Steps))
	}
	if len(plan.Steps[0]) != 1 || plan.Steps[0][0].ReleaseName != "isf-server" {
		t.Fatalf("first step = %#v, want isf-server", plan.Steps[0])
	}
}

func TestBuildInstallPlanRejectsNonBoolEnabledIfValue(t *testing.T) {
	dir := t.TempDir()
	dependencyPath := filepath.Join(dir, "isf.yaml")
	dependencyManifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: isf
version: 1.0.0
source:
  helmRepoName: local
releases:
  isf-server:
    chart: isf-server
    version: 1.0.0
`
	if err := os.WriteFile(dependencyPath, []byte(dependencyManifest), 0o600); err != nil {
		t.Fatalf("write dependency manifest: %v", err)
	}

	rootPath := filepath.Join(dir, "manifest.yaml")
	rootManifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
dependencies:
  - product: isf
    version: 1.0.0
    manifest: ./isf.yaml
    defaultEnabled: true
    enabledIf: auth.enabled
releases:
  demo-release:
    chart: demo-release
    version: 1.0.0
`
	if err := os.WriteFile(rootPath, []byte(rootManifest), 0o600); err != nil {
		t.Fatalf("write root manifest: %v", err)
	}

	_, err := BuildInstallPlanWithValues(rootPath, map[string]interface{}{
		"auth": map[string]interface{}{
			"enabled": "disabled",
		},
	})
	if err == nil {
		t.Fatal("BuildInstallPlanWithValues() error = nil, want non-bool enabledIf error")
	}
	if !strings.Contains(err.Error(), "must be a boolean") {
		t.Fatalf("BuildInstallPlanWithValues() error = %v, want boolean type error", err)
	}
}

func TestInstallWithValuesWaitsOnlyForStagePre(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
releases:
  pre-release:
    chart: pre-chart
    version: 1.0.0
    stage: pre
  normal-release:
    chart: normal-chart
    version: 1.0.0
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	logger := logrus.New()
	manager := &Manager{
		Helm:      helm3testing.New("kweaver", logger.Printf),
		Namespace: "kweaver",
		Logger:    logger,
	}

	err := manager.InstallWithValues(context.Background(), manifestPath, map[string]interface{}{}, InstallOptions{
		Namespace: "kweaver",
		Timeout:   "2m",
	})
	if err != nil {
		t.Fatalf("InstallWithValues() error = %v", err)
	}

	fakeHelm := manager.Helm.(*helm3testing.FakeHelm3)
	if len(fakeHelm.UpgradeCalls) != 2 {
		t.Fatalf("len(UpgradeCalls) = %d, want 2", len(fakeHelm.UpgradeCalls))
	}

	var sawPreWait bool
	var sawNormalNoWait bool
	for _, call := range fakeHelm.UpgradeCalls {
		switch call.Release {
		case "pre-release":
			if !call.Wait || call.Timeout != 2*time.Minute {
				t.Fatalf("pre-release call = %#v, want wait=true timeout=2m", call)
			}
			sawPreWait = true
		case "normal-release":
			if call.Wait {
				t.Fatalf("normal-release call = %#v, want wait=false", call)
			}
			sawNormalNoWait = true
		}
	}

	if !sawPreWait || !sawNormalNoWait {
		t.Fatalf("UpgradeCalls = %#v, want both pre and normal releases", fakeHelm.UpgradeCalls)
	}
}

func TestInstallWithValuesWaitsForDependencyProviders(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
releases:
  dip-data-migrator:
    chart: dip-data-migrator
    version: 1.0.0
  studio-web:
    chart: studio-web
    version: 1.0.0
    dependsOn:
      - dip-data-migrator
  agent-web:
    chart: agent-web
    version: 1.0.0
    dependsOn:
      - studio-web
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	logger := logrus.New()
	manager := &Manager{
		Helm:      helm3testing.New("kweaver", logger.Printf),
		Namespace: "kweaver",
		Logger:    logger,
	}

	err := manager.InstallWithValues(context.Background(), manifestPath, map[string]interface{}{}, InstallOptions{
		Namespace: "kweaver",
		Timeout:   "2m",
	})
	if err != nil {
		t.Fatalf("InstallWithValues() error = %v", err)
	}

	fakeHelm := manager.Helm.(*helm3testing.FakeHelm3)
	if len(fakeHelm.UpgradeCalls) != 3 {
		t.Fatalf("len(UpgradeCalls) = %d, want 3", len(fakeHelm.UpgradeCalls))
	}

	for _, call := range fakeHelm.UpgradeCalls {
		switch call.Release {
		case "dip-data-migrator":
			if !call.Wait {
				t.Fatalf("dip-data-migrator call = %#v, want wait=true because studio-web depends on it", call)
			}
		case "studio-web":
			if !call.Wait {
				t.Fatalf("studio-web call = %#v, want wait=true because agent-web depends on it", call)
			}
		case "agent-web":
			if call.Wait {
				t.Fatalf("agent-web call = %#v, want wait=false because no release depends on it", call)
			}
		}
	}
}

func TestInstallWithValuesLogsStageAwareWaitBehavior(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
releases:
  pre-release:
    chart: pre-chart
    version: 1.0.0
    stage: pre
  normal-release:
    chart: normal-chart
    version: 1.0.0
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	var logs bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&logs)
	logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true, DisableColors: true})

	manager := &Manager{
		Helm:      helm3testing.New("kweaver", logger.Printf),
		Namespace: "kweaver",
		Logger:    logger,
	}

	err := manager.InstallWithValues(context.Background(), manifestPath, map[string]interface{}{}, InstallOptions{
		Namespace: "kweaver",
		Timeout:   "2m",
	})
	if err != nil {
		t.Fatalf("InstallWithValues() error = %v", err)
	}

	output := logs.String()
	for _, want := range []string{
		"release=pre-release",
		"stage=pre",
		"wait=true",
		"timeout=2m0s",
		"waiting for release readiness",
		"Release readiness confirmed",
		"release=normal-release",
		"wait=false",
		"submitted without readiness wait",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("logs missing %q\nfull logs:\n%s", want, output)
		}
	}
}

func TestInstallWithValuesPassesReleaseValuesWithoutDipStudioValidation(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
releases:
  dip-studio:
    chart: dip-studio
    version: 0.4.0
    values:
      replicaCount: 2
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	logger := logrus.New()
	manager := &Manager{
		Helm:      helm3testing.New("kweaver", logger.Printf),
		Namespace: "kweaver",
		Logger:    logger,
	}

	err := manager.InstallWithValues(context.Background(), manifestPath, map[string]interface{}{
		"image": map[string]interface{}{"registry": "example.com/demo"},
	}, InstallOptions{
		Namespace: "kweaver",
		Timeout:   "2m",
	})
	if err != nil {
		t.Fatalf("InstallWithValues() error = %v", err)
	}

	fakeHelm := manager.Helm.(*helm3testing.FakeHelm3)
	if len(fakeHelm.UpgradeCalls) != 1 {
		t.Fatalf("len(UpgradeCalls) = %d, want 1", len(fakeHelm.UpgradeCalls))
	}

	call := fakeHelm.UpgradeCalls[0]
	if call.Release != "dip-studio" {
		t.Fatalf("call.Release = %q, want dip-studio", call.Release)
	}

	switch got := call.Values["replicaCount"].(type) {
	case int:
		if got != 2 {
			t.Fatalf("call.Values[replicaCount] = %#v, want 2", got)
		}
	case float64:
		if got != 2 {
			t.Fatalf("call.Values[replicaCount] = %#v, want 2", got)
		}
	default:
		t.Fatalf("call.Values[replicaCount] type = %T, want int or float64", got)
	}

	image, ok := call.Values["image"].(map[string]interface{})
	if !ok {
		t.Fatalf("call.Values[image] = %#v, want map", call.Values["image"])
	}
	if got := image["registry"]; got != "example.com/demo" {
		t.Fatalf("call.Values[image][registry] = %#v, want example.com/demo", got)
	}
}

func TestInstallWithValuesMergesReleaseValuesOverCLISetValues(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
releases:
  demo-release:
    chart: demo-release
    version: 1.0.0
    values:
      auth:
        enabled: true
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	logger := logrus.New()
	manager := &Manager{
		Helm:      helm3testing.New("kweaver", logger.Printf),
		Namespace: "kweaver",
		Logger:    logger,
	}

	err := manager.InstallWithValues(context.Background(), manifestPath, map[string]interface{}{
		"auth": map[string]interface{}{
			"enabled": false,
			"issuer":  "cli",
		},
	}, InstallOptions{
		Namespace: "kweaver",
		Timeout:   "2m",
		SetValues: map[string]interface{}{
			"auth": map[string]interface{}{
				"enabled": false,
			},
		},
	})
	if err != nil {
		t.Fatalf("InstallWithValues() error = %v", err)
	}

	fakeHelm := manager.Helm.(*helm3testing.FakeHelm3)
	if len(fakeHelm.UpgradeCalls) != 1 {
		t.Fatalf("len(UpgradeCalls) = %d, want 1", len(fakeHelm.UpgradeCalls))
	}

	auth, ok := fakeHelm.UpgradeCalls[0].Values["auth"].(map[string]interface{})
	if !ok {
		t.Fatalf("auth values = %#v, want map", fakeHelm.UpgradeCalls[0].Values["auth"])
	}
	if got := auth["enabled"]; got != true {
		t.Fatalf("auth.enabled = %#v, want true from release values override", got)
	}
	if got := auth["issuer"]; got != "cli" {
		t.Fatalf("auth.issuer = %#v, want cli preserved from base/CLI values", got)
	}
}

func TestInstallWithValuesHonorsOptionalDependencyEnabledIf(t *testing.T) {
	dir := t.TempDir()
	dependencyPath := filepath.Join(dir, "isf.yaml")
	dependencyManifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: isf
version: 1.0.0
source:
  helmRepoName: local
releases:
  isf-server:
    chart: isf-server
    version: 1.0.0
`
	if err := os.WriteFile(dependencyPath, []byte(dependencyManifest), 0o600); err != nil {
		t.Fatalf("write dependency manifest: %v", err)
	}

	rootPath := filepath.Join(dir, "manifest.yaml")
	rootManifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
dependencies:
  - product: isf
    version: 1.0.0
    manifest: ./isf.yaml
    defaultEnabled: true
    enabledIf: auth.enabled
releases:
  demo-release:
    chart: demo-release
    version: 1.0.0
`
	if err := os.WriteFile(rootPath, []byte(rootManifest), 0o600); err != nil {
		t.Fatalf("write root manifest: %v", err)
	}

	logger := logrus.New()
	manager := &Manager{
		Helm:      helm3testing.New("kweaver", logger.Printf),
		Namespace: "kweaver",
		Logger:    logger,
	}

	err := manager.InstallWithValues(context.Background(), rootPath, map[string]interface{}{
		"auth": map[string]interface{}{
			"enabled": false,
		},
	}, InstallOptions{
		Namespace: "kweaver",
		Timeout:   "2m",
		SetValues: map[string]interface{}{
			"auth": map[string]interface{}{
				"enabled": false,
			},
		},
	})
	if err != nil {
		t.Fatalf("InstallWithValues() error = %v", err)
	}

	fakeHelm := manager.Helm.(*helm3testing.FakeHelm3)
	if len(fakeHelm.UpgradeCalls) != 1 {
		t.Fatalf("len(UpgradeCalls) = %d, want 1 when dependency disabled", len(fakeHelm.UpgradeCalls))
	}
	if got := fakeHelm.UpgradeCalls[0].Release; got != "demo-release" {
		t.Fatalf("UpgradeCalls[0].Release = %q, want demo-release", got)
	}
}

func TestUninstallUsesReverseInstallOrder(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
releases:
  dip-data-migrator:
    chart: dip-data-migrator
    version: 1.0.0
    stage: pre
  studio-web:
    chart: studio-web
    version: 1.0.0
  flow-web:
    chart: flow-web
    version: 1.0.0
    dependsOn:
      - studio-web
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	logger := logrus.New()
	fakeHelm := helm3testing.New("kweaver", logger.Printf)
	manager := &Manager{
		Helm:      fakeHelm,
		Namespace: "kweaver",
		Logger:    logger,
	}

	err := manager.Uninstall(context.Background(), manifestPath, UninstallOptions{
		Namespace: "kweaver",
		Timeout:   "2m",
	})
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}

	got := strings.Join(fakeHelm.UninstallCalls, ",")
	want := "flow-web,studio-web,dip-data-migrator"
	if got != want {
		t.Fatalf("UninstallCalls = %q, want %q", got, want)
	}
}

func TestUninstallKeepsStudioWebUntilAllDependentsAreRemoved(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: kweaver-dip
version: 0.5.0
source:
  helmRepoName: local
releases:
  dip-data-migrator:
    chart: dip-data-migrator
    version: 1.0.0
    stage: pre
  deploy-web:
    chart: deploy-web
    version: 1.0.0
  studio-web:
    chart: studio-web
    version: 1.0.0
  business-system-service:
    chart: business-system-service
    version: 1.0.0
  data-semantic:
    chart: data-semantic
    version: 1.0.0
  task-center:
    chart: task-center
    version: 1.0.0
  business-system-frontend:
    chart: business-system-frontend
    version: 1.0.0
    dependsOn:
      - studio-web
  mf-model-manager-nginx:
    chart: mf-model-manager-nginx
    version: 1.0.0
    dependsOn:
      - studio-web
  vega-web:
    chart: vega-web
    version: 1.0.0
    dependsOn:
      - studio-web
  operator-web:
    chart: operator-web
    version: 1.0.0
    dependsOn:
      - studio-web
  agent-web:
    chart: agent-web
    version: 1.0.0
    dependsOn:
      - studio-web
  flow-web:
    chart: flow-web
    version: 1.0.0
    dependsOn:
      - studio-web
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	logger := logrus.New()
	fakeHelm := helm3testing.New("kweaver", logger.Printf)
	manager := &Manager{
		Helm:      fakeHelm,
		Namespace: "kweaver",
		Logger:    logger,
	}

	err := manager.Uninstall(context.Background(), manifestPath, UninstallOptions{
		Namespace: "kweaver",
		Timeout:   "2m",
	})
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}

	var studioIdx, mfIdx int = -1, -1
	for i, release := range fakeHelm.UninstallCalls {
		if release == "studio-web" {
			studioIdx = i
		}
		if release == "mf-model-manager-nginx" {
			mfIdx = i
		}
	}

	if studioIdx == -1 || mfIdx == -1 {
		t.Fatalf("UninstallCalls = %v, want both studio-web and mf-model-manager-nginx", fakeHelm.UninstallCalls)
	}
	if mfIdx > studioIdx {
		t.Fatalf("UninstallCalls = %v, want mf-model-manager-nginx before studio-web", fakeHelm.UninstallCalls)
	}
}

func TestUninstallUsesShorterDefaultTimeout(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `apiVersion: deploy.kweaver.ai/v1alpha1
kind: VersionSet
product: demo
version: 1.0.0
source:
  helmRepoName: local
releases:
  demo-release:
    chart: demo-release
    version: 1.0.0
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	logger := logrus.New()
	fakeHelm := helm3testing.New("kweaver", logger.Printf)
	manager := &Manager{
		Helm:      fakeHelm,
		Namespace: "kweaver",
		Logger:    logger,
	}

	err := manager.Uninstall(context.Background(), manifestPath, UninstallOptions{
		Namespace: "kweaver",
	})
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}

	if len(fakeHelm.UninstallCallDetails) != 1 {
		t.Fatalf("len(UninstallCallDetails) = %d, want 1", len(fakeHelm.UninstallCallDetails))
	}
	if got := fakeHelm.UninstallCallDetails[0].Timeout; got != time.Minute {
		t.Fatalf("default uninstall timeout = %v, want %v", got, time.Minute)
	}
}
