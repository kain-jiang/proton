package components

import (
	"fmt"

	"component-manage/pkg/util"
)

type ComponentKafka struct {
	Name         string                      `json:"name" yaml:"name"`
	Type         string                      `json:"type" yaml:"type"`
	Version      string                      `json:"version" yaml:"version"`
	Dependencies *KafkaComponentDependencies `json:"dependencies" yaml:"dependencies"`
	Params       *KafkaComponentParams       `json:"params" yaml:"params"`
	Info         *KafkaComponentInfo         `json:"info"  yaml:"info"`
}

///////////////////////////////////////////////////////////

type KafkaComponentDependencies struct {
	Zookeeper string `json:"zookeeper" yaml:"zookeeper"`
}

type KafkaComponentParams struct {
	Namespace         string            `json:"namespace" yaml:"namespace"`
	ReplicaCount      int               `json:"replica_count" yaml:"replica_count"`
	Hosts             []string          `json:"hosts,omitempty" yaml:"hosts"`
	DataPath          string            `json:"data_path,omitempty" yaml:"data_path"`
	Env               map[string]string `json:"env" yaml:"env"`
	Resources         Resources         `json:"resources" yaml:"resources"`
	ExporterResources *Resources        `json:"exporter_resources,omitempty" yaml:"exporter_resources"`
	StorageClassName  string            `json:"storageClassName,omitempty" yaml:"storageClassName"`
	StorageCapacity   string            `json:"storage_capacity,omitempty" yaml:"storage_capacity"`
	// Custom
	DisableExternalService bool             `json:"disable_external_service" yaml:"disable_external_service"`
	ExternalServiceList    []map[string]any `json:"external_service_list" yaml:"external_service_list"`
}

type KafkaComponentInfo struct {
	// 和 proton-cli 生成的保持一致
	MQHosts        string    `json:"mq_hosts" yaml:"mq_hosts"`
	MQLookupdHosts string    `json:"mq_lookupd_hosts,omitempty" yaml:"mq_lookupd_hosts"`
	MQLookupdPort  int       `json:"mq_lookupd_port,omitempty" yaml:"mq_lookupd_port"`
	MQPort         int       `json:"mq_port" yaml:"mq_port"`
	MQType         string    `json:"mq_type" yaml:"mq_type"`
	SourceType     string    `json:"source_type" yaml:"source_type"`
	Auth           KafkaAuth `json:"auth" yaml:"auth"`
}

type KafkaAuth struct {
	Mechanism string `json:"mechanism" yaml:"mechanism"`
	Username  string `json:"username" yaml:"username"`
	Password  string `json:"password" yaml:"password"`
}

func (r *Resources) ResourceMap() map[string]any {
	return mustToMap(r)
}

func (c *ComponentKafka) ToBase() *ComponentObject {
	return &ComponentObject{
		Name:         c.Name,
		Type:         c.Type,
		Version:      c.Version,
		Dependencies: mustToMap(c.Dependencies),
		Params:       mustToMap(c.Params),
		Info:         mustToMap(c.Info),
	}
}

func (o *ComponentObject) TryToKafka() (*ComponentKafka, error) {
	if o.Type != "kafka" {
		return nil, fmt.Errorf("component type is not kafka")
	}

	deps, err := util.FromMap[KafkaComponentDependencies](o.Dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dependencies: %w", err)
	}

	params, err := util.FromMap[KafkaComponentParams](o.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert params to KafkaComponentParams: %w", err)
	}

	info, err := util.FromMap[KafkaComponentInfo](o.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to convert info to KafkaComponentInfo: %w", err)
	}

	return &ComponentKafka{
		Name:         o.Name,
		Type:         o.Type,
		Version:      o.Version,
		Dependencies: deps,
		Params:       params,
		Info:         info,
	}, nil
}
