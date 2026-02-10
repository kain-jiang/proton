package components

import "component-manage/pkg/util"

type ComponentObject struct {
	Name         string         `json:"name" yaml:"name"`
	Type         string         `json:"type" yaml:"type"`
	Version      string         `json:"version" yaml:"version"`
	Dependencies map[string]any `json:"dependencies" yaml:"dependencies"`
	Params       map[string]any `json:"params" yaml:"params"`
	Info         map[string]any `json:"info" yaml:"info"`
}

type Resources struct {
	Limits   ResourceRequirements `json:"limits,omitempty" yaml:"limits,omitempty"`
	Requests ResourceRequirements `json:"requests,omitempty" yaml:"requests,omitempty"`
}

type ResourceRequirements struct {
	CPU    string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory string `json:"memory,omitempty" yaml:"memory,omitempty"`
}

type canToMapTypes interface {
	Resources | *Resources |
		KafkaComponentDependencies | *KafkaComponentDependencies |
		KafkaComponentParams | *KafkaComponentParams |
		KafkaComponentInfo | *KafkaComponentInfo |
		ZookeeperComponentInfo | *ZookeeperComponentInfo |
		ZookeeperComponentParams | *ZookeeperComponentParams |
		OpensearchComponentInfo | *OpensearchComponentInfo |
		OpensearchComponentParams | *OpensearchComponentParams |
		RedisComponentInfo | *RedisComponentInfo |
		RedisComponentParams | *RedisComponentParams |
		ETCDComponentInfo | *ETCDComponentInfo |
		ETCDComponentParams | *ETCDComponentParams |
		MariaDBComponentParams | *MariaDBComponentParams |
		MariaDBComponentInfo | *MariaDBComponentInfo |
		PolicyEngineComponentParams | *PolicyEngineComponentParams |
		PolicyEngineComponentInfo | *PolicyEngineComponentInfo |
		PolicyEngineComponentDependencies | *PolicyEngineComponentDependencies |
		MongoDBComponentParams | *MongoDBComponentParams |
		MongoDBComponentInfo | *MongoDBComponentInfo |
		NebulaComponentParams | *NebulaComponentParams |
		NebulaComponentInfo | *NebulaComponentInfo |
		PrometheusComponentDependencies | *PrometheusComponentDependencies |
		PrometheusComponentParams | *PrometheusComponentParams |
		PrometheusComponentInfo | *PrometheusComponentInfo
}

func mustToMap[T canToMapTypes](o T) map[string]any {
	m, err := util.ToMap(o)
	if err != nil {
		// unreachable
		panic(err)
	}
	return m
}
