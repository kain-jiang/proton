package nvidiadeviceplugin

import (
	"path/filepath"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/universal"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

const (
	// NvidiaDevicePlugin 的 Chart 名称
	ChartName = configuration.ChartNameNvDevPlugin
	// NvidiaDevicePlugin 的 Helm release 名称，与 chart 名称一致
	ReleaseName = configuration.ReleaseNameNvDevPlugin
	// NvidiaDevicePlugin 的 Helm release 所在的命名空间
)

func NewManager(helm3 helm3.Client, spec *configuration.NvidiaDevicePlugin, registry string, servicePackage string, charts servicepackage.Charts, namespace string) *universal.HelmV3Manager {
	ndpChart := charts.Get(ChartName, "")
	return &universal.HelmV3Manager{
		Release:   ReleaseName,
		ChartFile: filepath.Join(servicePackage, ndpChart.Path),
		Namespace: namespace,
		Helm3:     helm3,
		Values: map[string]interface{}{
			"image": map[string]interface{}{
				"registry": registry,
			},
			"namespace": namespace,
		},
	}

}
