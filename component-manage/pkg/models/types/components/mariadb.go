package components

import (
	"fmt"

	"component-manage/pkg/util"
)

type MariaDBComponentParams struct {
	Namespace    string   `json:"namespace" yaml:"namespace"`
	ReplicaCount int      `json:"replica_count,omitempty" yaml:"replica_count,omitempty"`
	Hosts        []string `json:"hosts" yaml:"hosts"`
	Config       *struct {
		LowerCaseTableNames      *int   `json:"lower_case_table_names,omitempty" yaml:"lower_case_table_names,omitempty"`
		Thread_handling          string `json:"thread_handling,omitempty" yaml:"thread_handling,omitempty"`
		Innodb_buffer_pool_size  string `json:"innodb_buffer_pool_size" yaml:"innodb_buffer_pool_size"`
		Resource_requests_memory string `json:"resource_requests_memory" yaml:"resource_requests_memory"`
		Resource_limits_memory   string `json:"resource_limits_memory" yaml:"resource_limits_memory"`
	} `json:"config" yaml:"config"`
	Admin_user   string `json:"admin_user" yaml:"admin_user"`
	Admin_passwd string `json:"admin_passwd" yaml:"admin_passwd"`
	Data_path    string `json:"data_path" yaml:"data_path"`

	StorageClassName string `json:"storageClassName,omitempty" yaml:"storageClassName,omitempty"`
	StorageCapacity  string `json:"storage_capacity,omitempty" yaml:"storage_capacity,omitempty"`

	// 新增配置
	Username        string `json:"username" yaml:"username" binding:"required"`
	Password        string `json:"password" yaml:"password" binding:"required"`
	AdminSecretName string `json:"admin_secret_name" yaml:"admin_secret_name"`
}

type MariaDBComponentInfo struct {
	SourceType string `json:"source_type,omitempty" yaml:"source_type,omitempty"`
	RdsType    string `json:"rds_type,omitempty" yaml:"rds_type,omitempty" `
	Hosts      string `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Port       int    `json:"port,omitempty" yaml:"port,omitempty"`
	Username   string `json:"username" yaml:"username"`
	Password   string `json:"password" yaml:"password"`
	HostsRead  string `json:"hosts_read,omitempty" yaml:"hosts_read,omitempty"`
	PortRead   int    `json:"port_read,omitempty" yaml:"port_read,omitempty"`
	// MgmtInfo
	MgmtHost string `json:"mgmt_host,omitempty" yaml:"mgmt_host,omitempty"`
	MgmtPort int    `json:"mgmt_port,omitempty" yaml:"mgmt_port,omitempty"`
	AdminKey string `json:"admin_key,omitempty" yaml:"admin_key,omitempty"`
}

type ComponentMariaDB struct {
	Name    string                  `json:"name" yaml:"name"`
	Type    string                  `json:"type" yaml:"type"`
	Version string                  `json:"version" yaml:"version"`
	Params  *MariaDBComponentParams `json:"params" yaml:"params"`
	Info    *MariaDBComponentInfo   `json:"info" yaml:"info"`
}

func (c *ComponentMariaDB) ToBase() *ComponentObject {
	return &ComponentObject{
		Name:    c.Name,
		Type:    c.Type,
		Version: c.Version,
		Params:  mustToMap(c.Params),
		Info:    mustToMap(c.Info),
	}
}

func (o *ComponentObject) TryToMariaDB() (*ComponentMariaDB, error) {
	if o.Type != "mariadb" {
		return nil, fmt.Errorf("component type is not kafka")
	}

	params, err := util.FromMap[MariaDBComponentParams](o.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert params to MariaDBComponentParams: %w", err)
	}

	info, err := util.FromMap[MariaDBComponentInfo](o.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to convert info to MariaDBComponentInfo: %w", err)
	}

	return &ComponentMariaDB{
		Name:    o.Name,
		Type:    o.Type,
		Version: o.Version,
		Params:  params,
		Info:    info,
	}, nil
}
