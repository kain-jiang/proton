/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package config

import (
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "show proton cluster conf or template conf",
	Long: `show proton cluster conf or template conf. For example:
     proton-cli get conf or proton-cli get template`,
	DisableSuggestions: false,
}

func NewGetCommand() *cobra.Command {
	return getCmd
}
