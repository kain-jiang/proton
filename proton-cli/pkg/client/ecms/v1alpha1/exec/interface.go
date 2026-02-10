package exec

import (
	"context"
	"io"
)

type ExecuteOptions struct {
	Stdin  bool
	Stdout bool
	Stderr bool
}

type Interface interface {
	Execute(ctx context.Context, command []string, input io.Reader, opts ExecuteOptions) (out []byte, err error)
}
