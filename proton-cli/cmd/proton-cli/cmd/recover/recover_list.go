/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package recover

import (
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/recover"
)

// recover list命令
var recoverListCmd = &cobra.Command{
	Use:     "list",
	Short:   "get the current node recover list",
	Example: `proton-cli recover list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := recover.GetRecoverConf()
		if err != nil {
			return err
		}
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetAutoIndex(true)
		t.AppendHeader(table.Row{"recovername", "Resources", "HostName", "CreateTime", "RunTime", "BackupPath", "status"})
		t.SetColumnConfigs([]table.ColumnConfig{
			{
				Name:     "Resources",
				Align:    text.AlignLeft,
				WidthMax: 64,
			},
		})
		if conf != nil && len(conf.List) > 0 {
			for _, info := range conf.List {
				var status = "success"
				if !info.Status {
					status = "fail"
				}
				t.AppendRow(table.Row{info.Name, info.Resource, conf.HostName, time.Unix(info.CreateTime, 0).Format("2006-01-02 15:04:05"), info.RunTime, info.FromBackupPath, status})
			}
		} else {
			t.SetCaption("无还原配置文件或还原记录.\n")
		}
		t.AppendSeparator()
		t.Render()
		return nil
	},
}
