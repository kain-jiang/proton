/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package backup

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/shellcommand"

	"github.com/spf13/cobra"
)

var (
	// 定时任务
	scheduleInfo string
)

// backup schedule create命令创建linux定时任务
var CreateScheduleGetCmd = &cobra.Command{
	Use:     "create",
	Short:   "create backup schedule",
	Example: `proton-cli backup schedule create --schedule="0 2 * * * proton-cli backup create --resources all" `,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !strings.Contains(scheduleInfo, "proton-cli backup create") {
			return errors.New("The backup task must be proton-cli backup create")
		}
		content, err := shellcommand.RunCommand("/bin/bash", "-c", `crontab -l 2>/dev/null| grep "proton-cli backup create" | wc -l`)
		if err != nil {
			return err
		}
		content = strings.Replace(content, "\n", "", -1)
		count, err := strconv.Atoi(content)
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New("Backup task task already exists")
		}
		_, err = shellcommand.RunCommand("/bin/bash", "-c", "(crontab -l |grep -v \"proton-cli backup create\";echo \""+scheduleInfo+"\")|crontab -")
		if err != nil {
			return err
		}
		fmt.Println("create schedule success :" + scheduleInfo)
		return nil
	},
}

func init() {
	CreateScheduleGetCmd.Flags().StringVar(&scheduleInfo, "schedule", "", "Timing task information")
	if err := CreateScheduleGetCmd.MarkFlagRequired("schedule"); err != nil {
		panic(err)
	}
}
