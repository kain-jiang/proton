package resources

import (
	// load doc

	_ "embed"

	"taskrunner/trait"
)

//go:embed schemas/graphdb.json
var _graphdbSChema []byte

// GraphDB graph database
type GraphDB struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	User         string `json:"user"`
	Password     string `json:"password"`
	ReadonlyUser string `json:"readonlyuser"`
	ReadonlyPass string `json:"readonlypassword"`
	Type         string `json:"type"`
	Source       string `json:"source_type"`
}

// ToDepMap convert into dep values
func (e *GraphDB) ToDepMap() (map[string]interface{}, *trait.Error) {
	return map[string]interface{}{
		"host":             e.Host,
		"port":             e.Port,
		"type":             e.Type,
		"readonlyuser":     e.ReadonlyUser,
		"readonlypassword": e.ReadonlyPass,
		"user":             e.User,
		"password":         e.Password,
	}, nil
}
