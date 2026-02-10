package v1alpha1

import (
	"errors"

	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
)

type Client struct {
	executor exec.Executor
}

func New(e exec.Executor) *Client {
	return &Client{executor: e}
}

// Start implements Interface.
func (c *Client) Start(name string) error {
	return c.executor.Command("systemctl", "start", name).Run()
}

// Enabled implements Interface.
func (c *Client) Enabled(name string, now bool) error {
	args := []string{"enable", name}
	if now {
		args = append(args, "--now")
	}
	return c.executor.Command("systemctl", args...).Run()
}

// IsActive implements Interface.
func (c *Client) IsActive(name string) (bool, error) {
	err := c.executor.Command("systemctl", "is-active", name).Run()
	if err == nil {
		return true, nil
	}

	ee := new(exec.ErrExitError)
	if !errors.As(err, &ee) {
		return false, err
	}

	return ee.ExitCode == 0, nil
}

// IsEnabled implements Interface.
func (c *Client) IsEnabled(name string) (bool, error) {
	err := c.executor.Command("systemctl", "is-enabled", name).Run()
	if err == nil {
		return true, nil
	}

	ee := new(exec.ErrExitError)
	if !errors.As(err, &ee) {
		return false, err
	}

	return ee.ExitCode == 0, nil
}

var _ Interface = &Client{}
