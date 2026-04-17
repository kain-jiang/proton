/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package edit

import (
	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit proton resources",
	Long: `Edit proton resources such as configuration.
For example:
    proton-cli edit conf`,
}

func NewEditCommand() *cobra.Command {
	return editCmd
}
