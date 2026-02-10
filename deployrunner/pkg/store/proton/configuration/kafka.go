package configuration

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	// kafka 的 Chart 名称
	ChartNameKafka = "proton-kafka"
	// kafka 的 Helm release 名称
	ReleaseNameKafka = "kafka"
)

// Kafka
type Kafka struct {
	ReplicaCount int `json:"replica_count,omitempty"`
	// Kafka 所在的节点名称列表
	Hosts []string `json:"hosts"`
	// Kafka 使用主机的此路径作为数据目录
	Data_path string `json:"data_path"`
	// Kafka 的环境变量
	Env map[string]string `json:"env"`
	// Kafka 所用的资源
	Resources         corev1.ResourceRequirements  `json:"resources"`
	ExporterResources *corev1.ResourceRequirements `json:"exporter_resources,omitempty"`
	// 使用的 storage class 的名称，空字符串代表不使用 storage class
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageCapacity  string `json:"storage_capacity,omitempty"`
}
