package helm

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
)

// Values define the helm values of prometheus
type Values struct {
	Namespace string `json:"namespace,omitempty"`

	Image ValuesImage `json:"image,omitempty"`

	ReplicaCount int `json:"replicaCount,omitempty"`

	Service ValuesService `json:"service,omitempty"`

	Storage ValuesStorage `json:"storage,omitempty"`

	Secret ValuesSecret `json:"secret,omitempty"`

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
	EnableDualStack bool `json:"enableDualStack"`
}

// ValuesService define the .storage of then helm values
type ValuesStorage struct {
	StorageClassName string                 `json:"storageClassName,omitempty"`
	Capacity         string                 `json:"capacity,omitempty"`
	Local            map[string]ValuesLocal `json:"local,omitempty"`
}

// ValuesLocal define the .storage.local[xxx] of then helm values
type ValuesLocal struct {
	Host string `json:"host,omitempty"`

	Path string `json:"path,omitempty"`
}

// ValuesSecret define the .secret of then helm values, to provide ETCD certs to Prometheus
type ValuesSecret struct {
	K8sEtcd    ValuesEtcdCertInfo `json:"k8sEtcd,omitempty"`
	ProtonEtcd ValuesEtcdCertInfo `json:"protonEtcd,omitempty"`
}

// ValuesEtcdCertInfo define the .secret.k8sEtcd or .secret.ProtonEtcd helm values that contains detailed info about PrometheusETCD certs
type ValuesEtcdCertInfo struct {
	Enabled    bool   `json:"enabled"`
	SecretName string `json:"secretName,omitempty"`
	CaName     string `json:"caName,omitempty"`
	CertName   string `json:"certName,omitempty"`
	KeyName    string `json:"keyName,omitempty"`
}
