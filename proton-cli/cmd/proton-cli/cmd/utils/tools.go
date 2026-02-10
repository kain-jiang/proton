package utils

import (
	"fmt"
	"net"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func NodeListFromIPList(ips []string) ([]configuration.Node, error) {
	var nodeList []configuration.Node
	for i := range ips {
		ip := net.ParseIP(ips[i])
		if ip == nil {
			return nil, fmt.Errorf("invalid ip %q", ips[i])
		}
		n := configuration.Node{Name: ips[i]}
		if ip.To4() != nil {
			n.IP4 = ips[i]
		} else {
			n.IP6 = ips[i]
		}
		nodeList = append(nodeList, n)
	}
	return nodeList, nil
}
