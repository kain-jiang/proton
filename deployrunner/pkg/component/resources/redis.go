package resources

import (
	// load doc
	_ "embed"
	"encoding/json"

	"taskrunner/trait"
)

// //go:embed schemas/redis_scheam.json
// var _RedisSchema []byte

// Redis redis config
type Redis struct {
	Source      string `json:"source_type"`
	ConnectType string `json:"connect_type"`
	User        string `json:"username"`
	Passwd      string `json:"password"`

	Host string  `json:"hosts"`
	Port float64 `json:"port"`

	MHost string  `json:"master_hosts"`
	MPort float64 `json:"master_port"`
	SHost string  `json:"slave_hosts"`
	SPort float64 `json:"slave_port"`

	MGroupName string  `json:"master_group_name"`
	SeHost     string  `json:"sentinel_hosts"`
	SePort     float64 `json:"sentinel_port"`
	SeUser     string  `json:"sentinel_username"`
	SePasswd   string  `json:"sentinel_password"`
}

// ToMap  convert redis obj into map
func (r *Redis) ToMap() (map[string]interface{}, *trait.Error) {
	var obj interface{}
	switch r.ConnectType {
	case "sentinel":
		if r.MGroupName == "" {
			r.MGroupName = "mymaster"
		}
		obj = SentinelRedisConnectInfo{
			MGroupName: r.MGroupName,
			SHost:      r.SeHost,
			SPort:      r.SePort,
			SUser:      r.SeUser,
			SPasswd:    r.SePasswd,
			User:       r.User,
			Passwd:     r.Passwd,
		}
	case "master-slave":
		obj = MasterSlaveRedisConnectInfo{
			MHost:  r.MHost,
			MPort:  r.MPort,
			SHost:  r.SHost,
			SPort:  r.SPort,
			User:   r.User,
			Passwd: r.Passwd,
		}
	case "standalone":
		obj = StandalondRedisConnectInfo{
			Host:   r.Host,
			Port:   r.Port,
			User:   r.User,
			Passwd: r.Passwd,
		}
	case "cluster":
		obj = StandalondRedisConnectInfo{
			Host:   r.Host,
			Port:   r.Port,
			User:   r.User,
			Passwd: r.Passwd,
		}
	}

	bs, err := json.Marshal(obj)
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECComponentDefined,
			Err:      err,
			Detail:   "encode redis type component into json bytes",
		}
	}
	mapObj := map[string]interface{}{
		"connectType": r.ConnectType,
		"sourceType":  r.Source,
	}
	info := map[string]interface{}{}
	err = json.Unmarshal(bs, &info)
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECComponentDefined,
			Err:      err,
			Detail:   "decode redis type component connectInfo",
		}
	}
	mapObj["connectInfo"] = info
	return mapObj, nil
}

// StandalondRedisConnectInfo standalone
type StandalondRedisConnectInfo struct {
	Host   string  `json:"host"`
	Port   float64 `json:"port"`
	User   string  `json:"username"`
	Passwd string  `json:"password"`
}

// MasterSlaveRedisConnectInfo master-slave
type MasterSlaveRedisConnectInfo struct {
	MHost  string  `json:"masterHost"`
	MPort  float64 `json:"masterPort"`
	SHost  string  `json:"slaveHost"`
	SPort  float64 `json:"slavePort"`
	User   string  `json:"username"`
	Passwd string  `json:"password"`
}

// SentinelRedisConnectInfo sentinal
type SentinelRedisConnectInfo struct {
	MGroupName string  `json:"masterGroupName"`
	SHost      string  `json:"sentinelHost"`
	SPort      float64 `json:"sentinelPort"`
	SUser      string  `json:"sentinelUsername"`
	SPasswd    string  `json:"sentinelPassword"`
	User       string  `json:"username"`
	Passwd     string  `json:"password"`
}
