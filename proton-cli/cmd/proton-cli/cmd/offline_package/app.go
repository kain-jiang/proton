package offline_package

import "github.com/spf13/cobra"

func newAppCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Export or import offline application packages",
	}

	cmd.AddCommand(newAppExportCommand())
	cmd.AddCommand(newAppImportCommand())

	return cmd
}
