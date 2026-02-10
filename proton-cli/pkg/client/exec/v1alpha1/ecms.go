package v1alpha1

import (
	"context"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1/exec"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

type ECMSExecutor struct {
	c exec.Interface
}

// TODO: rename to NewECMSExecutor
func NewECMSExecutorForHost(c exec.Interface) *ECMSExecutor {
	return &ECMSExecutor{
		c: c,
	}
}

// Command implements Executor.
func (e *ECMSExecutor) Command(cmd string, args ...string) Command {
	return &ECMSCommand{
		c:       e.c,
		command: append([]string{cmd}, args...),
	}
}

var _ Executor = &ECMSExecutor{}

type ECMSCommand struct {
	c       exec.Interface
	command []string
}

// Output implements Command.
func (c *ECMSCommand) Output() ([]byte, error) {
	logger.NewLogger().WithField("command", c.command).Debug("execute command and receive stdout")
	return c.c.Execute(context.TODO(), c.command, nil, exec.ExecuteOptions{Stdout: true})
}

// Run implements Command.
func (c *ECMSCommand) Run() error {
	logger.NewLogger().WithField("command", c.command).Debug("execute command")
	_, err := c.c.Execute(context.TODO(), c.command, nil, exec.ExecuteOptions{})
	return err
}

var _ Command = &ECMSCommand{}
