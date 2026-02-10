package tar

import (
	"os"
	"os/exec"
)

// 创建 tarball，target 为 tarball 文件路径，workDirectory 为工作路径，files 为 tarball 的内容路径
func CreateTarball(target string, workDirectory string, files ...string) error {
	args := []string{
		"--create",
		"--gzip",
		"--file",
		target,
		"--directory",
		workDirectory,
	}

	args = append(args, files...)

	cmd := exec.Command("tar", args...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

// 解压 tarball，target 为 tarball 文件路径，workDirectory 为解压目录
func DecompressTarball(target string, workDirectory string) error {
	args := []string{
		"--extract",
		"--gzip",
		"--file",
		target,
		"--directory",
		workDirectory,
	}
	cmd := exec.Command("tar", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
