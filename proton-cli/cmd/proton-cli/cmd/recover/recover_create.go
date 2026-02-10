/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package recover

import (
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/recover"
)

var (
	FromBackup  string
	RecoverName string
	resources   []string
	// 跳过的资源清单。支持资源多选，以逗号分隔
	skipResources []string
	// 备份压缩包有效时间，默认7天
)

// recover create命令
var recoverCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "recover the current node configuration file",
	Example: `proton-cli recover create --recovername --resources xxxxx --from-backup yyyyyyyy`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := recover.RecoverOpts{
			RecoverName:         RecoverName,
			Resource:            resources,
			FromBackup:          FromBackup,
			SkipRecoverResource: skipResources,
			Id:                  time.Now().Format("20060102150405") + strconv.Itoa(time.Now().Nanosecond()),
		}
		hook := &logger.FileHook{
			FileName: filepath.Join(recover.RecoverLogDir, opts.Id+".log"),
		}
		recover.Recoverlog.AddHook(hook)
		var err = recover.CreateRecover(opts)
		if err != nil {
			recover.Recoverlog.Error(err)
			return err
		}
		return nil
	},
}

func init() {
	//备份创建参数
	recoverCreateCmd.Flags().StringVar(&RecoverName, "recovername", time.Now().Format("20060102150405")+strconv.Itoa(time.Now().Nanosecond()), "recover name")
	recoverCreateCmd.Flags().StringVar(&FromBackup, "from-backup", "", "backup name")
	recoverCreateCmd.Flags().StringSliceVar(&resources, "resources", []string{""}, "recover resource list")
	recoverCreateCmd.Flags().StringSliceVar(&skipResources, "skip-resources", []string{""}, "skipped resources recover list")
	recoverCreateCmd.MarkFlagsRequiredTogether("resources", "from-backup")
	if err := recoverCreateCmd.MarkFlagRequired("resources"); err != nil {
		panic(err)
	}
}
