package v1alpha1

import (
	"bytes"
	"os/exec"

	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

type LocalShellExecutor struct {
	log logrus.FieldLogger
}

func NewLocalShellExecutor() Executor {
	return &LocalShellExecutor{log: logger.NewLogger()}
}

type LocalShellCommand struct {
	command   string
	arguments []string

	log logrus.FieldLogger
}

// Command implements Executor.
func (e *LocalShellExecutor) Command(cmd string, args ...string) Command {
	return &LocalShellCommand{
		command:   cmd,
		arguments: args,
		log:       e.log,
	}
}

// Output implements Command.
func (c *LocalShellCommand) Output() (out []byte, err error) {
	c.log.WithField("command", c.command).Debug("execute command via localshell")
	execCmd := exec.Command(c.command, c.arguments...)
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr
	err = execCmd.Run()
	ostr, estr := stdout.String(), stderr.String()
	c.log.WithFields(logrus.Fields{
		"command":   c.command,
		"arguments": c.arguments,
		"stdout":    ostr,
		"stderr":    estr,
	}).Debug("execute command via localshell")
	if err != nil {
		c.log.WithField("command", c.command).Debugf("error execute command: %v", err)
		exitCode := 1
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		err = &ErrExitError{ExitCode: exitCode, Stderr: []byte(estr)}
	}
	return []byte(ostr), err
}

// Run implements Command.
func (c *LocalShellCommand) Run() (err error) {
	_, err = c.Output()
	return
}
