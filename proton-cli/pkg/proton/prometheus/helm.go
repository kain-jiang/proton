package prometheus

import (
	"strconv"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/prometheus/helm"
)

// values retuns helm values of the prometheus configuration
func (m *Manager) values() *helm.Values {
	var (
		count   = len(m.spec.Hosts)
		storage = helm.ValuesStorage{StorageClassName: m.spec.StorageClassName}
	)
	if len(m.spec.StorageCapacity) > 0 {
		storage.Capacity = m.spec.StorageCapacity
	}
	if len(m.spec.Hosts) == 0 {
		count = ReplicasForHostedKubernetes
	} else {
		storage.Local = make(map[string]helm.ValuesLocal)
	}
	for i, h := range m.spec.Hosts {
		storage.Local[strconv.Itoa(i)] = helm.ValuesLocal{Host: h, Path: m.spec.DataPath}
	}
	useK8SETCDCert := (m.csProvisioner == configuration.KubernetesProvisionerLocal)
	return &helm.Values{
		Namespace:    m.namespace,
		ReplicaCount: count,
		Image:        helm.ValuesImage{Registry: m.registry},
		Service:      helm.ValuesService{EnableDualStack: global.EnableDualStack},
		Storage:      storage,
		Secret: helm.ValuesSecret{
			ProtonEtcd: helm.ValuesEtcdCertInfo{
				Enabled:    m.isExistProtonETCD,
				SecretName: ProtonETCDResultSecretName,
				CaName:     ProtonETCDResultCAName,
				CertName:   ProtonETCDResultCertName,
				KeyName:    ProtonETCDResultKeyName,
			},
			K8sEtcd: helm.ValuesEtcdCertInfo{
				Enabled:    useK8SETCDCert,
				SecretName: K8SETCDResultSecretName,
				CaName:     K8SETCDResultCAName,
				CertName:   K8SETCDResultCertName,
				KeyName:    K8SETCDResultKeyName,
			},
		},
		Resources: m.spec.Resources,
	}
}
