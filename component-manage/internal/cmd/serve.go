package cmd

import (
	"component-manage/internal/global"
	"component-manage/internal/server"

	"github.com/spf13/cobra"
)

var serveConfigPath string

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.PersistentFlags().StringVar(&serveConfigPath, "config", "./config.yaml", "http serve config path") // serve 服务配置文件
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a http service",
	Long:  "Run a http service",
	RunE: func(cmd *cobra.Command, args []string) error {
		global.InitConfig(serveConfigPath)
		global.InitLogger(global.Config)
		global.InitK8sCli()
		global.InitHelmCli(global.Config, global.Logger)
		global.InitPersist(global.Config, global.K8sCli, global.Logger)
		return server.Main()
	},
}
