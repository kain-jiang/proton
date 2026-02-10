package cmd

import (
	"github.com/spf13/cobra"
)

var takeoverConfigPath string

func init() {
	rootCmd.AddCommand(takeoverCmd)
	takeoverCmd.PersistentFlags().StringVar(&takeoverConfigPath, "config", "./config.yaml", "http serve config path") // serve 服务配置文件
}

var takeoverCmd = &cobra.Command{
	Use:   "takeover",
	Short: "take over configuration",
	Long:  "Take Over Other Application Configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
