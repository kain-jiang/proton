package firewall

import (
	"net"
	"net/netip"
	"strings"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	node "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func New(c *configuration.Firewall, nodes []node.Interface, podNetworkCIDR string, logger *logrus.Logger) Interface {
	switch c.Mode {
	case configuration.FirewallUserManaged:
		return &moduleUserManaged{
			logger: logger,
		}
	case configuration.FirewallFirewalld:
		// 节点 IP、节点内部 IP 列表
		var ips []net.IP = lo.FlatMap(nodes, func(n node.Interface, _ int) []net.IP { return []net.IP{n.IP(), n.InternalIP()} })
		// 节点 IP、节点内部 IP 列表 netip.Addr
		var addresses []netip.Addr
		addresses = lo.FilterMap(ips, func(ip net.IP, _ int) (netip.Addr, bool) { return netip.AddrFromSlice(ip) })
		// ::ffff:10.4.71.191 -> 10.4.71.191
		addresses = lo.Map(addresses, convert4In6To4)
		// 去重
		addresses = lo.Uniq(addresses)
		// podNetworkCIDR 可能同时包括 IPv4 和 IPv6 地址，使用逗号分隔
		nets := strings.Split(podNetworkCIDR, ",")

		return &moduleFirewalld{
			nodes:     nodes,
			addresses: addresses,
			nets:      nets,
			logger:    logger,
		}

	default:
		return &moduleUnsupportedMode{
			mode: c.Mode,
		}
	}
}

// Convert IPv4 in IPv6 to IPv6, others unchanged
func convert4In6To4(ip netip.Addr, _ int) netip.Addr {
	if !ip.Is4In6() {
		return ip
	}
	return netip.AddrFrom4(ip.As4())
}
