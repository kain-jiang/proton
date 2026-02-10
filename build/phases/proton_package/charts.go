package proton_package

import (
	"errors"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/pkg/chart"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/workflow"
)

func NewPhaseCharts() workflow.Phase {
	return workflow.Phase{
		Name:  "charts",
		Short: "Pull helm charts",
		Run:   runPhaseCharts,
		InheritFlags: []string{
			"output",
			"version",
			"architecture",
			"chart",
		},
	}
}

func runPhaseCharts(c workflow.RunData) error {
	data, ok := c.(ProtonPackageData)
	if !ok {
		return errors.New("charts phase invoked with a invalid struct")
	}

	path, err := chart.CreateProtonPackageChartsDirectoryInWorkspace(data.WorkspaceDir())
	if err != nil {
		return err
	}

	if err := chart.PackageDirectoryAll(data.Executor(), data.Version(), path, data.ChartSourceDir()); err != nil {
		return err
	}

	return nil
}
