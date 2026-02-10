package exec

import "fmt"

// Like os/exec.ExitError
type ExitError struct {
	// stdout, stderr or combined depend on argument
	Output []byte
	// exit code
	ExitCode int
}

func (e *ExitError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("exit status %d", e.ExitCode)
}
