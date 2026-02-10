package components

import (
	"fmt"

	"component-manage/pkg/util"
)

type ComponentOpensearch struct {
	Name    string                     `json:"name" yaml:"name"`
	Type    string                     `json:"type" yaml:"type"`
	Version string                     `json:"version" yaml:"version"`
	Params  *OpensearchComponentParams `json:"params" yaml:"params"`
	Info    *OpensearchComponentInfo   `json:"info"  yaml:"info"`
}

///////////////////////////////////////////////////////////

type OpensearchComponentParams struct {
	Namespace         string                 `json:"namespace" yaml:"namespace"`
	ExtraValues       map[string]interface{} `json:"extraValues,omitempty" yaml:"extraValues,omitempty"`
	ReplicaCount      int                    `json:"replica_count,omitempty" yaml:"replica_count,omitempty"`
	Hosts             []string               `json:"hosts" yaml:"hosts"`
	Data_path         string                 `json:"data_path" yaml:"data_path"`
	Mode              string                 `json:"mode" yaml:"mode"`
	Config            OpensearchConfigs      `json:"config" yaml:"config"`
	Settings          map[string]interface{} `json:"settings,omitempty" yaml:"settings,omitempty"`
	Resources         *Resources             `json:"resources,omitempty" yaml:"resources,omitempty"`
	ExporterResources *Resources             `json:"exporter_resources,omitempty" yaml:"exporter_resources,omitempty"`
	StorageClassName  string                 `json:"storageClassName,omitempty" yaml:"storageClassName,omitempty"`
	StorageCapacity   string                 `json:"storage_capacity,omitempty" yaml:"storage_capacity,omitempty"`
}

type OpensearchConfigs struct {
	JvmOptions              string `json:"jvmOptions" yaml:"jvmOptions"`
	HanlpRemoteextDict      string `json:"hanlpRemoteextDict" yaml:"hanlpRemoteextDict"`
	HanlpRemoteextStopwords string `json:"hanlpRemoteextStopwords" yaml:"hanlpRemoteextStopwords"`
}

type OpensearchComponentInfo struct {
	SourceType   string `json:"source_type" yaml:"source_type"`
	Hosts        string `json:"hosts" yaml:"hosts"`
	Port         int    `json:"port" yaml:"port"`
	Username     string `json:"username" yaml:"username"`
	Password     string `json:"password" yaml:"password"`
	Protocol     string `json:"protocol" yaml:"protocol"`
	Version      string `json:"version" yaml:"version"`
	Distribution string `json:"distribution" yaml:"distribution"`
}

func (c *ComponentOpensearch) ToBase() *ComponentObject {
	return &ComponentObject{
		Name:    c.Name,
		Type:    c.Type,
		Version: c.Version,
		Params:  mustToMap(c.Params),
		Info:    mustToMap(c.Info),
	}
}

func (o *ComponentObject) TryToOpensearch() (*ComponentOpensearch, error) {
	if o.Type != "opensearch" {
		return nil, fmt.Errorf("component type is not opensearch")
	}

	params, err := util.FromMap[OpensearchComponentParams](o.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert params to OpensearchComponentParams: %w", err)
	}

	info, err := util.FromMap[OpensearchComponentInfo](o.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to convert info to OpensearchComponentInfo: %w", err)
	}

	return &ComponentOpensearch{
		Name:    o.Name,
		Type:    o.Type,
		Version: o.Version,
		Params:  params,
		Info:    info,
	}, nil
}
