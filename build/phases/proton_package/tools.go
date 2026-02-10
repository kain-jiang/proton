package proton_package

import (
	"errors"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/pkg/tool"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/workflow"
)

func NewPhaseTools() workflow.Phase {
	return workflow.Phase{
		Name:  "tools",
		Short: "Download static binary tools",
		Run:   runPhaseTools,
		InheritFlags: []string{
			"output",
			"architecture",
			"proton-cli-path",
			"tool",
			"version",
		},
	}
}

func runPhaseTools(c workflow.RunData) error {
	data, ok := c.(ProtonPackageData)
	if !ok {
		return errors.New("tools phase invoked with a invalid struct")
	}

	if err := tool.CreateProtonPackageToolsDirectoryInWorkspace(data.WorkspaceDir()); err != nil {
		return err
	}

	for _, n := range tool.GenerateStaticToolsNames() {
		if err := tool.DownloadStaticTool(data.WorkspaceDir(), data.ToolRepositoryURL(), data.DistroArch(), n); err != nil {
			return err
		}
	}

	if err := tool.MoveProtonCLI(data.WorkspaceDir(), data.ProtonCLIPath()); err != nil {
		return err
	}

	return nil
}
