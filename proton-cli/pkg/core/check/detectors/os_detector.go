package detectors

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	ecms "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1"
	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
)

var (
	sysctlParas = map[string]string{
		"net.ipv4.ip_forward": "1",
	}
)

func checkNodeTime(nodes []corev1.Node) (map[string]int64, error) {
	results := map[string]int64{}
	for _, node := range nodes {
		executor := exec.NewECMSExecutorForHost(ecms.NewForHost(node.Status.Addresses[0].Address).Exec())
		// get remote node time by execute command date, return formeted date
		output, err := executor.Command("date", "+%s").Output()
		if err != nil {
			return results, fmt.Errorf("execute command on node %s error: %v", node.Status.Addresses[0].Address, err)
		} else {
			// parse remote node time
			remoteTime, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
			if err != nil {
				return results, fmt.Errorf("parse time on node %s error: %v", node.Status.Addresses[0].Address, err)
			}
			results[node.Status.Addresses[0].Address] = remoteTime

		}
	}
	return results, nil
}

func checkNodeSysctl(nodes []corev1.Node) (map[string]string, error) {
	results := map[string]string{}
	for _, node := range nodes {
		executor := exec.NewECMSExecutorForHost(ecms.NewForHost(node.Status.Addresses[0].Address).Exec())
		for k := range sysctlParas {
			output, err := executor.Command("sysctl", k).Output()
			if err != nil {
				return results, fmt.Errorf("execute command on node %s error: %v", node.Status.Addresses[0].Address, err)
			} else {
				results[node.Status.Addresses[0].Address] = string(output)

			}
		}

	}
	return results, nil
}

func checkDNSResolveConf(nodes []corev1.Node) (map[string]string, map[string]string, error) {
	results := map[string]string{}
	dnsServers := map[string]string{}
	for _, node := range nodes {
		executor := exec.NewECMSExecutorForHost(ecms.NewForHost(node.Status.Addresses[0].Address).Exec())

		output, err := executor.Command("cat", "/etc/resolv.conf").Output()
		if err != nil {
			return results, dnsServers, fmt.Errorf("execute command on node %s error: %v", node.Status.Addresses[0].Address, err)
		} else {
			dnsServers[node.Status.Addresses[0].Address] = ""
			// 逐行判断
			reader := bytes.NewReader(output)
			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "search") {
					results[node.Status.Addresses[0].Address] = line
				}

				if strings.HasPrefix(line, "nameserver") {
					dnsServers[node.Status.Addresses[0].Address] = line
				}
			}
			if err := scanner.Err(); err != nil {
				return results, dnsServers, err
			}
		}
	}
	return results, dnsServers, nil
}

func checkNodeOOM(nodes []corev1.Node) ([]string, error) {
	results := []string{}
	for _, node := range nodes {
		executor := exec.NewECMSExecutorForHost(ecms.NewForHost(node.Status.Addresses[0].Address).Exec())

		cmd := "grep 'oom-killer' /var/log/messages | awk '{print $6}' | sort | uniq -c | sort -nr"
		output, err := executor.Command("bash", "-c", cmd).Output()
		if err != nil {
			return results, fmt.Errorf("execute command on node %s error: %v", node.Status.Addresses[0].Address, err)
		} else {
			if output != nil {
				results = append(results, fmt.Sprintf("%s: %s", node.Status.Addresses[0].Address, string(output)))
			}
		}
	}
	return results, nil
}

func checkNodeRootSpaceUsage(nodes []corev1.Node) (map[string]int, error) {
	spaceUsage := make(map[string]int)
	for _, node := range nodes {
		executor := exec.NewECMSExecutorForHost(ecms.NewForHost(node.Status.Addresses[0].Address).Exec())

		output, err := executor.Command("df", "-lh", "/").Output()
		if err != nil {
			return spaceUsage, fmt.Errorf("execute command df -lh / error: %v", err)
		} else {
			// 仅获取 Use% 部分
			fields := strings.Fields(string(output))
			userPercentStr := strings.TrimSuffix(fields[len(fields)-2], "%")
			userPercent, _ := strconv.Atoi(userPercentStr)
			spaceUsage[node.Status.Addresses[0].Address] = userPercent
		}

	}
	return spaceUsage, nil
}

