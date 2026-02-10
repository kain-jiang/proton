package cmd

import (
	"os"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionExample = dedent.Dedent(`
	# Print the version information
	proton-cli version
`)

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "print the version information",
	Example: versionExample,
	Run: func(cmd *cobra.Command, args []string) {
		b, _ := yaml.Marshal(version.Get())
		os.Stdout.Write(b)
	},
}
