package k8s

import (
	"os"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	resolvFile    = "/etc/resolv.conf"
)

func SelfNameSpace() string {
	file, err := os.ReadFile(namespaceFile)
	if err != nil {
		return metav1.NamespaceDefault
	}
	return strings.TrimSpace(string(file))
}

func ClusterDomain() string {
	defaultName := "cluster.local"
	file, err := os.ReadFile(resolvFile)
	if err != nil {
		return defaultName
	}
	for _, line := range strings.Split(string(file), "\n") {
		if strings.HasPrefix(line, "search") {
			regex := regexp.MustCompile(`search\s+(.*)`)
			matches := regex.FindStringSubmatch(line)
			if len(matches) > 1 {
				domains := strings.Split(matches[1], " ")
				// 自动识别 cluster-domain
				for _, domain := range domains {
					if strings.Contains(domain, ".svc.") {
						// 提取 svc 后面的部分
						parts := strings.Split(domain, ".svc.")
						if len(parts) == 2 {
							return parts[1]
						}
					}
				}
			}
		}
	}
	return defaultName
}
