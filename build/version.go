package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/utils/exec"
)

type generateVersionOptions struct {
	buildNumber string
}

func (opts *generateVersionOptions) AddFlags(s *pflag.FlagSet) {
	s.StringVar(&opts.buildNumber, "build-number", opts.buildNumber, "Azure DevOps pipeline build number")
}

func newCommandGenerateVersion() *cobra.Command {
	opts := &generateVersionOptions{}

	cmd := &cobra.Command{
		Use:   "generate-version",
		Short: "Generate version from git, pipeline, etc.",
		Args:  cobra.NoArgs,
		RunE:  runCommandGenerateVersion(opts),
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func runCommandGenerateVersion(opts *generateVersionOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		executor := exec.New()
		out, err := executor.Command("git", "describe", "--tags", "--match=v*").Output()
		if err != nil {
			return err
		}

		trim := strings.TrimSpace(string(out))

		version, err := generateVersion(trim, opts.buildNumber)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), version)

		return nil
	}
}

func generateVersion(in, number string) (out string, err error) {
	// remove prefix v
	in = strings.TrimPrefix(in, "v")

	fields := strings.Split(in, "-")
	switch len(fields) {
	// v3.8.0
	case 1:
		out = in
		if number != "" {
			out = out + "+" + number
		}
	// v3.8.0-beta
	case 2:
		out = in
		if number != "" {
			out = out + "+" + number
		}
	// v3.8.0-34-g535d7fc
	case 3:
		out = fields[0] + "-" + fields[1] + "+" + fields[2][1:]
		if number != "" {
			out = out + "." + number
		}
	// v3.8.0-beta-34-g535d7fc
	case 4:
		out = fields[0] + "-" + fields[1] + "." + fields[2] + "+" + fields[3][1:]
		if number != "" {
			out = out + "." + number
		}
	default:
		err = fmt.Errorf("unsupported git describe output: %q", in)
	}

	return
}
