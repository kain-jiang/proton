/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package backup

import (
	"os"
	"path/filepath"

	"github.com/jhunters/goassist/arrayutil"
	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/backup"
)

var (
	// 备份名
	name string
)

// backup log命令
var backupLogCmd = &cobra.Command{
	Use:     "log",
	Short:   "get the Backup log",
	Example: `proton-cli backup log --name xxx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := backup.GetBackupConf()
		if err != nil {
			return err
		}
		if conf != nil && len(conf.List) > 0 {
			var p1 = arrayutil.Filter(conf.List, func(s1 backup.BackupInfo) bool { return s1.Name != name })
			if len(p1) > 0 {
				var logpath = filepath.Join(backup.BackupLogDir, p1[0].Id+".log")
				content, err := os.ReadFile(logpath)
				if err != nil {
					return err
				}
				if _, err := os.Stdout.Write(content); err != nil {
					return err
				}

			} else {
				backup.Backuplog.Info("no found " + name + " log")
			}
		} else {
			backup.Backuplog.Info("no backup record！")
		}

		return nil
	},
}

func init() {
	//查看备份日志参数
	backupLogCmd.Flags().StringVarP(&name, "name", "b", "", "backup file name")
	if err := backupLogCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
}
