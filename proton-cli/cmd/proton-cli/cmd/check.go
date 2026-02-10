package cmd

import (
	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/check"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "check Proton Runtime health after initial",
	Run: func(cmd *cobra.Command, args []string) {
		check.Check()
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
