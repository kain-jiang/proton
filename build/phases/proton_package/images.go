package proton_package

import (
	"errors"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/pkg/image"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/workflow"
)

func NewPhaseImages() workflow.Phase {
	return workflow.Phase{
		Name:  "images",
		Short: "Pull container images",
		Run: func(c workflow.RunData) error {
			data, ok := c.(ProtonPackageData)
			if !ok {
				return errors.New("images phase invoked with a invalid struct")
			}

			if err := image.CreateProtonPackageImagesDirectoryInWorkspace(data.WorkspaceDir()); err != nil {
				return err
			}

			for _, r := range image.GenerateImageReferences() {
				if err := image.PullFromHarbor(data.Executor(), data.Harbor(), data.WorkspaceDir(), &r, data.Architecture()); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
