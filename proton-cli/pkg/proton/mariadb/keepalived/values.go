package keepalived

import (
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

type HelmValues struct {
	Env          map[string]string `json:"env,omitempty"`
	Image        *HelmValuesImage  `json:"image,omitempty"`
	Namespace    string            `json:"namespace,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	RBAC         *HelmValuesRBAC   `json:"rbac,omitempty"`
	VIP          *HelmValuesVIP    `json:"vip,omitempty"`
}

func (in *HelmValues) Map() map[string]interface{} {
	rel := make(map[string]interface{})
	jsonBytes, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(jsonBytes, &rel)
	if err != nil {
		panic(err)
	}
	return rel
}

type HelmValuesImage struct {
	PullPolicy core_v1.PullPolicy `json:"pullPolicy,omitempty"`
	Registry   string             `json:"registry,omitempty"`
	Repository string             `json:"repository,omitempty"`
	Tag        string             `json:"tag,omitempty"`
}

type HelmValuesRBAC struct {
	Create bool `json:"create,omitempty"`
}

type HelmValuesVIP struct {
	HTTPPort     int    `json:"httpPort,omitempty"`
	Interface    string `json:"iface,omitempty"`
	IP           string `json:"ip,omitempty"`
	OnlySelfNode bool   `json:"onlySelfNode,omitempty"`
	ServiceName  string `json:"serviceName,omitempty"`
	UseUnicast   bool   `json:"useUnicast,omitempty"`
	VRID         int    `json:"vrid,omitempty"`
}
