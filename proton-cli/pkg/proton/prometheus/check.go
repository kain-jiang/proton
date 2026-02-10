package prometheus

import (
	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/check"
)

// checkEnvironment checks whether the environment meets the requirements of
// Prometheus configuration
func checkEnvironment(spec *configuration.Prometheus, nodes []v1alpha1.Interface, logger logrus.FieldLogger) error {
	var checkers []check.Checker
	for _, n := range nodes {
		checkers = append(checkers, &check.NodeDirAvailableChecker{Node: n.Name(), Path: spec.DataPath, Files: n.ECMS().Files()})
	}
	return check.RunChecks(checkers, logger, nil)
}
