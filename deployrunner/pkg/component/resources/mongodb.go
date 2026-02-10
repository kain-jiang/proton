package resources

import (
	// load doc

	_ "embed"

	"taskrunner/trait"
)

// //go:embed schemas/mongodb.json
// var _mongodbSChema []byte

// MongoDB mongodb connect config
type MongoDB struct {
	SourceType string                 `json:"source_type"`
	Host       string                 `json:"hosts"`
	Port       float64                `json:"port"`
	User       string                 `json:"username"`
	Passws     string                 `json:"password"`
	Options    map[string]interface{} `json:"options"`
	Replicaset string                 `json:"replica_set"`
	SSL        bool                   `json:"ssl"`
	AuthSource string                 `json:"auth_source"`
	// MgmtInfo
	MgmtHost string `json:"mgmt_host,omitempty" yaml:"mgmt_host,omitempty"`
	MgmtPort int    `json:"mgmt_port,omitempty" yaml:"mgmt_port,omitempty"`
	AdminKey string `json:"admin_key,omitempty" yaml:"admin_key,omitempty"`
}

// ToDepMap conver into dep values
func (m *MongoDB) ToDepMap() (map[string]interface{}, *trait.Error) {
	// buf := bytes.NewBufferString("")
	opts := map[string]interface{}{}
	if m.Options != nil {
		opts = m.Options
	}
	if m.AuthSource != "" {
		opts["authSource"] = m.AuthSource
	}
	m.Options = opts
	// if m.Options != nil {
	// 	switch val := m.Options.(type) {
	// 	case map[string]any:
	// 		// useless for result, just for ut
	// 		keys := make([]string, 0, len(val))
	// 		for k := range val {
	// 			keys = append(keys, k)
	// 		}
	// 		sort.Strings(keys)
	// 		// seless for result, just for ut

	// 		for _, k := range keys {
	// 			v := val[k]
	// 			switch vv := v.(type) {
	// 			case string:
	// 				_, err = buf.WriteString(fmt.Sprintf("%s=%s&", k, vv))
	// 			case bool:
	// 				_, err = buf.WriteString(fmt.Sprintf("%s=%v&", k, vv))
	// 			case float64, float32:
	// 				_, err = buf.WriteString(fmt.Sprintf("%s=%v&", k, vv))
	// 			case int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
	// 				_, err = buf.WriteString(fmt.Sprintf("%s=%d&", k, vv))
	// 			default:
	// 				err = fmt.Errorf("mongodb options key %s is invalid: %#v", k, m.Options)
	// 				return nil, err
	// 			}

	// 		}

	// 		if len(val) >= 1 {
	// 			// remov last chart &
	// 			bslen := buf.Len()
	// 			buf.Truncate(bslen - 1)
	// 		}
	// 	case string:
	// 		_, err = buf.WriteString(val)
	// 	default:
	// 		err = fmt.Errorf("mongodb options is invalid: %#v", m.Options)
	// 	}
	// }
	vs := map[string]interface{}{
		"options":     m.Options,
		"host":        m.Host,
		"password":    m.Passws,
		"port":        m.Port,
		"replicaSet":  m.Replicaset,
		"ssl":         m.SSL,
		"user":        m.User,
		"source_type": m.SourceType,
		"mgmt_host":   m.MgmtHost,
		"mgmt_port":   m.MgmtPort,
		"admin_key":   m.AdminKey,
	}
	return vs, nil
}
