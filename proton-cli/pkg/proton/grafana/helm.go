package grafana

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/grafana/helm"
)

// valuesFor returns helm values of the grafana configuration
func valuesFor(spec *configuration.Grafana, registry, namespace string, prometheus *corev1.Service) *helm.Values {
	return &helm.Values{
		Namespace:    namespace,
		Image:        helm.ValuesImage{Registry: registry},
		ReplicaCount: Replicas,
		Service:      helm.ValuesService{EnableDualStack: global.EnableDualStack, Grafana: helm.ValuesGrafanaService{Type: corev1.ServiceTypeNodePort, NodePort: NodePort}},
		Config:       helm.ValuesConfig{DataSource: helm.ValuesDataSource{Prometheus: valuesPrometheusFor(prometheus, namespace)}},
		Storage:      valuesStorageFor(spec.Hosts, spec.DataPath, spec.StorageClassName, spec.StorageCapacity),
		Resources:    spec.Resources,
	}
}

// valuesPrometheusFor returns helm values' prometheus config from the
// kubernetes service of prometheus. If the service has multi ports, config use
// the first port.
func valuesPrometheusFor(prometheus *corev1.Service, namespace string) helm.ValuesPrometheus {
	var host string
	if prometheus.Namespace == namespace {
		host = prometheus.Name
	} else {
		host = prometheus.Name + "." + prometheus.Namespace
	}
	var port int
	for _, p := range prometheus.Spec.Ports {
		port = int(p.Port)
		break
	}
	return helm.ValuesPrometheus{
		Enabled:  true,
		Protocol: helm.ValuesProtocolHTTP,
		Host:     host,
		Port:     port,
	}
}

func valuesStorageFor(hosts []string, dataPath string, storageClassName string, storageCapacity string) helm.ValuesStorage {
	var storage helm.ValuesStorage
	if len(hosts) != 0 {
		storage.Local = make(map[string]helm.ValuesLocal)
	}
	for i, h := range hosts {
		storage.Local[strconv.Itoa(i)] = helm.ValuesLocal{Host: h, Path: dataPath}
	}
	storage.StorageClassName = storageClassName
	if len(storageCapacity) > 0 {
		storage.Capacity = storageCapacity
	}
	return storage
}
