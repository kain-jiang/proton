package tool

import (
	"log/slog"
	"os"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/pkg/utils"
)

func MoveProtonCLI(workspace, src string) error {
	slog.Debug("Open proton-cli archive", "path", src)
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	path := generateProtonPackageToolPath(workspace, "proton-cli")

	slog.Info("Copy proton-cli")
	if err := utils.CreateFileFrom(path, f); err != nil {
		return err
	}

	if err := os.Chmod(path, 0755); err != nil {
		return err
	}

	return nil
}
