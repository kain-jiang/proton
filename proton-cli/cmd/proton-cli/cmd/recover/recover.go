/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package recover

import (
	"github.com/spf13/cobra"
)

// recover命令
var RecoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "recover the current node configuration file and data",
	Long: `show proton recover list. For example:
	proton-cli recover create or proton-cli recover list`,
	DisableSuggestions: false,
}

func init() {
	RecoverCmd.AddCommand(recoverCreateCmd)
	RecoverCmd.AddCommand(recoverListCmd)
	RecoverCmd.AddCommand(recoverLogCmd)

}
