package testing

import (
	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
)

type Command struct {
	Exec string
	Args []string

	Stdout, Stderr []byte

	ExitCode int
}

// Output implements v1alpha1.Command.
func (c *Command) Output() (out []byte, err error) {
	out = make([]byte, len(c.Stdout))
	copy(out, c.Stdout)

	if c.ExitCode == 0 {
		return
	}

	ee := &exec.ErrExitError{ExitCode: c.ExitCode}
	ee.Stderr = make([]byte, len(c.Stderr))
	copy(ee.Stderr, c.Stderr)

	err = ee

	return
}

// Run implements v1alpha1.Command.
func (c *Command) Run() (err error) {
	_, err = c.Output()
	return
}

var _ exec.Command = &Command{}
