package zookeeper

import (
	"errors"
	"fmt"
	"strconv"

	"component-manage/internal/logic/base"
	"component-manage/pkg/models/types/components"

	"component-manage/internal/global"
	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"
	"component-manage/pkg/util"
)

// tools
func toList(l []string) []string {
	if l == nil {
		return make([]string, 0)
	}
	return l
}

func zookeeperValues(param *types.ZookeeperComponentParams) helm3.M {
	values := helm3.M{
		"config": helm3.M{
			"zookeeperENV": util.AnyMapToMapAny(param.Env),
		},
		"image": helm3.M{
			"registry": global.Config.Config.Registry,
		},
		"namespace":    param.Namespace,
		"replicaCount": param.ReplicaCount,
		"resources":    param.Resources.ResourceMap(),
		"service": helm3.M{
			"enableDualStack": global.Config.Config.EnableDualStack,
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
	}

	if param.StorageCapacity != "" {
		values["storage"].(helm3.M)["capacity"] = param.StorageCapacity
	}

	return values
}

func prepareZookeeper(param *types.ZookeeperComponentParams) error {
	// 存储类无需准备
	if param.StorageClassName != "" {
		return nil
	}
	return base.PrepareStorage(param.Hosts, param.DataPath)
}

func clearZookeeper(param *types.ZookeeperComponentParams) error {
	// 存储类无需清理
	if param.StorageClassName != "" {
		return nil
	}
	return base.ClearStorage(param.Hosts, param.DataPath)
}

func generateInfo(name string, _ *types.ZookeeperComponentParams) *components.ZookeeperComponentInfo {
	return &types.ZookeeperComponentInfo{
		Host: fmt.Sprintf("%s-headless", base.TemplateName(name, "zookeeper")),
		Port: 2181,
		Sasl: types.ZookeeperSASL{
			Enabled:  true,
			Username: "kafka",
			Password: "FIXME",
		},
	}
}

func checkZookeeper(param, oldParam *types.ZookeeperComponentParams) error {
	param.Namespace = base.DefaultString(param.Namespace, "resource")
	// 计算 replicaCount，优先取 Hosts 长度，如果没有，再取 ReplicaCount
	param.ReplicaCount = func() int {
		replicas := len(param.Hosts)
		if replicas == 0 {
			return param.ReplicaCount
		}
		return replicas
	}()

	if oldParam != nil {
		// 更新
		if base.DefaultString(oldParam.Namespace, "resource") != param.Namespace {
			return errors.New("namespace is immutable")
		}

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
