package configuration

import v1 "k8s.io/api/core/v1"

// Nebula Graph 的部署配置
type Nebula struct {
	// 运行 nebula graph 的节点的名称
	Hosts []string `json:"hosts,omitempty"`
	// 数据目录的路径，需要是绝对路径
	DataPath string `json:"data_path,omitempty"`
	// root 账户的密码
	Password string `json:"password,omitempty"`
	// 组件 graphd 的配置
	Graphd NebulaComponent `json:"graphd,omitempty"`
	// 组件 metad 的配置
	Metad NebulaComponent `json:"metad,omitempty"`
	// 组件 storaged 的配置
	Storaged NebulaComponent `json:"storaged,omitempty"`
}

// Nebula 组件的配置
type NebulaComponent struct {
	// 组件需要的资源配额
	Resource *v1.ResourceRequirements `json:"resource,omitempty"`
}