func NewOSDetector() [][]string {
	results := [][]string{}
	_, c := client.NewK8sClient()
	if c == nil {
		results = append(results, []string{"Kubernetes", "Get Client", client.ErrKubernetesClientSetNil.Error(), colorRedOutput("Error"), "execute on master node or check kubernetes cluster"})
		return results
	}
	nodes, err := c.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		results = append(results, []string{"Kubernetes", "NodesAvaiable", err.Error(), colorRedOutput("Error"), "execute on master node or check kubernetes cluster"})
		return results
	}

	timeResult, err := checkNodeTime(nodes.Items)
	if err != nil {
		results = append(results, []string{"OS", "NodeTime", err.Error(), colorRedOutput("Error"), "check nodes ssh trust config"})
	} else {
		var maxTimestamp, minTimestamp int64
		for _, ts := range timeResult {
			if maxTimestamp == 0 || ts > maxTimestamp {
				maxTimestamp = ts
			}
			if minTimestamp == 0 || ts < minTimestamp {
				minTimestamp = ts
			}
		}
		inconsistentServers := make([]string, 0)
		for addr, ts := range timeResult {
			if maxTimestamp-ts > 3 || ts-minTimestamp > 3 {
				t := time.Unix(ts, 0).UTC().Format(time.RFC1123Z)
				inconsistentServers = append(inconsistentServers, fmt.Sprintf("%s: %s", addr, t))
			}
		}
		if len(inconsistentServers) > 0 {
			results = append(results, []string{"OS", "NodeTime", strings.Join(inconsistentServers, "\n"),
				colorYellowOutput("Warn"),
				"check chronyd service and ntp sources",
			})
		} else {
			results = append(results, []string{"OS", "NodeTime", "all nodes time dirrence is in 3 seconds",
				colorGreenOutput("OK"),
				"",
			})
		}
	}

	sysctlResults, err := checkNodeSysctl(nodes.Items)
	if err != nil {
		results = append(results, []string{"OS", "Sysctl", err.Error(), colorRedOutput("Error"), "check sysctl on all nodes"})
	} else {
		disMatches := []string{}
		for k, v := range sysctlResults {
			sysctlK := strings.TrimSpace(strings.Split(v, "=")[0])
			sysctlV := strings.TrimSpace(strings.Split(v, "=")[1])
			if sysctlV != sysctlParas[sysctlK] {
				disMatches = append(disMatches, fmt.Sprintf("%s: %s = %s, should: %s = %s", k, sysctlK, sysctlV, sysctlK, sysctlParas[sysctlK]))
			}
		}
		if len(disMatches) != 0 {
			results = append(results, []string{"OS", "Sysctl", strings.Join(disMatches, "\n"), colorRedOutput("Error"), "sysctl -w and persistent setting"})
		} else {
			results = append(results, []string{"OS", "Sysctl", "all node ip_forward = 1 ", colorGreenOutput("OK"), ""})
		}
	}

	oomResults, err := checkNodeOOM(nodes.Items)
	if err != nil {
		results = append(results, []string{"OS", "OOM", err.Error(), colorRedOutput("Error"), "check oom on all nodes"})
	} else {
		if len(oomResults) != 0 {
			results = append(results, []string{"OS", "OOM", fmt.Sprintf("found oom history in /var/log/messages: \n%s", strings.Join(oomResults, "\n")), colorYellowOutput("Warn"), "maybe process limits memory is too small"})
		} else {
			results = append(results, []string{"OS", "OOM", "not found oom history in /var/log/messages", colorGreenOutput("OK"), ""})
		}
	}

	searchDNS, nameServers, err := checkDNSResolveConf(nodes.Items)
	if err != nil {
		results = append(results, []string{"OS", "DNS resolv.conf", err.Error(), colorRedOutput("Error"), "please check /etc/resolv.conf"})
	} else {
		searchDomains := []string{}
		for k, v := range searchDNS {
			if v != "" {
				searchDomains = append(searchDomains, fmt.Sprintf("%s: %s", k, v))
			}
		}
		if len(searchDomains) != 0 {
			results = append(results, []string{"OS", "DNS resolv.conf", fmt.Sprintf("search domain not allowed: \n%s", strings.Join(searchDomains, "\n")), colorRedOutput("Error"), "comment `search` keyword line in /etc/resolv.conf"})
		} else {
			results = append(results, []string{"OS", "DNS resolv.conf", "dns resolv.conf not contain search domain", colorGreenOutput("OK"), ""})
		}

		servers := []string{}
		for k, v := range nameServers {
			if v == "" {
				servers = append(servers, k)
			}
		}
		if len(servers) != 0 {
			results = append(results, []string{"OS", "DNS resolv.conf", fmt.Sprintf("not found nameserver in /etc/resolv.conf: \n%s", strings.Join(servers, "\n")), colorRedOutput("Error"), "set real nameserver in /etc/resolv.conf"})
		} else {
			results = append(results, []string{"OS", "DNS resolv.conf", "found nameserver in /etc/resolv.conf", colorGreenOutput("OK"), ""})
		}
	}

	rootSpaceUsage, err := checkNodeRootSpaceUsage(nodes.Items)
	if err != nil {
		results = append(results, []string{"OS", "/ Space Usage", err.Error(), colorRedOutput("Error"), "check root space usage"})
	} else {
		outOfSpace := []string{}
		for k, v := range rootSpaceUsage {
			if v > 80 {
				outOfSpace = append(outOfSpace, fmt.Sprintf("%s: %d%%", k, v))
			}
		}
		if len(outOfSpace) != 0 {
			results = append(results, []string{"OS", "/ Space Usage", fmt.Sprintf("these node / space used bigger than 80%%: \n%s", strings.Join(outOfSpace, "\n")), colorYellowOutput("Warn"), "extend mariadb data directory space"})
		} else {
			results = append(results, []string{"OS", "/ Space Usage", "all node / space used smaller than 80%%", colorGreenOutput("OK"), ""})
		}
	}

	return results
}
