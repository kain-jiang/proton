package completion

import (
	"net"
	"net/url"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cs"
)

// CompletionCS 补全 CS 模块定义
func CompletionCS(c *configuration.Cs, nodes []configuration.Node) {
	// 判断是否为外置 Kubernetes
	if c.Provisioner == "" {
		if IsExternalKubernetes(clientcmd.RecommendedHomeFile) {
			c.Provisioner = configuration.KubernetesProvisionerExternal
		} else {
			c.Provisioner = configuration.KubernetesProvisionerLocal
		}
	}
	// Kubernetes 由本地（Proton）提供且未定义 IP Families 时补全 IP Families
	if c.Provisioner == configuration.KubernetesProvisionerLocal && c.IPFamilies == nil {
		// 具有 IPv4 协议地址的节点数量
		var ipv4_num int
		// 具有 IPv6 协议地址的节点数量
		var ipv6_num int
		for _, n := range nodes {
			if n.IP4 != "" {
				ipv4_num++
			}
			if n.IP6 != "" {
				ipv6_num++
			}
		}
		// 所有节点都有 IPv4 地址时, Kubernetes 使用 IPv4 地址
		if ipv4_num == len(nodes) {
			c.IPFamilies = append(c.IPFamilies, v1.IPv4Protocol)
		}
		// 所有节点都有 IPv6 地址时, Kubernetes 使用 IPv6 地址
		if ipv6_num == len(nodes) {
			c.IPFamilies = append(c.IPFamilies, v1.IPv6Protocol)
		}
	}
	if c.Addons == nil {
		c.Addons = configuration.DefaultCSAddons
	}

	// TODO: complete container runtime docker
}

// IsExternalKubernetes 返回 kubeconfig 是否包括外部 kubernetes
func IsExternalKubernetes(kubeconfigPath string) bool {
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return false
	}
	for _, cluster := range config.Clusters {
		u, err := url.Parse(cluster.Server)
		if err != nil {
			continue
		}
		h, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			continue
		}
		if h != cs.LocalKubernetesControlPlaneEndpointHost {
			return true
		}
	}
	return false
}
