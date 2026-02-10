package components

import (
	"fmt"

	"component-manage/pkg/util"
)

type ComponentPolicyEngine struct {
	Name         string                             `json:"name" yaml:"name"`
	Type         string                             `json:"type" yaml:"type"`
	Version      string                             `json:"version" yaml:"version"`
	Dependencies *PolicyEngineComponentDependencies `json:"dependencies" yaml:"dependencies"`
	Params       *PolicyEngineComponentParams       `json:"params" yaml:"params"`
	Info         *PolicyEngineComponentInfo         `json:"info"  yaml:"info"`
}

///////////////////////////////////////////////////////////

type PolicyEngineComponentDependencies struct {
	ETCD string `json:"etcd" yaml:"etcd"`
}

type PolicyEngineComponentParams struct {
	Namespace        string     `json:"namespace" yaml:"namespace"`
	ReplicaCount     int        `json:"replica_count,omitempty" yaml:"replica_count,omitempty"`
	Hosts            []string   `json:"hosts" yaml:"hosts"`
	Data_path        string     `json:"data_path" yaml:"data_path"`
	StorageClassName string     `json:"storageClassName,omitempty" yaml:"storageClassName,omitempty"`
	StorageCapacity  string     `json:"storage_capacity,omitempty" yaml:"storage_capacity,omitempty"`
	Resources        *Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
}

type PolicyEngineComponentInfo struct {
	SourceType string `json:"source_type,omitempty" mapstructure:"sourceType" yaml:"source_type,omitempty"`
	Hosts      string `json:"hosts,omitempty" mapstructure:"host" yaml:"hosts,omitempty"`
	Port       int    `json:"port,omitempty" mapstructure:"port" yaml:"port,omitempty"`
}

func (c *ComponentPolicyEngine) ToBase() *ComponentObject {
	return &ComponentObject{
		Name:         c.Name,
		Type:         c.Type,
		Version:      c.Version,
		Dependencies: mustToMap(c.Dependencies),
		Params:       mustToMap(c.Params),
		Info:         mustToMap(c.Info),
	}
}

func (o *ComponentObject) TryToPolicyEngine() (*ComponentPolicyEngine, error) {
	if o.Type != "policyengine" {
		return nil, fmt.Errorf("component type is not policyengine")
	}

	deps, err := util.FromMap[PolicyEngineComponentDependencies](o.Dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dependencies: %w", err)
	}

	params, err := util.FromMap[PolicyEngineComponentParams](o.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert params to PolicyEngineComponentParams: %w", err)
	}

	info, err := util.FromMap[PolicyEngineComponentInfo](o.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to convert info to PolicyEngineComponentInfo: %w", err)
	}

	return &ComponentPolicyEngine{
		Name:         o.Name,
		Type:         o.Type,
		Version:      o.Version,
		Dependencies: deps,
		Params:       params,
		Info:         info,
	}, nil
}
