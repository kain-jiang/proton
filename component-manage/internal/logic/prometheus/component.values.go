package prometheus

import (
	"errors"
	"fmt"
	"strconv"

	"component-manage/internal/global"
	"component-manage/internal/logic/base"
	"component-manage/internal/logic/etcd"
	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"
)

// tools
func toList(l []string) []string {
	if l == nil {
		return make([]string, 0)
	}
	return l
}

func defaultTo[T helm3.M](val, dft T) T {
	if val == nil {
		return dft
	}
	return val
}

func checkPrometheus(param, oldParam *types.PrometheusComponentParams, etcdObj *types.ComponentETCD) error {
	param.Namespace = base.DefaultString(param.Namespace, "resource")
	// 计算 replicaCount，优先取 Hosts 长度，如果没有，再取 ReplicaCount
	param.ReplicaCount = func() int {
		replicas := len(param.Hosts)
		if replicas == 0 {
			return param.ReplicaCount
		}
		return replicas
	}()

	param.CAInfo.ProtonEtcd = etcd.GetEtcdCAInfo(etcdObj)

	if oldParam != nil {
		// 更新
		if base.DefaultString(oldParam.Namespace, "resource") != param.Namespace {
			return errors.New("namespace is immutable")
		}
		// Secret信息不能变
		param.CAInfo = oldParam.CAInfo

		if param.StorageClassName != oldParam.StorageClassName {
			return errors.New("storageClassName can not be changed")
		}

		if param.StorageCapacity != oldParam.StorageCapacity {
			return errors.New("storageCapacity can not be changed")
		}

		if param.DataPath != oldParam.DataPath {
			return errors.New("data_path can not be changed")
		}

		if len(param.Hosts) < len(oldParam.Hosts) {
			return errors.New("hosts can not be reduced")
		}

		if param.ReplicaCount < oldParam.ReplicaCount {
			return errors.New("Real replicaCount can not be reduced")
		}

		for idx, oH := range toList(oldParam.Hosts) {
			h := param.Hosts[idx]
			if h != oH {
				return fmt.Errorf("hosts can not be change to %s from %s", h, oH)
			}
		}

	}

	return nil
}

func PrometheusValues(param *types.PrometheusComponentParams, certs etcdCertsForPrometheus) helm3.M {
	v := helm3.M{
		"namespace":    param.Namespace,
		"replicaCount": param.ReplicaCount,
		"image": helm3.M{
			"registry": global.Config.Config.Registry,
		},
		"service": helm3.M{
			"enableDualStack": global.Config.Config.EnableDualStack,
		},
		"secret": helm3.M{
			"k8sEtcd": helm3.M{
				"caName":     certs.K8sEtcdCert.CaCertFieldName,
				"certName":   certs.K8sEtcdCert.CertFieldName,
				"keyName":    certs.K8sEtcdCert.CertKeyFieldName,
				"secretName": certs.K8sEtcdCert.SecretName,
				"enabled":    true,
			},
			"protonEtcd": helm3.M{
				"caName":     certs.ProtonEtcdCert.CaCertFieldName,
				"certName":   certs.ProtonEtcdCert.CertFieldName,
				"keyName":    certs.ProtonEtcdCert.CertKeyFieldName,
				"secretName": certs.ProtonEtcdCert.SecretName,
				"enabled":    true,
			},
		},
		"storage": helm3.M{
			"storageClassName": param.StorageClassName,
			"local": func() helm3.M {
				rel := make(helm3.M)
				for i, host := range param.Hosts {
					rel[strconv.Itoa(i)] = helm3.M{
						"host": host,
						"path": param.DataPath,
					}
				}
				return rel
			}(),
		},
		"config": helm3.M{
			"prometheus": defaultTo(param.PrometheusConfig, helm3.M{}),
		},
	}
	if len(param.StorageCapacity) > 0 {
		v["storage"].(helm3.M)["capacity"] = param.StorageCapacity
	}
	if param.Resources != nil {
		v["resources"] = param.Resources
	}
	return v
}

func preparePrometheus(param *types.PrometheusComponentParams) error {
	// 存储类无需准备
	if param.StorageClassName != "" {
		return nil
	}

	return base.PrepareStorage(param.Hosts, param.DataPath)
}

func generateInfo(name string, params *types.PrometheusComponentParams) *types.PrometheusComponentInfo {
	_, _ = name, params
	return &types.PrometheusComponentInfo{}
}

func clearPrometheus(param *types.PrometheusComponentParams) error {
	// 存储类无需清理
	if param.StorageClassName != "" {
		return nil
	}
	return base.ClearStorage(param.Hosts, param.DataPath)
}
