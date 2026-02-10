package proton_package

import (
	"errors"
	"log/slog"
	"os"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/workflow"
)

func NewPhaseWorkspace() workflow.Phase {
	return workflow.Phase{
		Name:  "workspace",
		Short: "Create workspace for building",
		Run:   runPhaseWorkspace,
		InheritFlags: []string{
			"version",
			"architecture",
			"output",
		},
	}
}

func runPhaseWorkspace(c workflow.RunData) error {
	data, ok := c.(ProtonPackageData)
	if !ok {
		return errors.New("workspace phase invoked with a invalid struct")
	}

	path := data.WorkspaceDir()
	slog.Info("Create workspace", "path", path)
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	return nil
}
