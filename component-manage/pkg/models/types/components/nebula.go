package components

import (
	"fmt"

	"component-manage/pkg/util"
)

type NebulaComponentParams struct {
	Namespace string          `json:"namespace" yaml:"namespace"`
	Hosts     []string        `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	DataPath  string          `json:"data_path" yaml:"data_path" binding:"required"`
	Password  string          `json:"password,omitempty" yaml:"password,omitempty"`
	Graphd    NebulaContainer `json:"graphd,omitempty" yaml:"graphd,omitempty"`
	Metad     NebulaContainer `json:"metad,omitempty" yaml:"metad,omitempty"`
	Storaged  NebulaContainer `json:"storaged,omitempty" yaml:"storaged,omitempty"`
	// 不支持存储类
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// 新增配置
	// Username        string `json:"username" yaml:"username" binding:"required"`
	// Password        string `json:"password" yaml:"password" binding:"required"`
	AdminSecretName string `json:"admin_secret_name" yaml:"admin_secret_name"`
}

type NebulaContainer struct {
	Resource *Resources     `json:"resource,omitempty"`
	Config   map[string]any `json:"config,omitempty"`
}

type NebulaComponentInfo struct {
	Host     string `json:"host,omitempty" yaml:"host,omitempty"`
	Port     int    `json:"port,omitempty" yaml:"port,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

type ComponentNebula struct {
	Name    string                 `json:"name" yaml:"name"`
	Type    string                 `json:"type" yaml:"type"`
	Version string                 `json:"version" yaml:"version"`
	Params  *NebulaComponentParams `json:"params" yaml:"params"`
	Info    *NebulaComponentInfo   `json:"info" yaml:"info"`
}

func (c *ComponentNebula) ToBase() *ComponentObject {
	return &ComponentObject{
		Name:    c.Name,
		Type:    c.Type,
		Version: c.Version,
		Params:  mustToMap(c.Params),
		Info:    mustToMap(c.Info),
	}
}

func (c *ComponentObject) TryToNebula() (*ComponentNebula, error) {
	if c.Type != "nebula" {
		return nil, fmt.Errorf("component type is not nebula")
	}

	params, err := util.FromMap[NebulaComponentParams](c.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert params to NebulaComponentParams: %w", err)
	}

	info, err := util.FromMap[NebulaComponentInfo](c.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to convert info to NebulaComponentInfo: %w", err)
	}

	return &ComponentNebula{
		Name:    c.Name,
		Type:    c.Type,
		Version: c.Version,
		Params:  params,
		Info:    info,
	}, nil
}
