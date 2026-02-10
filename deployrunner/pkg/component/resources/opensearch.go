package resources

import (
	// load doc
	_ "embed"

	"taskrunner/trait"
)

// //go:embed schemas/opensearch.json
// var _OpensearchConfigSchema []byte

// Opensearch connect config
type Opensearch struct {
	SourceType   string  `json:"source_type"`
	Host         string  `json:"hosts"`
	Port         float64 `json:"port"`
	Passwd       string  `json:"password"`
	User         string  `json:"username"`
	Version      string  `json:"version"`
	Protocotl    string  `json:"protocol"`
	Distribution string  `json:"distribution"`
}

// ToDepMap convert into deps valeus
func (o *Opensearch) ToDepMap() (map[string]interface{}, *trait.Error) {
	return map[string]interface{}{
		"host":         o.Host,
		"port":         o.Port,
		"user":         o.User,
		"password":     o.Passwd,
		"protocol":     o.Protocotl,
		"version":      o.Version,
		"distribution": o.Distribution,
	}, nil
}
