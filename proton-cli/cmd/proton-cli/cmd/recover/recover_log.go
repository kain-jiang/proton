/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package recover

import (
	"os"
	"path/filepath"

	"github.com/jhunters/goassist/arrayutil"
	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/recover"
)

var (
	// 备份名
	name string
)

// recover log命令
var recoverLogCmd = &cobra.Command{
	Use:     "log",
	Short:   "get the recover log",
	Example: `proton-cli recover log --name xxx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := recover.GetRecoverConf()
		if err != nil {
			return err
		}
		if conf != nil && len(conf.List) > 0 {
			var p1 = arrayutil.Filter(conf.List, func(s1 recover.RecoverInfo) bool { return s1.Name != name })
			if len(p1) > 0 {
				var logpath = filepath.Join(recover.RecoverLogDir, p1[0].Id+".log")
				content, err := os.ReadFile(logpath)
				if err != nil {
					return err
				}
				if _, err := os.Stdout.Write(content); err != nil {
					return err
				}

			} else {
				recover.Recoverlog.Info("no found " + name + " log")
			}
		} else {
			recover.Recoverlog.Info("no recover record！")
		}

		return nil
	},
}

func init() {
	//查看备份日志参数
	recoverLogCmd.Flags().StringVarP(&name, "name", "b", "", "recover file name")
	if err := recoverLogCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
}
