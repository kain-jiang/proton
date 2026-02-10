package components

import (
	"fmt"

	"component-manage/pkg/util"
)

type ComponentETCD struct {
	Name    string               `json:"name" yaml:"name"`
	Type    string               `json:"type" yaml:"type"`
	Version string               `json:"version" yaml:"version"`
	Params  *ETCDComponentParams `json:"params" yaml:"params"`
	Info    *ETCDComponentInfo   `json:"info" yaml:"info"`
}

type ETCDComponentParams struct {
	Namespace        string     `json:"namespace" yaml:"namespace"`
	ReplicaCount     int        `json:"replica_count,omitempty" yaml:"replica_count,omitempty"`
	Hosts            []string   `json:"hosts" yaml:"hosts"`
	Data_path        string     `json:"data_path" yaml:"data_path"`
	StorageClassName string     `json:"storageClassName,omitempty" yaml:"storageClassName,omitempty"`
	StorageCapacity  string     `json:"storage_capacity,omitempty" yaml:"storage_capacity,omitempty"`
	Resources        *Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
}

type ETCDComponentInfo struct {
	Namespace  string `json:"namespace,omitempty" mapstructure:"namespace" yaml:"namespace,omitempty"`
	SourceType string `json:"source_type,omitempty" mapstructure:"connectType" yaml:"source_type,omitempty"`
	Hosts      string `json:"hosts,omitempty" mapstructure:"host" yaml:"hosts,omitempty"`
	Port       int    `json:"port,omitempty" mapstructure:"port" yaml:"port,omitempty"`
	Secret     string `json:"secret,omitempty" mapstructure:"secret" yaml:"secret,omitempty"`
}

func (c *ComponentETCD) ToBase() *ComponentObject {
	return &ComponentObject{
		Name:    c.Name,
		Type:    c.Type,
		Version: c.Version,
		Params:  mustToMap(c.Params),
		Info:    mustToMap(c.Info),
	}
}

func (c *ComponentObject) TryToETCD() (*ComponentETCD, error) {
	if c.Type != "etcd" {
		return nil, fmt.Errorf("component type is not etcd")
	}

	params, err := util.FromMap[ETCDComponentParams](c.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert params to ETCDComponentParams: %w", err)
	}

	info, err := util.FromMap[ETCDComponentInfo](c.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to convert info to ETCDComponentInfo: %w", err)
	}

	return &ComponentETCD{
		Name:    c.Name,
		Type:    c.Type,
		Version: c.Version,
		Params:  params,
		Info:    info,
	}, nil
}

type ETCDCAInfo struct {
	Namespace      string `json:"namespace" yaml:"namespace"`
	CertSecretName string `json:"cert_secret_name" yaml:"cert_secret_name"`
	CertSecretKey  string `json:"cert_secret_key" yaml:"cert_secret_key"`
	KeySecretName  string `json:"key_secret_name" yaml:"key_secret_name"`
	KeySecretKey   string `json:"key_secret_key" yaml:"key_secret_key"`
}
