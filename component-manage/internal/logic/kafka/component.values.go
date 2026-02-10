package kafka

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

//

func kafkaValues(param *types.KafkaComponentParams, zk *components.ComponentZookeeper) helm3.M {
	values := helm3.M{
		"config": helm3.M{
			"kafkaENV": util.AnyMapToMapAny(param.Env),
			"sasl": helm3.M{
				"enabled":   true,
				"mechanism": "PLAIN",
				"user": helm3.L{
					helm3.M{
						"username": "FIXME",
						"password": "FIXME",
					},
				},
			},
		},
		"depServices": helm3.M{
			"zookeeper": helm3.M{
				"host": zk.Info.Host,
				"port": zk.Info.Port,
				"sasl": helm3.M{
					"enabled":  zk.Info.Sasl.Enabled,
					"username": zk.Info.Sasl.Username,
					"password": zk.Info.Sasl.Password,
				},
				"ssl": helm3.M{"enabled": false},
			},
		},
		"image": helm3.M{
			"registry": global.Config.Config.Registry,
		},
		"namespace":    param.Namespace,
		"replicaCount": param.ReplicaCount,
		"resources": helm3.M{
			"kafka": param.Resources.ResourceMap(),
		},
		"service": helm3.M{
			"enableDualStack": global.Config.Config.EnableDualStack,
			"external": func() helm3.M {
				rel := helm3.M{
					"enabled": !param.DisableExternalService,
				}
				if ports := param.ExternalServiceList; ports != nil {
					rel["ports"] = ports
				}
				return rel
			}(),
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

	if param.ExporterResources != nil {
		values["resources"].(helm3.M)["exporter"] = param.ExporterResources.ResourceMap()
	}

	return values
}

func prepareKafka(param *types.KafkaComponentParams) error {
	// 存储类无需准备
	if param.StorageClassName != "" {
		return nil
	}

	return base.PrepareStorage(param.Hosts, param.DataPath)
}

func clearKafka(param *types.KafkaComponentParams) error {
	// 存储类无需准备
	if param.StorageClassName != "" {
		return nil
	}
	return base.ClearStorage(param.Hosts, param.DataPath)
}

func generateInfo(name string, params *types.KafkaComponentParams) *components.KafkaComponentInfo {
	return &types.KafkaComponentInfo{
		MQHosts:    fmt.Sprintf("%s-headless.%s.%s", base.TemplateName(name, "kafka"), params.Namespace, global.Config.ServiceSuffix()),
		MQPort:     9097,
		MQType:     "kafka",
		SourceType: "internal",
		Auth: types.KafkaAuth{
			Username:  "FIXME",
			Password:  "FIXME",
			Mechanism: "PLAIN",
		},
	}
}

func checkKafka(param, oldParam *types.KafkaComponentParams) error {
	param.Namespace = base.DefaultString(param.Namespace, "resource")
	// 计算 replicaCount，优先取 Hosts 长度，如果没有，再取 ReplicaCount
	param.ReplicaCount = func() int {
		replicas := len(param.Hosts)
		if replicas == 0 {
			return param.ReplicaCount
		}
		return replicas
	}()

	// 如果关闭ExternalService，那么ExternalService列表强制设置为空
	// if param.DisableExternalService {
	// 	param.ExternalServiceList = make([]map[string]any, 0)
	// }

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
