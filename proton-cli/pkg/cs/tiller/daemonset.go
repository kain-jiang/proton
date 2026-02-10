package tiller

import (
	"fmt"

	api_apps_v1 "k8s.io/api/apps/v1"
	api_core_v1 "k8s.io/api/core/v1"
	api_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ImageRepository = "public/kubernetes-helm/tiller"
	ImageTag        = "v2.16.9"

	DaemonSetName = "tiller-deploy"

	ContainerName = "tiller"

	ContainerArgStorageSecret = "-storage=secret"

	ContainerEnvTillerNamespaceName  = "TILLER_NAMESPACE"
	ContainerEnvTillerNamespaceValue = "kube-system"

	ContainerEnvTillerHistoryMaxName  = "TILLER_HISTORY_MAX"
	ContainerEnvTillerHistoryMaxValue = "0"

	ContainerEnvTillerKubernetesServiceHostName  = "KUBERNETES_SERVICE_HOST"
	ContainerEnvTillerKubernetesServiceHostValue = "kubernetes.default"

	ContainerEnvTillerKubernetesServicePortName  = "KUBERNETES_SERVICE_PORT"
	ContainerEnvTillerKubernetesServicePortValue = "443"

	ContainerPortTillerName          = "tiller"
	ContainerPortTillerContainerPort = Port
	ContainerPortHTTPName            = "http"
	ContainerPortHTTPContainerPort   = HTTPPort

	ContainerProbeInitialDelaySeconds = 1
	ContainerProbeTimeoutSeconds      = 1

	ContainerLivenessProbePath  = "/liveness"
	ContainerReadinessProbePath = "/readiness"
)

var DaemonSetPrototype = api_apps_v1.DaemonSet{
	ObjectMeta: api_meta_v1.ObjectMeta{
		Name:   DaemonSetName,
		Labels: TillerLabels,
	},
	Spec: api_apps_v1.DaemonSetSpec{
		Selector: &api_meta_v1.LabelSelector{
			MatchLabels: TillerLabels,
		},
		Template: api_core_v1.PodTemplateSpec{
			ObjectMeta: api_meta_v1.ObjectMeta{
				Labels: TillerLabels,
			},
			Spec: api_core_v1.PodSpec{
				Containers: []api_core_v1.Container{
					{
						Name: ContainerName,
						Args: []string{
							ContainerArgStorageSecret,
						},
						Ports: []api_core_v1.ContainerPort{
							{
								Name:          ContainerPortTillerName,
								ContainerPort: ContainerPortTillerContainerPort,
							},
							{
								Name:          ContainerPortHTTPName,
								ContainerPort: ContainerPortHTTPContainerPort,
							},
						},
						Env: []api_core_v1.EnvVar{
							{
								Name:  ContainerEnvTillerNamespaceName,
								Value: ContainerEnvTillerNamespaceValue,
							},
							{
								Name:  ContainerEnvTillerHistoryMaxName,
								Value: ContainerEnvTillerHistoryMaxValue,
							},
							{
								Name:  ContainerEnvTillerKubernetesServiceHostName,
								Value: ContainerEnvTillerKubernetesServiceHostValue,
							},
							{
								Name:  ContainerEnvTillerKubernetesServicePortName,
								Value: ContainerEnvTillerKubernetesServicePortValue,
							},
						},
						LivenessProbe: &api_core_v1.Probe{
							ProbeHandler: api_core_v1.ProbeHandler{
								HTTPGet: &api_core_v1.HTTPGetAction{
									Path: ContainerLivenessProbePath,
									Port: intstr.FromInt(ContainerPortHTTPContainerPort),
								},
							},
							InitialDelaySeconds: ContainerProbeInitialDelaySeconds,
							TimeoutSeconds:      ContainerProbeTimeoutSeconds,
						},
						ReadinessProbe: &api_core_v1.Probe{
							ProbeHandler: api_core_v1.ProbeHandler{
								HTTPGet: &api_core_v1.HTTPGetAction{
									Path: ContainerReadinessProbePath,
									Port: intstr.FromInt(ContainerPortHTTPContainerPort),
								},
							},
							InitialDelaySeconds: ContainerProbeInitialDelaySeconds,
							TimeoutSeconds:      ContainerProbeTimeoutSeconds,
						},
					},
				},
				ServiceAccountName: ServiceAccountName,
			},
		},
	},
}

func NewDaemonSet(registry string) *api_apps_v1.DaemonSet {
	ds := DaemonSetPrototype.DeepCopy()
	for i := range ds.Spec.Template.Spec.Containers {
		if ds.Spec.Template.Spec.Containers[i].Name == ContainerName {
			ds.Spec.Template.Spec.Containers[i].Image = fmt.Sprintf("%s/%s:%s", registry, ImageRepository, ImageTag)
		}
	}
	return ds
}
