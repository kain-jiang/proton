package resources

import (
	"encoding/json"
	// load  schema
	_ "embed"

	"taskrunner/trait"
)

// //go:embed schemas/mq_schema.json
// var _MQSchema []byte

// Nsq mq
type Nsq struct {
	Host        string  `json:"nsqHost"`
	Port        float64 `json:"nsqPort"`
	LookupdHost string  `json:"nsqLookupdHost"`
	LookUpdPort float64 `json:"nsqLookupdPort"`
	MqType      string  `json:"mqType"`
}

// ToMap convert obj into map
func (mq *Nsq) ToMap() (map[string]interface{}, *trait.Error) {
	bs, err := json.Marshal(mq)
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECComponentDefined,
			Err:      err,
			Detail:   "encdode mq type component's impl nsq into json bytes",
		}
	}
	vs := map[string]interface{}{}
	err = json.Unmarshal(bs, &vs)
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECComponentDefined,
			Err:      err,
			Detail:   "decode mq type component's impl nsq into map",
		}
	}
	return vs, nil
}

// MQAuth kafka auth config
type MQAuth struct {
	Username  string `json:"username,omitempty"`
	Passwd    string `json:"password,omitempty"`
	Mechanism string `json:"mechanism,omitempty"`
}

// MQ developer config
type MQ struct {
	SourceType string `json:"source_type,omitempty"`
	// copy from MqType
	MqType string  `json:"mq_type"`
	MqHost string  `json:"mq_hosts"`
	MqPort float64 `json:"mq_port"`
	// copy from mqHost
	MqLookupdHost string `json:"mq_lookupd_hosts"`
	// copy from mqPort
	MqLookupdPort float64 `json:"mq_lookupd_port"`
	// KAFKA auth
	Auth *MQAuth `json:"auth"`
}

// ToMap convert obj into map
func (mq *MQ) ToMap() (map[string]interface{}, *trait.Error) {
	auth := make(map[string]interface{}, 3)
	if mq.Auth != nil {
		auth["username"] = mq.Auth.Username
		auth["password"] = mq.Auth.Passwd
		auth["mechanism"] = mq.Auth.Mechanism
	}
	return map[string]interface{}{
		"mqType":        mq.MqType,
		"mqHost":        mq.MqHost,
		"mqPort":        mq.MqPort,
		"mqLookupdHost": mq.MqLookupdHost,
		"mqLookupdPort": mq.MqLookupdPort,
		"auth":          auth,
	}, nil
}
