package completion

import (
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/grafana"
)

func CompleteGrafana(c *configuration.Grafana) {
	if len(c.Hosts) > 0 && c.DataPath == "" {
		c.DataPath = grafana.DefaultDataPath
	}
}
