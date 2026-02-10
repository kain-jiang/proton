package tiller

import (
	api_core_v1 "k8s.io/api/core/v1"
	api_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ServiceName = "tiller-deploy"

	ServiceType = api_core_v1.ServiceTypeClusterIP

	ServicePortTillerName       = "tiller"
	ServicePortTillerTargetPort = "tiller"
)

var Service = api_core_v1.Service{
	ObjectMeta: api_meta_v1.ObjectMeta{
		Name:   ServiceName,
		Labels: TillerLabels,
	},
	Spec: api_core_v1.ServiceSpec{
		Ports: []api_core_v1.ServicePort{
			{
				Name:       ServicePortTillerName,
				Port:       Port,
				TargetPort: intstr.FromString(ServicePortTillerTargetPort),
			},
		},
		Selector: TillerLabels,
		Type:     ServiceType,
	},
}
