package detectors

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
)

var (
	APIServerCrt = "/etc/kubernetes/pki/apiserver.crt"
)

func checkAPIServerCertExpiry() [][]string {
	results := [][]string{}
	certData, err := os.ReadFile(APIServerCrt)
	if err != nil {
		results = append(results, []string{"Kubernetes", "Certificate Expiration", fmt.Sprintf("error reading certificate file: %v", err), colorRedOutput("Error"), "execute on master node"})
		return results
	}
	block, _ := pem.Decode(certData)
	if block == nil || block.Type != "CERTIFICATE" {
		results = append(results, []string{"Kubernetes", "Certificate Expiration", "invalid PEM data", colorRedOutput("Error"), "execute on master node"})
		return results
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		results = append(results, []string{"Kubernetes", "Certificate Expiration", fmt.Sprintf("error parsing certificate: %v", err), colorRedOutput("Error"), "execute on master node"})
		return results
	}
	now := time.Now()
	daysRemaining := int(cert.NotAfter.Sub(now).Hours()) / 24
	if daysRemaining < 30 {
		results = append(results, []string{"Kubernetes", "Certificate Expiration", fmt.Sprintf("certificate will expire in %d days", daysRemaining), colorYellowOutput("Warn"), "should renew certificate quickly"})
	} else {
		results = append(results, []string{"Kubernetes", "Certificate Expiration", fmt.Sprintf("certificate will expire in %d days", daysRemaining), colorGreenOutput("OK"), ""})
	}

	return results
}
func checkClusterStatus(k kubernetes.Interface) [][]string {
	results := [][]string{}
	nodes, err := k.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		results = append(results, []string{"Kubernetes", "Nodes", err.Error(), colorRedOutput("Error"), "check cluster status"})
	} else {
		if n := countNotReadyNodes(nodes.Items); len(n) != 0 {
			results = append(results, []string{"Kubernetes", "Nodes", fmt.Sprintf("not ready nodes: %s", n), colorYellowOutput("Warn"), "check cluster status"})
		} else {
			results = append(results, []string{"Kubernetes", "Nodes", "all nodes are ready", colorGreenOutput("OK"), ""})
		}
	}

	// 检查kube-system命名空间下的Pod运行状态
	pods, err := k.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		results = append(results, []string{"Kubernetes", "Pod status", err.Error(), colorRedOutput("Error"), "check cluster status"})
	} else {
		if pods := countNotRunningPods(pods.Items); len(pods) != 0 {
			results = append(results, []string{"Kubernetes", "Pod status", fmt.Sprintf("not running pods on kube-system: %s", pods), colorYellowOutput("Warn"), "check cluster status"})
		} else {
			results = append(results, []string{"Kubernetes", "Pod status", "all pods are running", colorGreenOutput("OK"), ""})
		}
	}

	// check etcd-node
	masterNodes := isMasterNode(nodes.Items)
	offlineMember, memberDBSize, raftIndex, err := getETCDStatus(masterNodes)
	if err != nil {
		results = append(results, []string{"Kubernetes", "ETCD Member", "get etcd status error: " + err.Error(), colorRedOutput("Error"), "check etcd luster status"})
	} else {
		if len(offlineMember) != 0 {
			results = append(results, []string{"Kubernetes", "ETCD Member", fmt.Sprintf("offline member: %s", offlineMember), colorYellowOutput("Warn"), "check etcd luster status"})
		} else {
			results = append(results, []string{"Kubernetes", "ETCD Member", "all member are online", colorGreenOutput("OK"), ""})
		}

		outOfDbSize := []string{}
		dbSize := []string{}
		for k, v := range memberDBSize {
			if float64(v)/1024/1024/1024 > 1.8 {
				outOfDbSize = append(outOfDbSize, fmt.Sprintf("%s DB Size: %.2f GB is greater than 1.8GB\n", k, float64(v)/1024/1024/1024))
			} else {
				dbSize = append(dbSize, fmt.Sprintf("%s DB Size: %.2f GB", k, float64(v)/1024/1024/1024))
			}
		}
		if len(outOfDbSize) != 0 {
			results = append(results, []string{"Kubernetes", "ETCD DBSize", fmt.Sprintf("out of db size: %s", outOfDbSize), colorYellowOutput("Warn"), "compact revision and defrag quickly, the limit is 2G"})
		} else {
			if len(offlineMember) != 0 {
				results = append(results, []string{"Kubernetes", "ETCD DBSize", fmt.Sprintf("online member db size is normal\n%s", strings.Join(dbSize, "\n")), colorYellowOutput("Warn"), "check offline member db size"})
			} else {
				results = append(results, []string{"Kubernetes", "ETCD DBSize", fmt.Sprintf("all member db size is normal\n%s", strings.Join(dbSize, "\n")), colorGreenOutput("OK"), ""})
			}
		}

		var firstKey string
		var firstValue uint64
		var misMatchedIndexs []string

		for k, v := range raftIndex {
			firstKey, firstValue = k, v
			break
		}

		if firstKey == "" {
			results = append(results, []string{"Kubernetes", "ETCD RaftIndex", "raft index is empty", colorRedOutput("Error"), "check etcd cluster endpoint status"})
		} else {
			for k, v := range raftIndex {
				if v != firstValue {
					misMatchedIndexs = append(misMatchedIndexs, fmt.Sprintf("%s: %d != %d", k, v, firstValue))
				}
			}
			if len(misMatchedIndexs) != 0 {
				results = append(results, []string{"Kubernetes", "ETCD RaftIndex", fmt.Sprintf("raft index is mismatched: \n%s", strings.Join(misMatchedIndexs, "\n")), colorYellowOutput("Warn"), "each member index is not completely consistent. It is normal when data is written and can be observed continuously"})
			} else {
				if len(offlineMember) != 0 {
					results = append(results, []string{"Kubernetes", "ETCD RaftIndex", "online member raft index is matched, but some member is offline", colorYellowOutput("Warn"), "check offline member endpoint status"})
				} else {
					results = append(results, []string{"Kubernetes", "ETCD RaftIndex", "raft index is matched", colorGreenOutput("OK"), ""})
				}

			}
		}
	}

	return results
}

