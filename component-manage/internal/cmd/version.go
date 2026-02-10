package cmd

import (
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var (
	versionStr = "version set failed"
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of component-manage",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(versionStr)
		},
	}
)

func SetVersion(v map[string]any) {
	vStr, _ := yaml.Marshal(v)
	versionStr = string(vStr)
}
