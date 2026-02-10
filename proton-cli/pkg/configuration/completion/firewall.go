package completion

import "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"

// CompleteFirewall 补全防火墙配置
func CompleteFirewall(c *configuration.Firewall) {
	if c.Mode == "" {
		c.Mode = configuration.FirewallFirewalld
	}
}
