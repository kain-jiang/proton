package store

import (
	"fmt"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/universal"
)

func (m *Manager) Reset() error {
	if m.Spec.Storage.Path == "" {
		return nil
	}
	for _, n := range m.Nodes {
		if err := universal.ClearDataDirViaNodeV1Alpha1(n, m.Spec.Storage.Path, m.Logger.WithField("node", n.Name())); err != nil {
			return fmt.Errorf("%s: %w", n.Name(), err)
		}
	}
	return nil
}
