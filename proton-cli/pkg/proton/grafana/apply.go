package grafana

import (
	"path"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/universal"
)

func (m *Manager) Apply() error {
	m.Logger.Info("applying module")

	if err := m.checkEnvironment(); err != nil {
		return err
	}

	if m.Spec.DataPath != "" {
		m.Logger.Debug("reconcile data directory")
		if err := universal.ReconcileDataDirectory(m.Node, m.Spec.DataPath, m.Logger.WithField("node", m.Node.Name())); err != nil {
			return err
		}
	}

	cht := m.ServicePackage.Charts().Get(ChartName, "")

	m.Logger.Debug("reconcile helm release")
	if err := m.Helm.Upgrade(
		HelmReleaseName,
		&helm3.ChartRef{File: path.Join(m.ServicePackage.BaseDir(), cht.Path)},
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeValues(valuesFor(m.Spec, m.Registry, m.Namespace, m.Prometheus).ToMap()),
	); err != nil {
		return err
	}

	return nil
}
