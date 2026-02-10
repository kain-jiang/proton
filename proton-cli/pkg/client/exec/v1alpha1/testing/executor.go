package testing

import (
	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
)

type Executor struct {
}

// Command implements v1alpha1.Executor.
func (e *Executor) Command(cmd string, args ...string) exec.Command {
	return &Command{Exec: cmd, Args: args}
}

var _ exec.Executor = &Executor{}
