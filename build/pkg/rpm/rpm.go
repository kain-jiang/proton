package rpm

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/pkg/utils"
)

const (
	reposName string = "repos"
)

func DownloadProtonPackageArchive(workspace, repo, version, arch string) error {
	repository := filepath.Join(workspace, "proton-packages", reposName)

	slog.Info("Create repository directory", "path", repository)
	if err := os.MkdirAll(repository, 0755); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/proton-package.%s.%s.tar", repo, version, arch)

	slog.Info("Download proton package repository archive", "url", url)
	if err := utils.ExtractTarballURL(url, repository); err != nil {
		return err
	}

	return nil
}
