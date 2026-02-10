package configuration

import corev1 "k8s.io/api/core/v1"

// Grafana 定义了 Grafana 的部署配置
type Grafana struct {
	Hosts            []string `json:"hosts,omitempty"`
	DataPath         string   `json:"data_path,omitempty"`
	StorageClassName string   `json:"storageClassName,omitempty"`
	StorageCapacity  string   `json:"storage_capacity,omitempty"`

	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}
