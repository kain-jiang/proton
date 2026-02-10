package version

import (
	"fmt"

	"taskrunner/trait"

	"github.com/spf13/cobra"
)

// Version binary version
var Version string

func init() {
	Version = trait.GitVersion
}

// NewVersionCmd  return a version cmd
func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "version",
		Long: `task runner manager use to manage application system.
			manager can create new job record but don't execute.
			manager can has multi replica.`,
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}
	return cmd
}
