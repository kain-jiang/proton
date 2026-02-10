package store

import (
	"fmt"
	"path"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/store/helm"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/universal"
)

func (m *Manager) Apply() error {
	if m.Spec.Hosts != nil {
		m.Logger.Debug("reconcile data directory")
		if err := m.reconcileDataDirectories(); err != nil {
			return fmt.Errorf("reconcile data directory fail: %w", err)
		}
	}

	m.Logger.Debug("reconcile helm release")
	cht := m.ServicePackage.Charts().Get(helm.ChartName, "")
	if err := m.Helm.Upgrade(
		helm.ReleaseName,
		&helm3.ChartRef{File: path.Join(m.ServicePackage.BaseDir(), cht.Path)},
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeValues(helm.ValuesFor(m.Spec, m.Registry, m.RDS, DatabaseName, m.Namespace).ToMap()),
	); err != nil {
		return fmt.Errorf("reconcile helm release %v fail: %w", helm.ReleaseName, err)
	}

	return nil
}

// reconcileDataDirectories creates data directories on nodes if it not exists.
func (m *Manager) reconcileDataDirectories() error {
	for _, n := range m.Nodes {
		if err := universal.ReconcileDataDirectory(n, m.Spec.Storage.Path, m.Logger.WithField("node", n.Name())); err != nil {
			return fmt.Errorf("%v: %w", n.Name(), err)
		}
	}
	return nil
}
