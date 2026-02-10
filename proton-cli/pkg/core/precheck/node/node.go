package node

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

var validOS = []string{
	"red hat",
	"centos",
	"kylin",
	"suse",
	"openeuler",
	"uos",
	"ctyunos",
	"nfschina",
	"bigcloud",
}
var kylinKernelVersion = "4.19.90-23.35.v2101.ky10"

func CompareKernelVersions(version1, version2 string) int {
	ver1Parts := strings.Split(version1, ".")
	ver2Parts := strings.Split(version2, ".")

	for i := 0; ; i++ {
		part1, part2 := "", ""

		if i < len(ver1Parts) {
			part1 = ver1Parts[i]
		}
		if i < len(ver2Parts) {
			part2 = ver2Parts[i]
		}

		if part1 == "" {
			part1 = "0"
		}
		if part2 == "" {
			part2 = "0"
		}

		part1Num, _ := strconv.Atoi(part1)
		part2Num, _ := strconv.Atoi(part2)

		if part1Num > part2Num {
			return 1
		} else if part1Num < part2Num {
			return -1
		}

		if i+1 >= len(ver1Parts) && i+1 >= len(ver2Parts) {
			return 0
		}
	}
}

func getOSVersion() (string, error) {
	osrelease, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(osrelease), "\n") {
		parts := strings.SplitN(string(line), "=", 2)
		key := strings.Trim(parts[0], " \t\"'")
		if key == "PRETTY_NAME" {
			return strings.Trim(parts[1], " \t\"'"), nil
		}
	}
	return "", fmt.Errorf("cannot found PRETTY_NAME in /etc/os-release")
}

func getKernelVersion() (string, error) {
	output, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getMemoryCapacity() (uint64, error) {
	meminfo, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}

	var capacity uint64
	lines := strings.Split(string(meminfo), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == "MemTotal:" {
			capacity, _ = strconv.ParseUint(fields[1], 10, 64)
			break
		}
	}

	return capacity / 1024 / 1024, nil
}

func getCPUInfo() int {
	numCores := runtime.NumCPU()

	return numCores
}

func getDiskSpace() (uint64, error) {
	var statfs syscall.Statfs_t
	err := syscall.Statfs("/", &statfs)
	if err != nil {
		return 0, err
	}

	available := statfs.Bavail * uint64(statfs.Bsize)
	return available / 1024 / 1024 / 1024, nil
}

func getResolvSearch() (string, error) {
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return "", fmt.Errorf("failed to open /etc/resolv.conf: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 去除行首空格并判断是否以'#'开头，如果不是注释行则进行处理
		cleanLine := strings.TrimSpace(line)
		if !strings.HasPrefix(cleanLine, "#") {
			// 判断行中是否存在 'search' 关键字
			if strings.Contains(cleanLine, "search") {
				// 打印含有search关键字且未被注释的行
				return cleanLine, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error while reading file: %v", err)
	}
	return "", nil
}

func getHostLocalDomain() (map[string]bool, error) {
	file, err := os.Open("/etc/hosts")
	if err != nil {
		return nil, fmt.Errorf("failed to open /etc/hosts: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	found := make(map[string]bool)
	for scanner.Scan() {
		line := scanner.Text()
		// 去除行首空格并判断是否以'#'开头，如果不是注释行则进行处理
		cleanLine := strings.TrimSpace(line)
		if !strings.HasPrefix(cleanLine, "#") {
			if strings.Contains(line, "127.0.0.1") && (strings.Contains(line, "localhost") || strings.Contains(line, "localhost.localdomain")) {
				found["127.0.0.1"] = true
			} else if strings.Contains(line, "::1") && (strings.Contains(line, "localhost") || strings.Contains(line, "localhost.localdomain")) {
				found["::1"] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error while reading file: %v", err)
	}
	return found, nil
}

func getSELinuxStatus() (map[string]string, error) {
	var result = map[string]string{
		"getenforce": "",
		"config":     "",
	}
	cmd := exec.Command("getenforce")
	output, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			result["getenforce"] = "skip"
			return result, err
		} else {
			return nil, err
		}
	} else {
		status := strings.TrimSpace(string(output))
		result["getenforce"] = status

	}

	content, err := os.ReadFile("/etc/selinux/config")
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			result["config"] = "skip"
			return result, err
		} else {
			return nil, err
		}
	} else {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "SELINUX=") {
				value := strings.TrimPrefix(line, "SELINUX=")
				result["config"] = value
			}
		}
	}

	return result, nil
}

