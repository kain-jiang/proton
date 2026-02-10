/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package backup

import (
	"fmt"

	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/backup"
)

// backup directory命令获取备份目录
var backupDirectoryGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "get backup directory",
	Example: `proton-cli backup directory get`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := backup.GetBackupConf()
		if err != nil {
			return err
		}
		if conf != nil && conf.BackupDirectory != "" {
			fmt.Println(conf.BackupDirectory)

		} else {
			backup.Backuplog.Info("no backup config")
		}
		return nil
	},
}
