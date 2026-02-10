package universal

import (
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
)

// HelmV3Manager 基于 Helm V3 部署、更新 Proton、AnyShare 服务
type HelmV3Manager struct {
	Release   string
	ChartFile string
	Namespace string
	Helm3     helm3.Client

	Values map[string]interface{}
}

func (m *HelmV3Manager) Apply() error {
	return m.Helm3.Upgrade(
		m.Release,
		helm3.ChartRefFromFile(m.ChartFile),
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeAtoMic(false),
		helm3.WithUpgradeValues(m.Values),
	)
}
