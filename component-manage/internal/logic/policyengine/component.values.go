package policyengine

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"component-manage/internal/global"
	"component-manage/internal/logic/base"
	"component-manage/pkg/models/types/components"

	"component-manage/internal/logic/etcd"
	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"

	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	PPEDefaultName             = "proton-policy-engine"
	PEClusterServiceNameSuffix = "cluster"
	PEClusterServicePort       = 9800
)

func policyengineValues(p *types.PolicyEngineComponentParams, e *components.ComponentETCD) helm3.M {
	v := helm3.M{
		"namespace": p.Namespace,
		"image": helm3.M{
			"registry": global.Config.Config.Registry,
		},
		"replicaCount": func() int {
			replicas := len(p.Hosts)
			if replicas == 0 {
				replicas = p.ReplicaCount
			}
			return replicas
		}(),
		"env": helm3.M{
			"language": "en_US.UTF-8",
			"timezone": "Asia/Shanghai",
		},
		"service": helm3.M{
			"type":            "ClusterIP",
			"port":            9800,
			"enableDualStack": global.Config.Config.EnableDualStack,
		},
		"storage": helm3.M{
			"storageClassName": p.StorageClassName,
			"local": func() helm3.M {
				rel := make(helm3.M)
				for i, host := range p.Hosts {
					rel[strconv.Itoa(i)] = helm3.M{
						"host": host,
						"path": p.Data_path,
					}
				}
				return rel
			}(),
		},
		"depServices": helm3.M{
			"etcd": helm3.M{
				"endpoints": strings.Join(etcd.GetETCDEndpointsInKubernetes(e), ","),
				"tlsConfig": helm3.M{
					"caName":     "ca.crt",
					"certName":   "peer.crt",
					"keyName":    "peer.key",
					"secretName": etcd.GetETCDName4MultiUser(etcd.ETCDSSLSecretNameBase, e.Name),
					"requireSSL": true,
				},
			},
		},
	}
	if len(p.StorageCapacity) > 0 {
		v["storage"].(helm3.M)["capacity"] = p.StorageCapacity
	}
	if p.Resources != nil {
		v["resources"] = p.Resources
	}
	return v
}

func preparePolicyEngine(param *types.PolicyEngineComponentParams) error {
	// 存储类无需准备
	if param.StorageClassName != "" {
		return nil
	}

	return base.PrepareStorage(param.Hosts, param.Data_path)
}

func clearPolicyEngine(param *types.PolicyEngineComponentParams) error {
	// 存储类无需准备
	if param.StorageClassName != "" {
		return nil
	}
	return base.ClearStorage(param.Hosts, param.Data_path)
}

func generateInfo(name string, params *types.PolicyEngineComponentParams) *components.PolicyEngineComponentInfo {
	return &types.PolicyEngineComponentInfo{
		SourceType: "internal",
		Hosts:      fmt.Sprintf("%s-%s.%s.%s", name, fmt.Sprintf("%s-%s", PPEDefaultName, PEClusterServiceNameSuffix), params.Namespace, global.Config.ServiceSuffix()),
		Port:       PEClusterServicePort,
	}
}

func checkPolicyEngine(p, oldp *types.PolicyEngineComponentParams) error {
	p.Namespace = base.DefaultString(p.Namespace, "resource")
	if p.StorageClassName != "" && len(p.Hosts) > 0 {
		return errors.New(".storageClassName and .hosts cannot be set at the same time")
	}
	if p.StorageClassName != "" && p.Data_path != "" {
		return errors.New(".storageClassName and .data_path cannot be set at the same time")
	}
	if len(p.Hosts) > 0 {
		set := sets.New[string](p.Hosts...)
		if set.Len() != 1 && set.Len() != 3 {
			return errors.New("policy engine only support 1 or 3 hosts")
		}
		if set.Len() < len(p.Hosts) {
			return errors.New("policy engine host list contains duplicate items")
		}
	}
	if oldp != nil {
		if base.DefaultString(oldp.Namespace, "resource") != p.Namespace {
			return errors.New("namespace is immutable")
		}
		if p.Data_path != oldp.Data_path {
			return errors.New("data path is immutable")
		}
		if p.StorageClassName != oldp.StorageClassName {
			return errors.New("storage class name is immutable")
		}
		// 仅支持扩容,不支持缩容
		replicaCountActual := func(p0 *components.PolicyEngineComponentParams) int {
			replicas := len(p0.Hosts)
			if replicas == 0 {
				return p0.ReplicaCount
			}
			return replicas
		}
		if replicaCountActual(p) < replicaCountActual(oldp) {
			return errors.New("Real replicaCount can not be reduced")
		} else if p.Hosts != nil && oldp.Hosts != nil {
			// 扩容时，新配置节点列表必须满足旧节点在最前
			for _, diff := range deep.Equal(p.Hosts[:len(oldp.Hosts)], oldp.Hosts) {
				return fmt.Errorf("previous hosts must be in front of new hosts when expanding policyengine deployment: %v, %v", p.Hosts, diff)
			}
		}
	}
	return nil
}
