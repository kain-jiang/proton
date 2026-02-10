package resources

import (
	"taskrunner/trait"
)

// Etcd etcd info
type Etcd struct {
	Hosts      string  `json:"hosts"`
	Port       float64 `json:"port"`
	Secret     string  `json:"secret"`
	SourceType string  `json:"source_type"`
	Namespace  string  `json:"namespace"`
}

// ToDepMap convert into dep values
func (e *Etcd) ToDepMap(ns string) (map[string]interface{}, *trait.Error) {
	mp := map[string]interface{}{
		"host":        e.Hosts,
		"port":        e.Port,
		"secret":      e.Secret,
		"sourceType":  e.SourceType,
		"source_type": e.SourceType,
		"namespace":   e.Namespace,
	}
	if ns != "" {
		mp["namespace"] = ns
	}
	return mp, nil
}
