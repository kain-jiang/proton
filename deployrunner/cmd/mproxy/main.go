package main

import (
	"os"
)

//	@title			安装/更新执行器HTTP API文档
//	@version		0.1.0
//	@description	安装/更新执行器HTTP API接口文档

//	@contact.name	API Support
//	@contact.email	tiga.gan@aishu.cn

//	@BasePath	/

func main() {
	cmd := proxyCmd
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
