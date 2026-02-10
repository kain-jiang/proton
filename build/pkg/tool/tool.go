package tool

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/pkg/utils"
)

func GenerateStaticToolsNames() []string {
	return []string{
		"skopeo",
		"sshpass",
		"yq",
	}
}

func generateProtonPackageToolsDirectoryPath(workspace string) string {
	return filepath.Join(workspace, "proton-packages", "scripts")
}

func generateProtonPackageToolPath(workspace, tool string) string {
	return filepath.Join(generateProtonPackageToolsDirectoryPath(workspace), tool)
}

func CreateProtonPackageToolsDirectoryInWorkspace(workspace string) error {
	path := generateProtonPackageToolsDirectoryPath(workspace)
	slog.Info("Create static tools directory", "path", path)
	return os.MkdirAll(path, 0755)
}

func DownloadStaticTool(workspace, repoURL, arch, tool string) error {
	url := generateStaticToolURL(repoURL, arch, tool)
	path := generateProtonPackageToolPath(workspace, tool)

	slog.Info("Download static tool", "tool", tool)
	if err := utils.Download(url, path); err != nil {
		return err
	}

	if err := os.Chmod(path, 0755); err != nil {
		return err
	}

	return nil
}

func generateStaticToolURL(repo, arch, tool string) string {
	return fmt.Sprintf("%s/%s/%s/%s", repo, tool, arch, tool)
}
