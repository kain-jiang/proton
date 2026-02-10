package registry

import (
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
)

type Config struct {
	// container registry address, such as host or host:port
	Address string
}

func ConfigForCR(cr *configuration.Cr) *Config {
	r, _, _ := global.ImageRepository(cr)
	return &Config{Address: r}
}
