/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package backup

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/backup"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

var (
	// 备份名
	backupname string
	// 资源清单支持资源多选，以逗号分隔
	resources []string
	// 跳过的资源清单。支持资源多选，以逗号分隔
	skipResources []string
	// 备份压缩包有效时间，默认3天
	ttl int
	// 多副本 MariaDB，只有此配置为 true 时才备份
	backupMariaDB bool
	// 多副本 MongoDB，只有此配置为 true 时才备份
	backupMongoDB bool
	// 备份路径，非空时覆盖 backupConf 的 BackupDirectory
	backupDirectory string
)

// backup create命令
var backupCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Backup the current node configuration file",
	Example: `proton-cli backup create --backupname --resources xxxxx --ttl 3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := backup.BackupOpts{
			Ttl:                ttl,
			BackupName:         backupname,
			Resource:           resources,
			SkipBackupResource: skipResources,
			Id:                 time.Now().Format("20060102150405") + strconv.Itoa(time.Now().Nanosecond()),
			BackupMariaDB:      backupMariaDB,
			BackupMongoDB:      backupMongoDB,
			BackupDirectory:    backupDirectory,
		}
		hook := &logger.FileHook{
			FileName: filepath.Join(backup.BackupLogDir, opts.Id+".log"),
		}
		backup.Backuplog.AddHook(hook)

		// 多副本 MariaDB 必须通过命令行参数指定备份路径
		if opts.BackupMariaDB && opts.BackupDirectory == "" {
			return errors.New("--backup-directory is required to backup mariadb")
		}
		// 多副本 MongoDB 必须通过命令行参数指定备份路径
		if opts.BackupMongoDB && opts.BackupDirectory == "" {
			return errors.New("--backup-directory is required to backup mongodb")
		}

		var err = backup.CleanupExpiredBackUp()
		if err != nil {
			backup.Backuplog.Error(err)
			return err
		}
		err = backup.CreateBackUp(opts)
		if err != nil {
			backup.Backuplog.Error(err)
			return err
		}
		return nil
	},
}

func init() {
	//备份创建参数
	backupCreateCmd.Flags().StringVar(&backupname, "backupname", time.Now().Format("20060102150405")+strconv.Itoa(time.Now().Nanosecond()), "backup file name")
	backupCreateCmd.Flags().StringSliceVar(&resources, "resources", []string{"all"}, "Backup resource list")
	backupCreateCmd.Flags().StringSliceVar(&skipResources, "skip-resources", []string{""}, "skipped resources backup list")
	backupCreateCmd.Flags().BoolVar(&backupMariaDB, "backup-mariadb", backupMariaDB, "backup proton mariadb")
	backupCreateCmd.Flags().BoolVar(&backupMongoDB, "backup-mongodb", backupMongoDB, "backup proton mongodb")
	backupCreateCmd.Flags().StringVar(&backupDirectory, "backup-directory", backupDirectory, fmt.Sprintf("backup directory, use %s backupDirectory for empty value.", filepath.Join(backup.BackupDir, backup.BackupConfName)))
	backupCreateCmd.Flags().IntVar(&ttl, "ttl", 3, "backup valid time")
	backupCreateCmd.MarkFlagsRequiredTogether("resources")
	if err := backupCreateCmd.MarkFlagRequired("resources"); err != nil {
		panic(err)
	}
}
