package tarball

import (
	"archive/tar"
	"bytes"
	_ "embed"
	"log/slog"
	"os"
	"path/filepath"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/pkg/utils"
)

const (
	suffix = ".tar"
)

//go:embed install_deps.sh
var install_deps_bytes []byte

// 把目录 workspace 打包成 tarball 输出到目录 release，名称为 name
func CreateTarball(workspace, release, name string) error {
	slog.Info("Create install_deps.sh")
	{
		path := filepath.Join(workspace, "proton-packages", "install_deps.sh")
		if err := utils.CreateFileFrom(path, bytes.NewReader(install_deps_bytes)); err != nil {
			return err
		}
		if err := os.Chmod(path, 0755); err != nil {
			return err
		}
	}

	{
		slog.Info("Create version file")
		path := filepath.Join(workspace, "proton-packages", "proton-version.txt")
		if err := utils.CreateFileFrom(path, bytes.NewReader([]byte(name+"\n"))); err != nil {
			return err
		}
	}

	path := filepath.Join(release, name+suffix)
	slog.Info("Create release tarball", "path", path)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	if err := tw.AddFS(os.DirFS(workspace)); err != nil {
		return err
	}

	return nil
}
