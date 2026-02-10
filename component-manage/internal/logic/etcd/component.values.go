package etcd

import (
	"errors"
	"fmt"
	"strconv"

	"component-manage/internal/global"
	"component-manage/internal/logic/base"
	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"
	"component-manage/pkg/models/types/components"

	"github.com/go-test/deep"
)

const (
	ETCDClientPort           = 2379
	ETCDSSLSecretKeyNameBase = "etcdssl-secret-key"
	ETCDSSLSecretNameBase    = "etcdssl-secret"
	ETCDDefaultName          = "proton-etcd"
)

func ETCDValues(p *types.ETCDComponentParams, name string) helm3.M {
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
		"service": helm3.M{
			"enableDualStack": global.Config.Config.EnableDualStack,
			"type":            "ClusterIP",
			"port":            2379,
		},
		"auth": helm3.M{
			"client": helm3.M{
				"enableAuthentication": true,
				"existingSecret":       GetETCDName4MultiUser(ETCDSSLSecretNameBase, name),
				"certFilename":         "peer.crt",
				"certKeyFilename":      "peer.key",
				"caFilename":           "ca.crt",
			},
			"peer": helm3.M{
				"enableAuthentication": true,
				"existingSecret":       GetETCDName4MultiUser(ETCDSSLSecretNameBase, name),
				"certFilename":         "peer.crt",
				"certKeyFilename":      "peer.key",
				"caFilename":           "ca.crt",
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

func prepareETCD(param *types.ETCDComponentParams) error {
	// 存储类无需准备
	if param.StorageClassName != "" {
		return nil
	}

	return base.PrepareStorage(param.Hosts, param.Data_path)
}

func clearETCD(param *types.ETCDComponentParams) error {
	// 存储类无需准备
	if param.StorageClassName != "" {
		return nil
	}
	return base.ClearStorage(param.Hosts, param.Data_path)
}

func generateInfo(name string, param *types.ETCDComponentParams) *components.ETCDComponentInfo {
	return &types.ETCDComponentInfo{
		Namespace:  param.Namespace,
		SourceType: "internal",
		Hosts:      fmt.Sprintf("%s.%s.%s", base.TemplateName(name, "proton-etcd"), param.Namespace, global.Config.ServiceSuffix()),
		Port:       ETCDClientPort,
		Secret:     GetETCDName4MultiUser(ETCDSSLSecretNameBase, name),
	}
}

func checkETCD(p, oldp *types.ETCDComponentParams, name string) error {
	p.Namespace = base.DefaultString(p.Namespace, "resource")

	// 预防输入的名称刚好与用默认名称获取的secret名称相同
	for _, n := range []string{
		GetETCDName4MultiUser(ETCDSSLSecretKeyNameBase, name),
		GetETCDName4MultiUser(ETCDSSLSecretNameBase, name),
	} {
		if name == n {
			return fmt.Errorf("name cannot be %s", n)
		}
	}

	if p.StorageClassName != "" && len(p.Hosts) > 0 {
		return errors.New(".storageClassName and .hosts cannot be set at the same time")
	}
	if p.StorageClassName != "" && p.Data_path != "" {
		return errors.New(".storageClassName and .data_path cannot be set at the same time")
	}
	if oldp != nil {

		if base.DefaultString(oldp.Namespace, "resource") != p.Namespace {
			return errors.New("namespace is immutable")
		}

		if p.Hosts != nil && oldp.Hosts != nil && deep.Equal(p.Hosts, oldp.Hosts) != nil {
			return errors.New("hosts are immutable")
		}
		replicaCountActual := func(p0 *components.ETCDComponentParams) int {
			replicas := len(p0.Hosts)
			if replicas == 0 {
				return p0.ReplicaCount
			}
			return replicas
		}
		if replicaCountActual(p) != replicaCountActual(oldp) {
			return errors.New("replica count is immutable")
		}
		if p.Data_path != oldp.Data_path {
			return errors.New("data path is immutable")
		}
		if p.StorageClassName != oldp.StorageClassName {
			return errors.New("storage class name is immutable")
		}
		if p.StorageCapacity != oldp.StorageCapacity {
			return errors.New("storageCapacity can not be changed")
		}
	}
	return nil
}
