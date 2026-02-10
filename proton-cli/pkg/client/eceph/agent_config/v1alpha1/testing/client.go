package testing

import (
	eceph_agent_config "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/eceph/agent_config/v1alpha1"
)

type Client struct {
	Err       error
	ErrShould error
}

// ExecuteWithTimeout implements v1alpha1.Interface.
func (c *Client) ExecuteWithTimeout(timeout string, nodeIP string, internalIP string, objectIP string, hostname string) error {
	return c.Err
}

func (c *Client) ShouldExecuteAgentConfig() error {
	return c.ErrShould
}

var _ eceph_agent_config.Interface = &Client{}
