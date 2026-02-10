package components

import (
	"fmt"

	"component-manage/pkg/util"
)

type MongoDBComponentParams struct {
	Namespace    string   `json:"namespace" yaml:"namespace"`
	ReplicaCount int      `json:"replica_count,omitempty" yaml:"replica_count,omitempty"`
	Hosts        []string `json:"hosts" yaml:"hosts"`
	Admin_user   string   `json:"admin_user" yaml:"admin_user"`
	Admin_passwd string   `json:"admin_passwd" yaml:"admin_passwd"`
	Data_path    string   `json:"data_path" yaml:"data_path"`

	StorageClassName string `json:"storageClassName,omitempty" yaml:"storageClassName,omitempty"`
	StorageCapacity  string `json:"storage_capacity,omitempty" yaml:"storage_capacity,omitempty"`

	Resources *Resources `json:"resources,omitempty" yaml:"resources,omitempty"`

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// 新增配置
	Username        string `json:"username" yaml:"username" binding:"required"`
	Password        string `json:"password" yaml:"password" binding:"required"`
	AdminSecretName string `json:"admin_secret_name" yaml:"admin_secret_name"`
}
type MongoDBComponentInfo struct {
	SourceType string      `json:"source_type,omitempty" yaml:"source_type,omitempty"`
	Hosts      string      `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Port       int         `json:"port,omitempty" yaml:"port,omitempty"`
	ReplicaSet string      `json:"replica_set,omitempty" yaml:"replica_set,omitempty"`
	Username   string      `json:"username,omitempty" yaml:"username,omitempty"`
	Password   string      `json:"password,omitempty" yaml:"password,omitempty"`
	SSL        bool        `json:"ssl" yaml:"ssl"`
	AuthSource string      `json:"auth_source,omitempty" yaml:"auth_source,omitempty"`
	Options    interface{} `json:"options,omitempty" yaml:"options,omitempty"`
	// MgmtInfo
	MgmtHost string `json:"mgmt_host,omitempty" yaml:"mgmt_host,omitempty"`
	MgmtPort int    `json:"mgmt_port,omitempty" yaml:"mgmt_port,omitempty"`
	AdminKey string `json:"admin_key,omitempty" yaml:"admin_key,omitempty"`
}

type ComponentMongoDB struct {
	Name    string                  `json:"name" yaml:"name"`
	Type    string                  `json:"type" yaml:"type"`
	Version string                  `json:"version" yaml:"version"`
	Params  *MongoDBComponentParams `json:"params" yaml:"params"`
	Info    *MongoDBComponentInfo   `json:"info" yaml:"info"`
}

func (c *ComponentMongoDB) ToBase() *ComponentObject {
	return &ComponentObject{
		Name:    c.Name,
		Type:    c.Type,
		Version: c.Version,
		Params:  mustToMap(c.Params),
		Info:    mustToMap(c.Info),
	}
}

func (o *ComponentObject) TryToMongoDB() (*ComponentMongoDB, error) {
	if o.Type != "mongodb" {
		return nil, fmt.Errorf("component type is not kafka")
	}

	params, err := util.FromMap[MongoDBComponentParams](o.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert params to MariaDBComponentParams: %w", err)
	}

	info, err := util.FromMap[MongoDBComponentInfo](o.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to convert info to MariaDBComponentInfo: %w", err)
	}

	return &ComponentMongoDB{
		Name:    o.Name,
		Type:    o.Type,
		Version: o.Version,
		Params:  params,
		Info:    info,
	}, nil
}
