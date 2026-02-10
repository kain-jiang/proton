package cms

import (
	"path/filepath"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/universal"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

const (
	// CMS 的 Chart 名称
	ChartName = configuration.ChartNameCMS
	// CMS 的 Helm release 名称，与 chart 名称一致
	ReleaseName = configuration.ReleaseNameCMS
)

func NewManager(helm3 helm3.Client, spec *configuration.CMS, registry string, servicePackage string, charts servicepackage.Charts, serviceaccount string) *universal.HelmV3Manager {
	cmsChart := charts.Get(ChartName, "")
	if serviceaccount != "" {
		return &universal.HelmV3Manager{
			Release:   ReleaseName,
			ChartFile: filepath.Join(servicePackage, cmsChart.Path),
			Namespace: configuration.GetProtonResourceNSFromFile(),
			Helm3:     helm3,
			Values: map[string]interface{}{
				"image": map[string]interface{}{
					"registry": registry,
				},
				"serviceAccount": map[string]interface{}{
					"create": false,
					"name":   serviceaccount,
				},
				"service": map[string]interface{}{
					"protoncliNamespace": configuration.GetProtonCliConfigNSFromFile(),
				},
				"namespace":    configuration.GetProtonResourceNSFromFile(),
				"nodeSelector": spec.NodeSelector,
			},
		}
	} else {
		return &universal.HelmV3Manager{
			Release:   ReleaseName,
			ChartFile: filepath.Join(servicePackage, cmsChart.Path),
			Namespace: configuration.GetProtonResourceNSFromFile(),
			Helm3:     helm3,
			Values: map[string]interface{}{
				"image": map[string]interface{}{
					"registry": registry,
				},
				"service": map[string]interface{}{
					"protoncliNamespace": configuration.GetProtonCliConfigNSFromFile(),
				},
				"namespace":    configuration.GetProtonResourceNSFromFile(),
				"nodeSelector": spec.NodeSelector,
			},
		}
	}

}
