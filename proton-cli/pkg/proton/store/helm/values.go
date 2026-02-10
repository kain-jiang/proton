package helm

import (
	"encoding/json"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

// Values defines helm values of proton package store.
type Values struct {
	Image Image `json:"image"`

	ReplicaCount int `json:"replicaCount"`

	DepServices DepServices `json:"depServices"`

	Storage Storage `json:"storage"`

	Resources *Resources `json:"resources,omitempty"`

	Namespace string `json:"namespace,omitempty"`
}

func ValuesFor(spec *configuration.PackageStore, registry string, rds *configuration.RdsInfo, database, namespace string) *Values {
	return &Values{
		Image:        imageFor(registry),
		ReplicaCount: *spec.Replicas,
		DepServices:  depServicesFor(rds, database),
		Storage:      storageFor(spec),
		Resources:    resourcesFor(spec.Resources),
		Namespace:    namespace,
	}
}

// ToMap returns map[string]interface{} for helm installing or upgrading.
func (v *Values) ToMap() map[string]interface{} {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		panic(err)
	}
	return m
}
