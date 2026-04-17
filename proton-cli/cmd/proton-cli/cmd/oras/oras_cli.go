package oras

import (
	"github.com/spf13/cobra"
	"oras.land/oras/cmd/oras/root"
)

func NewOrasCommand() *cobra.Command { return root.New() }
