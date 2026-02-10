/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package backup

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/shellcommand"
)

var (
	// 定时任务
	schedule string
)

// backup schedule update命令获取定时任务
var UpdateScheduleGetCmd = &cobra.Command{
	Use:     "update",
	Short:   "update backup  schedule",
	Example: `proton-cli backup  schedule update  --schedule="0 2 * * * proton-cli backup create --resources all" `,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !strings.Contains(schedule, "proton-cli backup create") {
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
		if count == 0 {
			return errors.New("The backup task does not exist, please create it")
		}
		_, err = shellcommand.RunCommand("/bin/bash", "-c", "(crontab -l |grep -v \"proton-cli backup create\";echo \""+schedule+"\")|crontab -")
		if err != nil {
			return err
		}
		fmt.Println("update schedule success:" + schedule)
		return nil
	},
}

func init() {
	UpdateScheduleGetCmd.Flags().StringVar(&schedule, "schedule", "", "Timing task information")
	if err := UpdateScheduleGetCmd.MarkFlagRequired("schedule"); err != nil {
		panic(err)
	}
}
