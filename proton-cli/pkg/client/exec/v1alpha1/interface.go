package v1alpha1

type Executor interface {
	// Command returns a Cmd instance which can be used to run a single command.
	// This follows the pattern of package os/exec.
	Command(cmd string, args ...string) Command
}

type Command interface {
	// Run runs the command to the completion.
	Run() error

	// Output runs the command and returns standard output, but not standard err
	Output() ([]byte, error)
}
