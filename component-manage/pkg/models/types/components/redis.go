package components

import (
	"fmt"

	"component-manage/pkg/util"
)

type ComponentRedis struct {
	Name    string                `json:"name" yaml:"name"`
	Type    string                `json:"type" yaml:"type"`
	Version string                `json:"version" yaml:"version"`
	Params  *RedisComponentParams `json:"params" yaml:"params"`
	Info    *RedisComponentInfo   `json:"info" yaml:"info"`
}

type RedisComponentParams struct {
	Namespace        string     `json:"namespace" yaml:"namespace"`
	ReplicaCount     int        `json:"replica_count,omitempty" yaml:"replica_count,omitempty"`
	Hosts            []string   `json:"hosts" yaml:"hosts"`
	Admin_user       string     `json:"admin_user" yaml:"admin_user"`
	Admin_passwd     string     `json:"admin_passwd" yaml:"admin_passwd"`
	Data_path        string     `json:"data_path" yaml:"data_path"`
	StorageClassName string     `json:"storageClassName,omitempty" yaml:"storageClassName,omitempty"`
	StorageCapacity  string     `json:"storage_capacity,omitempty" yaml:"storage_capacity,omitempty"`
	Resources        *Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
}

type RedisComponentInfo struct {
	SourceType  string `json:"source_type,omitempty" mapstructure:"connectType" yaml:"source_type,omitempty"`
	ConnectType string `json:"connect_type,omitempty" yaml:"connect_type,omitempty"`
	Username    string `json:"username,omitempty" mapstructure:"username" yaml:"username,omitempty"`
	Password    string `json:"password,omitempty" mapstructure:"password" yaml:"password,omitempty"`

	MasterHosts string `json:"master_hosts,omitempty" mapstructure:"masterHost" yaml:"master_hosts,omitempty"`
	MasterPort  int    `json:"master_port,omitempty" mapstructure:"masterPort" yaml:"master_port,omitempty"`
	SlaveHosts  string `json:"slave_hosts,omitempty" mapstructure:"slaveHost" yaml:"slave_hosts,omitempty"`
	SlavePort   int    `json:"slave_port,omitempty" mapstructure:"slavePort" yaml:"slave_port,omitempty"`

	SentinelHosts    string `json:"sentinel_hosts,omitempty" mapstructure:"sentinelHost" yaml:"sentinel_hosts,omitempty"`
	SentinelPort     int    `json:"sentinel_port,omitempty" mapstructure:"sentinelPort" yaml:"sentinel_port,omitempty"`
	SentinelUsername string `json:"sentinel_username,omitempty" mapstructure:"sentinelUsername" yaml:"sentinel_username,omitempty"`
	SentinelPassword string `json:"sentinel_password,omitempty" mapstructure:"sentinelPassword" yaml:"sentinel_password,omitempty"`
	MasterGroupName  string `json:"master_group_name,omitempty" mapstructure:"masterGroupName" yaml:"master_group_name,omitempty"`

	Hosts string `json:"hosts,omitempty" mapstructure:"host" yaml:"hosts,omitempty"`
	Port  int    `json:"port,omitempty" mapstructure:"port" yaml:"port,omitempty"`
}

func (c *ComponentRedis) ToBase() *ComponentObject {
	return &ComponentObject{
		Name:    c.Name,
		Type:    c.Type,
		Version: c.Version,
		Params:  mustToMap(c.Params),
		Info:    mustToMap(c.Info),
	}
}

func (c *ComponentObject) TryToRedis() (*ComponentRedis, error) {
	if c.Type != "redis" {
		return nil, fmt.Errorf("component type is not redis")
	}

	params, err := util.FromMap[RedisComponentParams](c.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert params to RedisComponentParams: %w", err)
	}

	info, err := util.FromMap[RedisComponentInfo](c.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to convert info to RedisComponentInfo: %w", err)
	}

	return &ComponentRedis{
		Name:    c.Name,
		Type:    c.Type,
		Version: c.Version,
		Params:  params,
		Info:    info,
	}, nil
}
