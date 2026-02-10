package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"taskrunner/trait"

	"github.com/spf13/cobra"
)

func newrootCmd() *cobra.Command {
	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newApplicationCmd(),
		NewSystemCmd(),
		newjobCmd(),
	)

	return cmd
}

func main() {
	cmd := newrootCmd()
	ctx, cancel := trait.WithCancelCauesContext(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		defer cancel(&trait.Error{
			Internal: trait.ECExit,
			Err:      fmt.Errorf("receive stop signal"),
			Detail:   "exit",
		})
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
		<-ch
	}()
	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
