package grafana

import "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/check"

func (m *Manager) checkEnvironment() error {
	var checkers []check.Checker

	if m.Spec.Hosts != nil && m.Spec.DataPath != "" {
		checkers = append(checkers, &check.NodeDirAvailableChecker{Node: m.Node.Name(), Path: m.Spec.DataPath, Files: m.Node.ECMS().Files()})
	}

	return check.RunChecks(checkers, m.Logger, nil)
}
