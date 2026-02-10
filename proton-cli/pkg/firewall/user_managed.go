package firewall

import (
	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

type moduleUserManaged struct {
	logger *logrus.Logger
}

// Apply implements Interface.
func (m *moduleUserManaged) Apply() error {
	m.logger.WithField("mode", configuration.FirewallUserManaged).Debug("skip applying firewall")
	return nil
}

var _ Interface = &moduleUserManaged{}
