/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package backup

import (
	"fmt"

	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/shellcommand"
)

// backup schedule 命令获取定时任务
var backupScheduleGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "get backup  schedule",
	Example: `proton-cli backup  schedule get`,
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := shellcommand.RunCommand("/bin/bash", "-c", `printf "%s" "$(crontab -l 2>/dev/null| grep "proton-cli backup create")"`)
		if err != nil {
			return err
		}
		if content == "" {
			fmt.Println("no found!")
		} else {

			fmt.Println(content)
		}
		return nil
	},
}
