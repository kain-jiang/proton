package proton_package

import (
	"errors"
	"log/slog"
	"os"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/pkg/tarball"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/workflow"
)

func NewPhaseTarball() workflow.Phase {
	return workflow.Phase{
		Name:  "tarball",
		Short: "Create release tarball",
		Run:   runPhaseTarball,
		InheritFlags: []string{
			"architecture",
			"output",
			"version",
		},
	}
}

func runPhaseTarball(c workflow.RunData) error {
	data, ok := c.(ProtonPackageData)
	if !ok {
		return errors.New("tarball phase invoked with a invalid struct")
	}

	{
		path := data.ReleaseDir()
		slog.Info("Create release directory", "path", path)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return tarball.CreateTarball(data.WorkspaceDir(), data.ReleaseDir(), data.ReleaseName())
}
