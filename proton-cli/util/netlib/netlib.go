package netlib

import (
	"net"
	"strconv"

	"k8s.io/utils/exec"

	"github.com/c-robinson/iplib"
)

var execer exec.Interface = exec.New()

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func IsIPv4(ip string) bool {
	nodeIP := net.ParseIP(ip)
	if nodeIP.To4() != nil {
		return true
	} else {
		return false
	}
}

func GetnumMask(CIDR string) (string, error) {
	_, ipNet, err := net.ParseCIDR(CIDR)
	if err != nil {
		return "", err
	}
	mask := ipNet.Mask
	numMask, _ := mask.Size()

	return strconv.Itoa(numMask), nil
}

func GetAvailableIPList(CIDR string) ([]string, error) {
	var ipList []string
	ip, ipNet, err := net.ParseCIDR(CIDR)
	if err != nil {
		return ipList, err
	}
	mask := ipNet.Mask
	maskNum, _ := mask.Size()

	if IsIPv4(ip.String()) {
		net := iplib.NewNet4(ip, maskNum)
		for _, ip := range net.Enumerate(0, 1) {
			ipList = append(ipList, ip.String())
		}
	} else {
		//可用IP太大枚举会导致OOM，使用120还有255个可用IP，足够了
		net := iplib.NewNet6(ip, 120, 0)
		for _, ip := range net.Enumerate(0, 1) {
			ipList = append(ipList, ip.String())
		}
	}
	return ipList, nil
}

func NetworkAvaiable(ip string) bool {
	var pingCmd string
	if IsIPv4(ip) {
		pingCmd = "ping"
	} else {
		pingCmd = "ping6"
	}
	cmd := execer.Command(pingCmd, ip, "-c", "1", "-W", "1")
	err := cmd.Run()
	if err != nil {
		return false
	} else {
		return true
	}
}
