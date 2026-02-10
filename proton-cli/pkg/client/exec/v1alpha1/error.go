package v1alpha1

import "fmt"

type ErrExitError struct {
	ExitCode int

	Stderr []byte
}

// Error implements error.
func (e *ErrExitError) Error() string {
	return fmt.Sprintf("exit code: %d, stderr:\n%s", e.ExitCode, string(e.Stderr))
}

var _ error = &ErrExitError{}
