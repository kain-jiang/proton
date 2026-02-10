package cmd

import (
	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/precheck"
)

var (
	sshpassword string
	ntpServer   string
)

var preCheckCmd = &cobra.Command{
	Use:   "precheck",
	Short: "check node environment before install",
	Run: func(cmd *cobra.Command, args []string) {
		precheck.PreCheck(sshpassword, ntpServer)
	},
}

func init() {
	rootCmd.AddCommand(preCheckCmd)
	preCheckCmd.Flags().StringVarP(&sshpassword, "password", "p", "", "SSH password")
	preCheckCmd.Flags().StringVarP(&ntpServer, "ntpserver", "t", "", "NTP Server Address")
}
