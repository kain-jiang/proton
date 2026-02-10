package proton_package

import (
	"errors"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/pkg/rpm"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/workflow"
)

func NewPhaseRPMs() workflow.Phase {
	return workflow.Phase{
		Name:  "rpms",
		Short: "Download rpms",
		Run:   runPhaseRPMs,
	}
}

func runPhaseRPMs(c workflow.RunData) error {
	data, ok := c.(ProtonPackageData)
	if !ok {
		return errors.New("rpms phase invoked with a invalid struct")
	}

	return rpm.DownloadProtonPackageArchive(data.WorkspaceDir(), data.RPMRepositoryURL(), data.RPMRepositoryVersion(), data.DistroArch())
}
