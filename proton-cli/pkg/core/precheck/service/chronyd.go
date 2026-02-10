package service

import (
	"os/exec"
	"strings"
)

func ChronydAlived() [][]string {
	svcInfo := [][]string{}
	// 执行命令检查chronyd服务状态
	cmd := exec.Command("systemctl", "is-active", "chronyd")
	output, err := cmd.Output()
	if err != nil {
		svcInfo = append(svcInfo, []string{"Service Chronyd", err.Error(), "\033[31mNO PASS\033[0m", "check chronyd installed or systemctl command broken"})
		return svcInfo
	}

	serviceStatus := strings.TrimSpace(string(output))

	// 判断chronyd服务状态
	if serviceStatus == "active" {
		svcInfo = append(svcInfo, []string{"Service Chronyd", "running", "\033[32mPASS\033[0m", ""})
	} else {
		svcInfo = append(svcInfo, []string{"Service Chronyd", "stopped", "\033[31mNO PASS\033[0m", "start chronyd service"})
	}
	return svcInfo
}
