package configuration

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	// ZooKeeper 的 Chart 名称
	ChartNameZooKeeper = "proton-zookeeper"
	// ZooKeeper 的 Helm release 名称
	ReleaseNameZooKeeper = "zookeeper"
	// MaxHostNumberForZooKeeper 是 ZooKeeper 允许的最多的节点数，每个节点一个副本，也等于是副本数
	MaxHostNumberForZooKeeper = 3
)

// ZooKeeper
type ZooKeeper struct {
	ReplicaCount int `json:"replica_count,omitempty"`
	// ZooKeeper 所在的节点名称列表
	Hosts []string `json:"hosts"`
	// ZooKeeper 使用主机的此路径作为数据目录
	Data_path string `json:"data_path"`
	// ZooKeeper 的环境变量
	Env map[string]string `json:"env"`
	// ZooKeeper 所用的资源
	Resources *corev1.ResourceRequirements `json:"resources"`
	// 使用的 storage class 的名称，空字符串代表不使用 storage class
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageCapacity  string `json:"storage_capacity,omitempty"`
}
