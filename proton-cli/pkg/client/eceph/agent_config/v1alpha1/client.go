package v1alpha1

import (
	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
)

type Client struct {
	executor exec.Executor
}

func New(e exec.Executor) *Client {
	return &Client{executor: e}
}

const AgentConfigPath = "/opt/minotaur/tools/agent_config"
const ECephConfigAgentHealthCheck = "curl http://127.0.0.1:14321"

func (c *Client) ExecuteWithTimeout(timeout, nodeIP, internalIP, objectIP, hostname string) error {
	var args = []string{
		"-f=init_config",
		"--node_ip=" + nodeIP,
		"--internal_ip=" + internalIP,
		"--object_ip=" + nodeIP,
		"--host_name=" + hostname,
	}
	return c.executor.Command(AgentConfigPath, args...).Run()
}

func (c *Client) ShouldExecuteAgentConfig() error {
	return c.executor.Command(ECephConfigAgentHealthCheck).Run()
}

var _ Interface = &Client{}
