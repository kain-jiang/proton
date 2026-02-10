/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package backup

import (
	"github.com/spf13/cobra"
)

// backup命令
var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup the current node configuration file",
	Long: `show proton backup list. For example:
	proton-cli backup create or proton-cli backup list`,
	DisableSuggestions: false,
}

func init() {
	BackupCmd.AddCommand(backupListCmd)
	BackupCmd.AddCommand(backupCreateCmd)
	BackupCmd.AddCommand(backupLogCmd)
	BackupCmd.AddCommand(backupDirectoryCmd)
	BackupCmd.AddCommand(backupScheduleCmd)
}
