package resources

import (
	// load  schema
	_ "embed"

	"taskrunner/trait"

	"github.com/mitchellh/mapstructure"
	"github.com/mohae/deepcopy"
)

const (
	MariadbDBType = "MARIADB"
	MysqlDBType   = "MYSQL"
	TIDBType      = "TIDB"
	GOLDENDBType  = "GOLDENDB"
	KDB9DBType    = "KDB9"
	DM8DBType     = "DM8"
)

//go:embed schemas/rds_schema.json
var _RdsSchema []byte

// RDS rds developer config
type RDS struct {
	// Type rds type
	Type string `json:"rds_type" mapstructure:"type"`
	// Host connect host
	Host string `json:"hosts" mapstructure:"host"`
	// Port connect Port
	Port float64 `json:"port" mapstructure:"port"`
	// user connect user
	User string `json:"username" mapstructure:"user"`
	// password user password
	Password string `json:"password" mapstructure:"password"`
	// SourceType source
	SourceType string `json:"source_type" mapstructure:"source_type"`
	// HostRead readonly connect host when using internal db
	HostRead string `json:"hosts_read" mapstructure:"hostsRead"`
	// PortRead readonly  connect port when using internal db
	PortRead float64 `json:"port_read" mapstructure:"portRead"`
	// AdminKey base64(adminUser:adminPasswd)
	AdminKey string `json:"admin_key" mapstructure:"admin_key"`

	MgmtHost string `json:"mgmt_host,omitempty" yaml:"mgmt_host,omitempty" mapstructure:"mgmt_host"`
	MgmtPort int    `json:"mgmt_port,omitempty" yaml:"mgmt_port,omitempty" mapstructure:"mgmt_port"`
	// DBName is the rds object config now, the sname time it's also database client config
	DBName string `json:"dbName,omitempty" mapstructure:"dbName"`
}

// ToMap convert obj into map
func (rds *RDS) ToMap() (map[string]interface{}, map[string]interface{}, *trait.Error) {
	attr := map[string]interface{}{
		"admin_key":   rds.AdminKey,
		"host":        rds.Host,
		"password":    rds.Password,
		"port":        rds.Port,
		"type":        rds.Type,
		"source_type": rds.SourceType,
		"user":        rds.User,
		"hostRead":    rds.HostRead,
		"portRead":    rds.PortRead,
		"mgmt_host":   rds.MgmtHost,
		"mgmt_port":   rds.MgmtPort,
	}

	cfg := deepcopy.Copy(attr).(map[string]interface{})

	return attr, cfg, nil
}

func RdsFromMap(m map[string]interface{}) (*RDS, *trait.Error) {
	rds := &RDS{}
	if err := mapstructure.Decode(m, rds); err != nil {
		return nil, &trait.Error{
			Internal: trait.ErrParam,
			Err:      err,
			Detail:   "decode rds from map error",
		}
	}
	return rds, nil
}
