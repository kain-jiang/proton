package firewall

import (
	"fmt"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

type moduleUnsupportedMode struct {
	mode configuration.FirewallMode
}

// Apply implements Interface.
func (m *moduleUnsupportedMode) Apply() error {
	return fmt.Errorf("unsupported firewall mode: %v", m.mode)
}

var _ Interface = &moduleUnsupportedMode{}