func countNotRunningPods(pods []corev1.Pod) []string {
	var notRunningPods []string
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodSucceeded {
			continue
		}
		if pod.Status.Phase != corev1.PodRunning {
			notRunningPods = append(notRunningPods, pod.Name)
			continue
		}

		ready := false
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				ready = true
				break
			}
		}
		if !ready {
			notRunningPods = append(notRunningPods, pod.Name)
		}
	}
	return notRunningPods
}

func countNotReadyNodes(nodes []corev1.Node) []string {
	var notReadyNodes []string
	for _, node := range nodes {
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && (condition.Status == corev1.ConditionFalse || condition.Status == corev1.ConditionUnknown) {
				notReadyNodes = append(notReadyNodes, node.GetName())
				break
			}
		}
	}
	return notReadyNodes
}

func isMasterNode(nodes []corev1.Node) []corev1.Node {
	masterNodes := []corev1.Node{}
	for _, node := range nodes {
		labels := node.GetLabels()
		if val, exists := labels["node-role.kubernetes.io/master"]; exists && val == "" {
			masterNodes = append(masterNodes, node)
		}
	}
	return masterNodes
}

func NewK8SDetector() [][]string {
	results := [][]string{}
	_, c := client.NewK8sClient()
	if c == nil {
		results = append(results, []string{"Kubernetes", "Get Client", client.ErrKubernetesClientSetNil.Error(), colorRedOutput("Error"), "execute on master node or check kubernetes cluster"})
		return results
	}

	results = append(results, checkAPIServerCertExpiry()...)
	results = append(results, checkClusterStatus(c)...)

	return results
}
