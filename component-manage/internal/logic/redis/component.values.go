package redis

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"

	"component-manage/internal/global"
	"component-manage/internal/logic/base"
	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"
	"component-manage/pkg/models/types/components"

	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	RedisSentinelServiceName = "proton-redis-sentinel"
	RedisSentinelServicePort = 26379
	RedisMasterGroupName     = "mymaster"
)

func RedisValues(p *types.RedisComponentParams) helm3.M {
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
			"enableDualStack": global.Config.Config.EnableDualStack,
			"sentinel":        helm3.M{"port": 26379},
		},
		"enableSecurityContext": false,
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
		"redis": helm3.M{
			"masterGroupName": "mymaster",
			"rootUsername":    p.Admin_user,
			"rootPassword":    base64.StdEncoding.EncodeToString([]byte(p.Admin_passwd)),
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

func prepareRedis(param *types.RedisComponentParams) error {
	// 存储类无需准备
	if param.StorageClassName != "" {
		return nil
	}

	return base.PrepareStorage(param.Hosts, param.Data_path)
}

func clearRedis(param *types.RedisComponentParams) error {
	// 存储类无需准备
	if param.StorageClassName != "" {
		return nil
	}
	return base.ClearStorage(param.Hosts, param.Data_path)
}

func generateInfo(name string, p *types.RedisComponentParams) *components.RedisComponentInfo {
	return &types.RedisComponentInfo{
		SourceType:  "internal",
		ConnectType: "sentinel",

		SentinelHosts:   fmt.Sprintf("%s-%s.%s.%s", name, RedisSentinelServiceName, p.Namespace, global.Config.ServiceSuffix()),
		SentinelPort:    RedisSentinelServicePort,
		MasterGroupName: RedisMasterGroupName,

		Username:         p.Admin_user,
		Password:         p.Admin_passwd,
		SentinelUsername: p.Admin_user,
		SentinelPassword: p.Admin_passwd,
	}
}

func checkRedis(p, oldp *types.RedisComponentParams) error {
	p.Namespace = base.DefaultString(p.Namespace, "resource")
	if p.StorageClassName != "" && len(p.Hosts) > 0 {
		return errors.New(".storageClassName and .hosts cannot be set at the same time")
	}
	if p.StorageClassName != "" && p.Data_path != "" {
		return errors.New(".storageClassName and .data_path cannot be set at the same time")
	}
	// 由于组件管理服务没有（本地K8S集群）所有节点主机的信息，所以不进行“节点是否属于集群节点”类校验，此处只检查节点列表是否重复
	// 数据目录创建逻辑会检测数据目录路径是否有效，不在此处检验
	if len(p.Hosts) > 0 {
		set := sets.New[string](p.Hosts...)
		if set.Len() != 1 && set.Len() != 3 {
			return errors.New("redis only support 1 or 3 host")
		}
		if set.Len() < len(p.Hosts) {
			return errors.New("redis host list contains duplicated items")
		}
	}
	if oldp != nil {
		if base.DefaultString(oldp.Namespace, "resource") != p.Namespace {
			return errors.New("namespace is immutable")
		}
		if p.Data_path != oldp.Data_path {
			return errors.New("redis data path is immutable")
		}
		if p.StorageClassName != oldp.StorageClassName {
			return errors.New("redis storage class name is immutable")
		}
		if p.Admin_user != oldp.Admin_user {
			return errors.New("redis Admin_user is immutable")
		}
		if p.Admin_passwd != oldp.Admin_passwd {
			return errors.New("redis Admin_passwd is immutable")
		}
		if p.StorageCapacity != oldp.StorageCapacity {
			return errors.New("storageCapacity can not be changed")
		}
		// 仅支持扩容,不支持缩容
		replicaCountActual := func(p0 *components.RedisComponentParams) int {
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
				return fmt.Errorf("previous hosts must be in front of new hosts when expanding redis deployment: %v, %v", p.Hosts, diff)
			}
		}
	}
	return nil
}
