package cs

import (
	"fmt"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	utilsclock "k8s.io/utils/clock"
)

const (
	// KubernetesDir is the directory Kubernetes owns for storing various configuration files
	KubernetesDir = "/etc/kubernetes"

	// AdminKubeConfigFileName defines name for the kubeconfig aimed to be used by the superuser/admin of the cluster
	AdminKubeConfigFileName = "admin.conf"

	// LabelNodeRoleOldControlPlane specifies that a node hosts control-plane components
	// DEPRECATED: https://github.com/kubernetes/kubeadm/issues/2200
	LabelNodeRoleOldControlPlane = "node-role.kubernetes.io/master"

	// LabelNodeRoleControlPlane specifies that a node hosts control-plane components
	LabelNodeRoleControlPlane = "node-role.kubernetes.io/control-plane"
)

const (
	// Proton CLI 部署的 Kubernetes 的 control plane 的 host
	LocalKubernetesControlPlaneEndpointHost = "proton-cs.lb.aishu.cn"
)

// NewKubernetesClient 创建 Kubernetes 客户端
func NewKubernetesClient() (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(KubernetesDir, AdminKubeConfigFileName))
	if err != nil {
		return nil, fmt.Errorf("unable to build rest config for kubernetes client: %w", err)
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create kubernetes client: %w", err)
	}

	return c, nil
}

// filterControlPlaneNode 从 corev1.NodeList 过滤出第一个 control plane 节点。如
// 果未找到则返回 nil
func filterControlPlaneNode(list *corev1.NodeList) *corev1.Node {
	for _, n := range list.Items {
		if _, ok := n.Labels[LabelNodeRoleControlPlane]; ok {
			return n.DeepCopy()
		}
		if _, ok := n.Labels[LabelNodeRoleOldControlPlane]; ok {
			return n.DeepCopy()
		}
	}
	return nil
}

// isKubernetesReady 返回 Kubernetes API 是否可用，重试 8 次，间隔 1 秒
func IsKubernetesAPIReady(c discovery.DiscoveryInterface, clock utilsclock.Clock) bool {
	for i := 0; i < 8; i++ {
		if i != 0 {
			clock.Sleep(time.Second)
		}

		if _, err := c.ServerVersion(); err == nil {
			return true
		}
	}
	return false
}
