package detectors

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	ecms "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1"
	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
)

func getMariaDBStatus(nodes map[string]string) (map[string]int, map[string]int, map[string]int, error) {
	var ctx = context.TODO()
	gcacheNum := make(map[string]int)
	galeraCrash := make(map[string]int)
	spaceUsage := make(map[string]int)

	for host, path := range nodes {
		var ecms = ecms.NewForHost(host)
		var f = ecms.Files()
		var executor = exec.NewECMSExecutorForHost(ecms.Exec())

		dir, err := f.ListDirectory(ctx, path)
		if err != nil {
			return gcacheNum, galeraCrash, spaceUsage, fmt.Errorf("node %s sftpClient readdir error: %v", host, err)
		}

		for _, file := range dir {
			if strings.Contains(file.Name(), "gcache.page.") {
				gcacheNum[host]++
			}
			if strings.Contains(file.Name(), "core.") {
				galeraCrash[host]++
			}
		}

		output, err := executor.Command("df", "-lh", path).Output()
		if err != nil {
			return gcacheNum, galeraCrash, spaceUsage, fmt.Errorf("execute command df -lh error: %v", err)
		}

		s := bufio.NewScanner(bytes.NewReader(output))
		for i := 0; s.Scan(); i++ {
			// skip first line for heading
			if i == 0 {
				continue
			}
			// 仅获取 Use% 部分
			fields := strings.Fields(s.Text())
			userPercentStr := strings.TrimSuffix(fields[len(fields)-2], "%")
			userPercent, _ := strconv.Atoi(userPercentStr)
			spaceUsage[host] = userPercent
			break
		}
	}

	return gcacheNum, galeraCrash, spaceUsage, nil
}
