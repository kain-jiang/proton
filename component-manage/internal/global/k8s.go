package global

import (
	"sync"

	"component-manage/pkg/k8s"
)

var (
	K8sCli  k8s.Client
	k8sOnce sync.Once
)

func InitK8sCli() {
	k8sOnce.Do(func() {
		K8sCli = k8s.New()
	})
}
