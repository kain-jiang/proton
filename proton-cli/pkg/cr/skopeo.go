package cr

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/utils/exec"
)

// SkopeoCopyOptions 是命令 `skopeo sync` 的参数。
type SkopeoCopyOptions struct {
	// run the tool without any policy check
	InsecurePolicy bool
	// Weather disable requiring HTTPS and verify certificates when talking to the container registry or daemon
	DisableDestinationTLSVerify bool
	// the number of times to possibly retry
	RetryTimes int
}

// RunSkopeoCopy 调用 `skopeo copy` push/pull 镜像
func RunSkopeoCopy(execer exec.Interface, source, destination string, opts SkopeoCopyOptions) error {
	// 构建命令参数
	cmdArgs := []string{"copy"}
	// `--dest-tls-verify 的默认值是 true`
	if opts.DisableDestinationTLSVerify {
		cmdArgs = append(cmdArgs, "--dest-tls-verify=false")
	}
	if opts.InsecurePolicy {
		cmdArgs = append(cmdArgs, "--insecure-policy")
	}
	if opts.RetryTimes > 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--retry-times=%d", opts.RetryTimes))
	}
	// command arguments
	cmdArgs = append(cmdArgs, source, destination)

	// 首先尝试从系统路径执行 skopeo
	err := execer.Command("skopeo", cmdArgs...).Run()
	if err == nil {
		return nil
	}

	// 如果系统路径上的 skopeo 执行失败，尝试从当前路径执行
	err = execer.Command("./skopeo", cmdArgs...).Run()
	if err != nil {
		return fmt.Errorf("cannot execute skopeo command: %v", err)
	}
	return nil
}

// GetSkopeoVersion 调用 `skopeo --version` 返回 skopeo 的版本
func GetSkopeoVersion(execer exec.Interface) (string, error) {
	// 首先尝试从系统路径获取 skopeo 版本
	out, err := execer.Command("skopeo", "--version").Output()
	if err == nil {
		return parseSkopeoVersion(out)
	}

	// 如果系统路径上的 skopeo 执行失败，尝试从当前路径获取版本
	out, err = execer.Command("./skopeo", "--version").Output()
	if err != nil {
		return "", errors.New("cannot execute 'skopeo --version', tried both system path and current directory")
	}

	return parseSkopeoVersion(out)
}

// parseSkopeoVersion 解析 skopeo --version 的输出
func parseSkopeoVersion(out []byte) (string, error) {
	s := strings.Split(string(out), " ")
	if len(s) < 2 || len(s[1]) == 0 {
		return "", fmt.Errorf("unable to parse output from skopeo: %q", string(out))
	}

	return s[1], nil
}
