/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package backup

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/backup"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/file"
)

// backup list命令
var backupListCmd = &cobra.Command{
	Use:     "list",
	Short:   "get the current node Backup list",
	Example: `proton-cli backup list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := backup.GetBackupConf()

		if err != nil {
			return err
		}
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetAutoIndex(true)
		t.AppendHeader(table.Row{"backupname", "UseSpace", "Resources", "HostName", "CreateTime", "runTime\n(second)", "backupPath", "status", "Effectivetime\n(hour)"})
		t.SetColumnConfigs([]table.ColumnConfig{
			{
				Name:     "Resources",
				Align:    text.AlignLeft,
				WidthMax: 50,
			},
			{
				Name:     "backupPath",
				Align:    text.AlignLeft,
				WidthMax: 32,
			},
			{
				Number:   6,
				Align:    text.AlignCenter,
				WidthMax: 32,
			},
			{
				Number:   9,
				Align:    text.AlignCenter,
				WidthMax: 32,
			},
		})
		if conf != nil && len(conf.List) > 0 {
			for _, info := range conf.List {
				var status = "success"
				if !info.Status {
					status = "fail"
				}
				var ttl float64
				var expirationDate = time.Unix(info.CreateTime, 0).Add(time.Hour * 24 * time.Duration(info.Ttl))
				if expirationDate.Before(time.Now()) {
					ttl = 0
				} else {
					// ttl = expirationDate.Sub(time.Now()).Hours()
					ttl = time.Until(expirationDate).Hours()
				}
				var ttlHours, _ = strconv.ParseFloat(fmt.Sprintf("%.0f", ttl), 64)
				if ttl < 1 {
					ttlHours, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", ttl), 64)
				}
				t.AppendRow(table.Row{info.Name, file.FormatFileSize(info.UseSpace), info.Resource, conf.HostName, time.Unix(info.CreateTime, 0).Format("2006-01-02 15:04:05"), info.RunTime, info.Path, status, ttlHours})
			}
		} else {
			t.SetCaption("无备份配置文件或备份列表.\n")
		}
		t.AppendSeparator()
		t.Render()
		return nil
	},
}
