package components

import (
	"fmt"

	"component-manage/pkg/util"
)

type ComponentPrometheus struct {
	Name         string                           `json:"name" yaml:"name"`
	Type         string                           `json:"type" yaml:"type"`
	Version      string                           `json:"version" yaml:"version"`
	Dependencies *PrometheusComponentDependencies `json:"dependencies" yaml:"dependencies"`
	Params       *PrometheusComponentParams       `json:"params" yaml:"params"`
	Info         *PrometheusComponentInfo         `json:"info" yaml:"info"`
}

///////////////////////////////////////////////////////////

type PrometheusComponentDependencies struct {
	ProtonETCD string `json:"etcd" yaml:"etcd"`
}

type CAInfo struct {
	Namespace   string `json:"namespace" yaml:"namespace"`
	SecretName  string `json:"secret_name" yaml:"secret_name"`
	CertKeyname string `json:"cert_keyname" yaml:"cert_keyname"`
	KeyKeyname  string `json:"key_keyname" yaml:"key_keyname"`
}

type PrometheusComponentParams struct {
	Namespace        string   `json:"namespace" yaml:"namespace"`
	ReplicaCount     int      `json:"replica_count,omitempty" yaml:"replica_count,omitempty"`
	Hosts            []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	DataPath         string   `json:"data_path,omitempty" yaml:"data_path,omitempty"`
	StorageClassName string   `json:"storageClassName,omitempty" yaml:"storageClassName,omitempty"`
	StorageCapacity  string   `json:"storage_capacity,omitempty" yaml:"storage_capacity,omitempty"`
	CAInfo           *struct {
		K8sEtcd    *CAInfo     `json:"k8sEtcd" yaml:"k8sEtcd" binding:"required"` // k8s etcd 的 ca 信息，需要传入
		ProtonEtcd *ETCDCAInfo `json:"protonEtcd" yaml:"protonEtcd"`              // proton etcd 的 ca 信息，由依赖自动注入
	} `json:"caInfo" yaml:"caInfo" binding:"required"`
	PrometheusConfig map[string]any `json:"prometheusConfig,omitempty" yaml:"prometheusConfig,omitempty"`
	Resources        *Resources     `json:"resources,omitempty" yaml:"resources,omitempty"`
}

// TODO
type PrometheusComponentInfo struct{}

func (c *ComponentPrometheus) ToBase() *ComponentObject {
	return &ComponentObject{
		Name:         c.Name,
		Type:         c.Type,
		Version:      c.Version,
		Dependencies: mustToMap(c.Dependencies),
		Params:       mustToMap(c.Params),
		Info:         mustToMap(c.Info),
	}
}

func (o *ComponentObject) TryToPrometheus() (*ComponentPrometheus, error) {
	if o.Type != "prometheus" {
		return nil, fmt.Errorf("component type is not prometheus")
	}

	deps, err := util.FromMap[PrometheusComponentDependencies](o.Dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dependencies: %w", err)
	}

	params, err := util.FromMap[PrometheusComponentParams](o.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert params to PrometheusComponentParams: %w", err)
	}

	info, err := util.FromMap[PrometheusComponentInfo](o.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to convert info to PrometheusComponentInfo: %w", err)
	}

	return &ComponentPrometheus{
		Name:         o.Name,
		Type:         o.Type,
		Version:      o.Version,
		Dependencies: deps,
		Params:       params,
		Info:         info,
	}, nil
}
