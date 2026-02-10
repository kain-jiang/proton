/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package backup

import (
	"github.com/spf13/cobra"
)

// backup schedule命令
var backupScheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "get backup schedule",
	Long: `show proton backup schedule. For example:
	proton-cli backup schedule get or proton-cli backup  schedule update  --schedule="0 2 * * * proton-cli backup create --resources all"`,
	DisableSuggestions: false,
}

func init() {
	backupScheduleCmd.AddCommand(backupScheduleGetCmd)
	backupScheduleCmd.AddCommand(UpdateScheduleGetCmd)
	backupScheduleCmd.AddCommand(CreateScheduleGetCmd)
}