func getSwap() (bool, error) {
	file, err := os.Open("/etc/fstab")
	if err != nil {
		return false, fmt.Errorf("failed to open /etc/fstab: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 去除行首空格并判断是否以'#'开头，如果不是注释行则进行处理
		cleanLine := strings.TrimSpace(line)
		if !strings.HasPrefix(cleanLine, "#") {
			if strings.Contains(line, "swap") {
				return true, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error while reading file: %v", err)
	}
	return false, nil
}

func NodeInfo() [][]string {
	nodeInfo := [][]string{}
	osVersion, err := getOSVersion()
	if err != nil {
		nodeInfo = append(nodeInfo, []string{"OS Version", err.Error(), "\033[31mNO PASS\033[0m", "check /etc/os-release"})
	} else {
		flag := false
		for _, os := range validOS {
			if strings.Contains(strings.ToLower(osVersion), os) {
				flag = true
			}
		}
		if flag {
			nodeInfo = append(nodeInfo, []string{"OS Version", osVersion, "\033[32mPASS\033[0m", ""})
		} else {
			nodeInfo = append(nodeInfo, []string{"OS Version", osVersion, "\033[31mNO PASS\033[0m", "replace supported os vendor"})
		}

	}

	kernelVersion, err := getKernelVersion()
	if err != nil {
		nodeInfo = append(nodeInfo, []string{"Kernel Version", err.Error(), "\033[31mNO PASS\033[0m", "check uname -r"})
	} else {
		if strings.Contains(strings.ToLower(osVersion), "kylin") {
			compareResult := CompareKernelVersions(kylinKernelVersion, kernelVersion)
			switch compareResult {
			case -1:
				nodeInfo = append(nodeInfo, []string{"Kernel Version", kernelVersion, "\033[31mNO PASS\033[0m", fmt.Sprintf("kylin kernel versio must > %s", kylinKernelVersion)})
			case 0:
				nodeInfo = append(nodeInfo, []string{"Kernel Version", kernelVersion, "\033[32mPASS\033[0m", ""})
			case 1:
				nodeInfo = append(nodeInfo, []string{"Kernel Version", kernelVersion, "\033[32mPASS\033[0m", ""})
			}
		} else {
			nodeInfo = append(nodeInfo, []string{"Kernel Version", kernelVersion, "\033[32mPASS\033[0m", ""})
		}
	}

	memCapacity, err := getMemoryCapacity()
	if err != nil {
		nodeInfo = append(nodeInfo, []string{"Memory Capacity", err.Error(), "\033[31mNO PASS\033[0m", "check /etc/meminfo"})
	} else {
		if memCapacity < 64 {
			nodeInfo = append(nodeInfo, []string{"Memory Capacity", strconv.FormatUint(memCapacity, 10) + " GB", "\033[33mWARN\033[0m", "memory capacity should >= 64 GB"})
		} else {
			nodeInfo = append(nodeInfo, []string{"Memory Capacity", strconv.FormatUint(memCapacity, 10) + " GB", "\033[32mPASS\033[0m", ""})
		}
	}

	numCores := getCPUInfo()
	if numCores < 32 {
		nodeInfo = append(nodeInfo, []string{"CPU Cores", strconv.Itoa(numCores), "\033[33mWARN\033[0m", "CPU cores should >= 32"})
	} else {
		nodeInfo = append(nodeInfo, []string{"CPU Cores", strconv.Itoa(numCores), "\033[32mPASS\033[0m", ""})
	}

	diskSpace, err := getDiskSpace()
	if err != nil {
		nodeInfo = append(nodeInfo, []string{"Disk / Free Size", err.Error(), "\033[31mNO PASS\033[0m", "check / disk space"})
	} else {
		if diskSpace < 300 {
			nodeInfo = append(nodeInfo, []string{"Disk / Free Size", strconv.FormatUint(diskSpace, 10) + " GB", "\033[33mWARN\033[0m", "Disk / free size should >= 300 GB"})
		} else {
			nodeInfo = append(nodeInfo, []string{"Disk / Free Size", strconv.FormatUint(diskSpace, 10) + " GB", "\033[32mPASS\033[0m", ""})
		}
	}

	searchStr, err := getResolvSearch()
	if err != nil {
		nodeInfo = append(nodeInfo, []string{"DNS Search Domain", err.Error(), "\033[31mNO PASS\033[0m", "check /etc/resolv.conf"})
	}
	if searchStr != "" {
		nodeInfo = append(nodeInfo, []string{"DNS Search Domain", searchStr, "\033[31mNO PASS\033[0m", "remove search domain from /etc/resolv.conf"})
	} else {
		nodeInfo = append(nodeInfo, []string{"DNS Search Domain", "", "\033[32mPASS\033[0m", ""})
	}

	localDomain, err := getHostLocalDomain()
	if err != nil {
		nodeInfo = append(nodeInfo, []string{"HOST LocalDomain Map", err.Error(), "\033[31mNO PASS\033[0m", "check /etc/hosts file include 127.0.0.1 map"})
	}
	flag := true
	if !localDomain["127.0.0.1"] {
		nodeInfo = append(nodeInfo, []string{"HOST LocalDomain Map", "", "\033[31mNO PASS\033[0m", "fill [127.0.0.1   localhost localhost.localdomain localhost4 localhost4.localdomain4] into /etc/hosts"})
		flag = false
	}
	if !localDomain["::1"] {
		nodeInfo = append(nodeInfo, []string{"HOST LocalDomain Map", "", "\033[31mNO PASS\033[0m", "fill [::1         localhost localhost.localdomain localhost6 localhost6.localdomain6] into /etc/hosts"})
		flag = false
	}
	if flag {
		nodeInfo = append(nodeInfo, []string{"HOST LocalDomain Map", "found 127.0.0.1 and ::1", "\033[32mPASS\033[0m", ""})
	}

	selinuxStatus, err := getSELinuxStatus()
	if err != nil {
		if selinuxStatus["getenforce"] == "skip" || selinuxStatus["config"] == "skip" {
			nodeInfo = append(nodeInfo, []string{"SELinux", err.Error(), "\033[32mPASS\033[0m", "skip check SELinux like suse system"})
		} else {
			nodeInfo = append(nodeInfo, []string{"SELinux Disable", err.Error(), "\033[31mNO PASS\033[0m", "check getenforce or /etc/selinux/config file"})
		}
	} else {
		if selinuxStatus["getenforce"] == "Enforcing" {
			nodeInfo = append(nodeInfo, []string{"SELinux Current Disable", selinuxStatus["getenforce"], "\033[31mNO PASS\033[0m", "execute setenforce to disable selinux"})
		} else {
			nodeInfo = append(nodeInfo, []string{"SELinux Current Disable", "", "\033[32mPASS\033[0m", ""})
		}
		if selinuxStatus["config"] == "enforcing" {
			nodeInfo = append(nodeInfo, []string{"SELinux Config Disable", selinuxStatus["config"], "\033[31mNO PASS\033[0m", "change /etc/selinux/config SELINUX key to disabled or permissive"})
		} else {
			nodeInfo = append(nodeInfo, []string{"SELinux Config Disable", "", "\033[32mPASS\033[0m", ""})
		}
	}

	status, err := getSwap()
	if err != nil {
		nodeInfo = append(nodeInfo, []string{"Swap Memory Disable", err.Error(), "\033[31mNO PASS\033[0m", "manual check swap memory status"})
	} else {
		if status {
			nodeInfo = append(nodeInfo, []string{"Swap Memory Disable", "", "\033[31mNO PASS\033[0m", "disable swap memory current and persistent in /etc/fstab"})
		} else {
			nodeInfo = append(nodeInfo, []string{"Swap Memory Disable", "disabled", "\033[32mPASS\033[0m", ""})
		}
	}

	return nodeInfo
}
