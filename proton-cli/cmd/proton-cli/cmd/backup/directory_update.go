/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package backup

import (
	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/backup"
)

var (
	// 备份压缩存储目录
	path string
)

// backup directory命令-修改备份目录
var backupDirectoryUpdateCmd = &cobra.Command{
	Use:     "update",
	Short:   "update backup directory",
	Example: `proton-cli backup directory update --path /xxxx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := backup.UpgradeBackUpPath(path)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	//备份目录修改参数
	backupDirectoryUpdateCmd.Flags().StringVarP(&path, "path", "b", "", "backup file name")
	if err := backupDirectoryUpdateCmd.MarkFlagRequired("path"); err != nil {
		panic(err)
	}
}
