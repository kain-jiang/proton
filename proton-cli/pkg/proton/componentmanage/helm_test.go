package componentmanage

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

func TestNewManager(t *testing.T) {
	// Mock dependencies
	helm3Mock := (helm3.Client)(nil)
	oldCfgMock := &configuration.ClusterConfig{}
	newCfgMock := &configuration.ClusterConfig{
		Deploy: &configuration.Deploy{
			ServiceAccount: "serviceaccount",
		},
		Cr: &configuration.Cr{
			External: &configuration.ExternalCR{
				ChartRepo: configuration.RepoChartmuseum,
				Chartmuseum: &configuration.Chartmuseum{
					Host:     "https://chartmuseum.example.com",
					Username: "",
					Password: "",
				},
			},
		},
		ComponentManage: &configuration.ComponentManagement{},
	}
	registryMock := "registry.example.com"
	servicePackageMock := "my-service-package"
	chartsMock := servicepackage.Charts{
		{Metadata: chart.Metadata{
			Name:    ChartName,
			Version: "1.0.0",
		}},
	}
	imagesMock := make([]string, 0)

	// Call the function
	manager := NewManager(helm3Mock, oldCfgMock, newCfgMock, registryMock, servicePackageMock, chartsMock, imagesMock, "resource")

	// Assertions
	expectedChartFile := filepath.Join(servicePackageMock, chartsMock.Get(ChartName, "").Path)
	expectedValues := map[string]any{
		"image": map[string]any{
			"registry": registryMock,
		},
		"serviceAccount": map[string]any{
			"create": false,
			"name":   "serviceaccount",
		},
		"namespace": "resource",
		"service": map[string]any{
			"enableDualStack": false,
			"config": map[string]any{
				"chartmuseum": map[string]any{
					"url":      "https://chartmuseum.example.com",
					"username": "",
					"password": "",
					"enable":   true,
				},
			},
		},
		"nodeSelector": (map[string]string)(nil),
	}
	assert.Equal(t, ReleaseName, manager.Release)
	assert.Equal(t, chartsMock, manager.charts)
	assert.Equal(t, expectedChartFile, manager.ChartFile)
	assert.Equal(t, "resource", manager.Namespace)
	assert.Equal(t, helm3Mock, manager.Helm3)
	assert.Equal(t, oldCfgMock, manager.OldCfg)
	assert.Equal(t, newCfgMock, manager.NewCfg)
	assert.Equal(t, expectedValues, manager.Values)
}

func TestNewManagerWithoutDeploy(t *testing.T) {
	helm3Mock := (helm3.Client)(nil)
	oldCfgMock := &configuration.ClusterConfig{}
	newCfgMock := &configuration.ClusterConfig{
		Cr: &configuration.Cr{
			External: &configuration.ExternalCR{
				ChartRepo: configuration.RepoChartmuseum,
				Chartmuseum: &configuration.Chartmuseum{
					Host: "https://chartmuseum.example.com",
				},
			},
		},
		ComponentManage: &configuration.ComponentManagement{},
	}
	registryMock := "registry.example.com"
	servicePackageMock := "my-service-package"
	chartsMock := servicepackage.Charts{
		{Metadata: chart.Metadata{
			Name:    ChartName,
			Version: "1.0.0",
		}},
	}
	imagesMock := make([]string, 0)

	manager := NewManager(helm3Mock, oldCfgMock, newCfgMock, registryMock, servicePackageMock, chartsMock, imagesMock, "resource")

	expectedValues := map[string]any{
		"image": map[string]any{
			"registry": registryMock,
		},
		"serviceAccount": map[string]any{
			"create": true,
			"name":   "",
		},
		"namespace": "resource",
		"service": map[string]any{
			"enableDualStack": false,
			"config": map[string]any{
				"chartmuseum": map[string]any{
					"url":      "https://chartmuseum.example.com",
					"username": "",
					"password": "",
					"enable":   true,
				},
			},
		},
		"nodeSelector": (map[string]string)(nil),
	}

	assert.Equal(t, expectedValues, manager.Values)
}
