package completion

import "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"

func CompletionCR(c *configuration.Cr) {
	if c.External != nil {
		if c.External.ChartRepo == configuration.RepoDefault {
			c.External.ChartRepo = configuration.RepoChartmuseum
		}
		if c.External.ImageRepo == configuration.RepoDefault {
			c.External.ImageRepo = configuration.RepoRegistry
		}
	}
}
