/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package backup

import (
	"github.com/spf13/cobra"
)

// backup directory命令
var backupDirectoryCmd = &cobra.Command{
	Use:   "directory",
	Short: "get backup directory",
	Long: `show proton backup directory. For example:
	proton-cli backup directory get or proton-cli backup directory update --path /xxx`,
	DisableSuggestions: false,
}

func init() {
	backupDirectoryCmd.AddCommand(backupDirectoryUpdateCmd)
	backupDirectoryCmd.AddCommand(backupDirectoryGetCmd)
}
