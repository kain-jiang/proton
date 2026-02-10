package completion

import (
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/prometheus"
)

func CompletePrometheus(c *configuration.Prometheus) {
	if len(c.Hosts) > 0 && c.DataPath == "" {
		c.DataPath = prometheus.DefaultDataPath
	}
}
