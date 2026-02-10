package helm

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
)

// Values define the helm values of grafana
type Values struct {
	Namespace string `json:"namespace,omitempty"`

	Image ValuesImage `json:"image,omitempty"`

	ReplicaCount int `json:"replicaCount,omitempty"`

	Service ValuesService `json:"service,omitempty"`

	Config ValuesConfig `json:"config,omitempty"`

	Storage ValuesStorage `json:"storage,omitempty"`

	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// ToMap returns map[string]interface{} used for helm release's config
func (v *Values) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	b, _ := json.Marshal(&v)
	_ = json.Unmarshal(b, &m)
	return m
}

// ValuesImage define the .image of then helm values
type ValuesImage struct {
	Registry string `json:"registry,omitempty"`
}

// ValuesService define the .service of then helm values
type ValuesService struct {
	EnableDualStack bool `json:"enableDualStack,omitempty"`

	Grafana ValuesGrafanaService `json:"grafana,omitempty"`
}

type ValuesGrafanaService struct {
	Type corev1.ServiceType `json:"type,omitempty"`

	NodePort int32 `json:"nodePort,omitempty"`
}

type ValuesConfig struct {
	DataSource ValuesDataSource `json:"datasource,omitempty"`
}

type ValuesDataSource struct {
	Prometheus ValuesPrometheus `json:"prometheus,omitempty"`
}

type ValuesPrometheus struct {
	Enabled bool `json:"enabled,omitempty"`

	Protocol ValuesProtocol `json:"protocol,omitempty"`

	Host string `json:"host,omitempty"`

	Port int `json:"port,omitempty"`
}

type ValuesProtocol string

const (
	// ValuesProtocolHTTP means that the protocol used will be http
	ValuesProtocolHTTP = "http"
)

// ValuesService define the .storage of then helm values
type ValuesStorage struct {
	StorageClassName string `json:"storageClassName,omitempty"`
	Capacity         string `json:"capacity,omitempty"`

	Local map[string]ValuesLocal `json:"local,omitempty"`
}

// ValuesLocal define the .storage.local[xxx] of then helm values
type ValuesLocal struct {
	Host string `json:"host,omitempty"`

	Path string `json:"path,omitempty"`
}
