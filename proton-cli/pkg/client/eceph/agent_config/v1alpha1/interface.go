package v1alpha1

type Interface interface {
	ExecuteWithTimeout(timeout string, nodeIP string, internalIP string, objectIP string, hostname string) error
	ShouldExecuteAgentConfig() error
}
