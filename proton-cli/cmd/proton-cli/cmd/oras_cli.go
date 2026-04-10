package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"oras.land/oras/cmd/oras/root"
)

func newOrasCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "oras",
		Short:              "oras OCI artifact CLI",
		Long:               "oras CLI for managing OCI artifacts and container registries",
		RunE:               runOrasCommand,
		DisableFlagParsing: true,
	}

	return cmd
}

func runOrasCommand(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	orasCmd := root.New()
	orasCmd.SetArgs(args)
	return orasCmd.ExecuteContext(ctx)
}
