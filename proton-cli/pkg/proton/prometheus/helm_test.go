package prometheus

import (
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/prometheus/helm"
)

func TestManager_values(t *testing.T) {
	const (
		registry = "registry.example.org"

		node0 = "node-0"
		node1 = "node-1"

		dataPath = "/var/lib/prometheus"

		storageClassName = "standard"
	)
	var (
		logger = logrus.Logger{Out: os.Stdout, Hooks: make(logrus.LevelHooks), Formatter: new(logrus.TextFormatter), Level: logrus.DebugLevel}

		valuesLocal0 = helm.ValuesLocal{Host: node0, Path: dataPath}
		valuesLocal1 = helm.ValuesLocal{Host: node1, Path: dataPath}

		storageLocalSingle = helm.ValuesStorage{Local: map[string]helm.ValuesLocal{"0": valuesLocal0}}
		storageLocalMulti  = helm.ValuesStorage{Local: map[string]helm.ValuesLocal{"0": valuesLocal0, "1": valuesLocal1}}
		storageHosted      = helm.ValuesStorage{StorageClassName: storageClassName}

		predefSecretStructLocal = helm.ValuesSecret{
			ProtonEtcd: helm.ValuesEtcdCertInfo{
				Enabled:    true,
				SecretName: ProtonETCDResultSecretName,
				CaName:     ProtonETCDResultCAName,
				CertName:   ProtonETCDResultCertName,
				KeyName:    ProtonETCDResultKeyName,
			},
			K8sEtcd: helm.ValuesEtcdCertInfo{
				Enabled:    true,
				SecretName: K8SETCDResultSecretName,
				CaName:     K8SETCDResultCAName,
				CertName:   K8SETCDResultCertName,
				KeyName:    K8SETCDResultKeyName,
			},
		}
		predefSecretStructLocalWithoutETCD = helm.ValuesSecret{
			ProtonEtcd: helm.ValuesEtcdCertInfo{
				Enabled:    false,
				SecretName: ProtonETCDResultSecretName,
				CaName:     ProtonETCDResultCAName,
				CertName:   ProtonETCDResultCertName,
				KeyName:    ProtonETCDResultKeyName,
			},
			K8sEtcd: helm.ValuesEtcdCertInfo{
				Enabled:    true,
				SecretName: K8SETCDResultSecretName,
				CaName:     K8SETCDResultCAName,
				CertName:   K8SETCDResultCertName,
				KeyName:    K8SETCDResultKeyName,
			},
		}
		predefSecretStructExternal = helm.ValuesSecret{
			ProtonEtcd: helm.ValuesEtcdCertInfo{
				Enabled:    true,
				SecretName: ProtonETCDResultSecretName,
				CaName:     ProtonETCDResultCAName,
				CertName:   ProtonETCDResultCertName,
				KeyName:    ProtonETCDResultKeyName,
			},
			K8sEtcd: helm.ValuesEtcdCertInfo{
				Enabled:    false,
				SecretName: K8SETCDResultSecretName,
				CaName:     K8SETCDResultCAName,
				CertName:   K8SETCDResultCertName,
				KeyName:    K8SETCDResultKeyName,
			},
		}

		quantity50m   = resource.MustParse("50m")
		quantity200m  = resource.MustParse("200m")
		quantity16Mi  = resource.MustParse("16Mi")
		quantity128Mi = resource.MustParse("128Mi")

		resources = &corev1.ResourceRequirements{
			Limits:   corev1.ResourceList{corev1.ResourceCPU: quantity200m, corev1.ResourceMemory: quantity128Mi},
			Requests: corev1.ResourceList{corev1.ResourceCPU: quantity50m, corev1.ResourceMemory: quantity16Mi},
		}
	)
	tests := []struct {
		name string
		m    *Manager
		want *helm.Values
	}{
		{
			name: "local kubernetes single node",
			m: &Manager{
				registry:          registry,
				spec:              &configuration.Prometheus{Hosts: []string{node0}, DataPath: dataPath, Resources: resources},
				csProvisioner:     configuration.KubernetesProvisionerLocal,
				isExistProtonETCD: true,
			},
			want: &helm.Values{
				Image:        helm.ValuesImage{Registry: registry},
				ReplicaCount: 1,
				Service:      helm.ValuesService{EnableDualStack: global.EnableDualStack},
				Storage:      storageLocalSingle,
				Secret:       predefSecretStructLocal,
				Resources:    resources,
			},
		},
		{
			name: "local kubernetes multi nodes",
			m: &Manager{
				registry:          registry,
				spec:              &configuration.Prometheus{Hosts: []string{node0, node1}, DataPath: dataPath, Resources: resources},
				csProvisioner:     configuration.KubernetesProvisionerLocal,
				isExistProtonETCD: true,
			},
			want: &helm.Values{
				Image:        helm.ValuesImage{Registry: registry},
				ReplicaCount: 2,
				Service:      helm.ValuesService{EnableDualStack: global.EnableDualStack},
				Storage:      storageLocalMulti,
				Secret:       predefSecretStructLocal,
				Resources:    resources,
			},
		},
		{
			name: "local kubernetes multi nodes without Proton ETCD",
			m: &Manager{
				registry:          registry,
				spec:              &configuration.Prometheus{Hosts: []string{node0, node1}, DataPath: dataPath, Resources: resources},
				csProvisioner:     configuration.KubernetesProvisionerLocal,
				isExistProtonETCD: false,
			},
			want: &helm.Values{
				Image:        helm.ValuesImage{Registry: registry},
				ReplicaCount: 2,
				Service:      helm.ValuesService{EnableDualStack: global.EnableDualStack},
				Storage:      storageLocalMulti,
				Secret:       predefSecretStructLocalWithoutETCD,
				Resources:    resources,
			},
		},
		{
			name: "hosted kubernetes",
			m: &Manager{
				registry:          registry,
				spec:              &configuration.Prometheus{StorageClassName: storageClassName, Resources: resources},
				csProvisioner:     configuration.KubernetesProvisionerExternal,
				isExistProtonETCD: true,
			},
			want: &helm.Values{
				Image:        helm.ValuesImage{Registry: registry},
				ReplicaCount: ReplicasForHostedKubernetes,
				Service:      helm.ValuesService{EnableDualStack: global.EnableDualStack},
				Storage:      storageHosted,
				Secret:       predefSecretStructExternal,
				Resources:    resources,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.logger = logger.WithField("test", tt.name)
			got := tt.m.values()
			for _, d := range deep.Equal(got, tt.want) {
				t.Errorf("Manager.values() got != want: %v", d)
			}
		})
	}
}
