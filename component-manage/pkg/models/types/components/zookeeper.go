package components

import (
	"fmt"

	"component-manage/pkg/util"
)

type ComponentZookeeper struct {
	Name    string                    `json:"name" yaml:"name"`
	Type    string                    `json:"type" yaml:"type"`
	Version string                    `json:"version" yaml:"version"`
	Params  *ZookeeperComponentParams `json:"params" yaml:"params"`
	Info    *ZookeeperComponentInfo   `json:"info" yaml:"info"`
}

type ZookeeperComponentParams struct {
	Namespace         string            `json:"namespace" yaml:"namespace"`
	ReplicaCount      int               `json:"replica_count" yaml:"replica_count"`
	Hosts             []string          `json:"hosts,omitempty" yaml:"hosts"`
	DataPath          string            `json:"data_path,omitempty" yaml:"data_path"`
	Env               map[string]string `json:"env" yaml:"env"`
	Resources         Resources         `json:"resources" yaml:"resources"`
	ExporterResources *Resources        `json:"exporter_resources,omitempty"`
	StorageClassName  string            `json:"storageClassName,omitempty" yaml:"storageClassName"`
	StorageCapacity   string            `json:"storage_capacity,omitempty" yaml:"storage_capacity"`
}

type ZookeeperComponentInfo struct {
	Host string        `json:"host" yaml:"host"`
	Port int           `json:"port" yaml:"port"`
	Sasl ZookeeperSASL `json:"sasl" yaml:"sasl"`
}

type ZookeeperSASL struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

func (c *ComponentZookeeper) ToBase() *ComponentObject {
	return &ComponentObject{
		Name:    c.Name,
		Type:    c.Type,
		Version: c.Version,
		Params:  mustToMap(c.Params),
		Info:    mustToMap(c.Info),
	}
}

func (c *ComponentObject) TryToZookeeper() (*ComponentZookeeper, error) {
	if c.Type != "zookeeper" {
		return nil, fmt.Errorf("component type is not zookeeper")
	}

	params, err := util.FromMap[ZookeeperComponentParams](c.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert params to ZookeeperComponentParams: %w", err)
	}

	info, err := util.FromMap[ZookeeperComponentInfo](c.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to convert info to ZookeeperComponentInfo: %w", err)
	}

	return &ComponentZookeeper{
		Name:    c.Name,
		Type:    c.Type,
		Version: c.Version,
		Params:  params,
		Info:    info,
	}, nil
}
