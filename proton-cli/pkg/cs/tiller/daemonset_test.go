package tiller

import (
	"testing"

	"github.com/go-test/deep"
	api_apps_v1 "k8s.io/api/apps/v1"
	api_core_v1 "k8s.io/api/core/v1"
	api_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestNewDaemonSet(t *testing.T) {
	type args struct {
		registry string
	}
	tests := []struct {
		name string
		args args
		want *api_apps_v1.DaemonSet
	}{
		{
			name: "acr.aishu.cn",
			args: args{
				registry: "acr.aishu.cn",
			},
			want: &api_apps_v1.DaemonSet{
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
									Name:  ContainerName,
									Image: "acr.aishu.cn/public/kubernetes-helm/tiller:v2.16.9",
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
			},
		},
		{
			name: "registry.aishu.cn:15000",
			args: args{
				registry: "registry.aishu.cn:15000",
			},
			want: &api_apps_v1.DaemonSet{
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
									Name:  ContainerName,
									Image: "registry.aishu.cn:15000/public/kubernetes-helm/tiller:v2.16.9",
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDaemonSet(tt.args.registry)
			for _, diff := range deep.Equal(got, tt.want) {
				t.Errorf("NewDaemonSet() got vs want: %v", diff)
			}
		})
	}
}
