package completion

import (
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

// CompletionChrony 补全 Chrony 时间源服务器配置
func CompleteChrony(ch *configuration.Chrony, cs *configuration.Cs) *configuration.Chrony {
	// 如果时间源配置为空，则设置为UserManaged模式
	if ch == nil {
		ch = &configuration.Chrony{
			Mode:   configuration.ChronyModeUserManaged,
			Server: []string{},
		}
		return ch
	}
	// 如果时间源配置为内置主节点，且配置文件中未指出使用哪个节点，则选择列表中第一个主节点。是否为真随机选择并不重要
	if ch.Mode == configuration.ChronyModeLocalMaster {
		if len(ch.Server) == 0 {
			ch.Server = []string{cs.Master[0]}
		}
	}
	return ch
}
