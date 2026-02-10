package mongodb

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"component-manage/internal/global"
	"component-manage/internal/logic/base"
	"component-manage/pkg/models/types"
)

// tools
func toList(l []string) []string {
	if l == nil {
		return make([]string, 0)
	}
	return l
}

func defaultTo(val, dft string) string {
	if val == "" {
		return dft
	}
	return val
}

func checkMongoDB(name string, param, oldParam *types.MongoDBComponentParams) error {
	// 计算 replicaCount，优先取 Hosts 长度，如果没有，再取 ReplicaCount
	param.ReplicaCount = func() int {
		replicas := len(param.Hosts)
		if replicas == 0 {
			return param.ReplicaCount
		}
		return replicas
	}()

	param.AdminSecretName = defaultTo(param.AdminSecretName, fmt.Sprintf("mongodb-admin-account-%s", name))
	param.Namespace = defaultTo(param.Namespace, "resource")

	if oldParam != nil {
		// admin secret name 强制使用旧值
		param.AdminSecretName = oldParam.AdminSecretName
		param.Namespace = oldParam.Namespace

		// 更新
		if param.StorageClassName != oldParam.StorageClassName {
			return errors.New("storageClassName can not be changed")
		}

		if param.StorageCapacity != oldParam.StorageCapacity {
			return errors.New("storageCapacity can not be changed")
		}

		if param.Data_path != oldParam.Data_path {
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

		if param.Admin_user != oldParam.Admin_user {
			return errors.New("admin username can not be changed")
		}

		if param.Admin_passwd != oldParam.Admin_passwd {
			return errors.New("admin password can not be changed")
		}

	}

	return nil
}

func mongoDBManifest(name string, param *types.MongoDBComponentParams, pluginInfo *types.MongoDBPluginConfig) map[string]any {
	type m = map[string]any
	type l = []any

	imagePullPolicy := "IfNotPresent"

	manifest := m{
		"apiVersion": "mongodb.proton.aishu.cn/v1",
		"kind":       "MongodbOperator",
		"metadata": m{
			"name":      name,
			"namespace": param.Namespace,
		},
		"spec": m{
			"exporter": m{
				"image":           pluginInfo.Images.Exporter,
				"imagePullPolicy": imagePullPolicy,
			},
			"logrotate": m{
				"image":           pluginInfo.Images.Logrotate,
				"imagePullPolicy": imagePullPolicy,
				"logcount":        5,
				"logsize":         "500M",
				"schedule":        "*/30 * * * *",
			},
			"mgmt": m{
				"image":           pluginInfo.Images.Mgmt,
				"imagePullPolicy": imagePullPolicy,
				"logLevel":        "info",
				"service": m{
					"enableDualStack": global.Config.Config.EnableDualStack,
					"port":            30281,
					"type":            "ClusterIP",
				},
			},
			"mongodb": m{
				"conf": m{
					"tls": m{
						"enabled": false,
					},
					"wiredTigerCacheSizeGB": 4,
				},
				"debug":           "0",
				"image":           pluginInfo.Images.MongoDB,
				"imagePullPolicy": imagePullPolicy,
				"replicas":        param.ReplicaCount,
				"replset": m{
					"name": "rs0",
				},
				"resources": m{
					"requests": m{
						"cpu":    "1",
						"memory": "1Gi",
					},
				},
				"service": m{
					"enableDualStack": global.Config.Config.EnableDualStack,
					"port":            30280,
					"type":            "ClusterIP",
				},
				"storage": m{
					"capacity":         defaultTo(param.StorageCapacity, "10Gi"),
					"storageClassName": param.StorageClassName,
					"volume": func() l {
						rel := make(l, 0, len(param.Hosts))
						for _, host := range param.Hosts {
							rel = append(rel, m{
								"host": host,
								"path": param.Data_path,
							})
						}
						return rel
					}(),
				},
			},
			"secretname": param.AdminSecretName,
		},
	}

	if param.Resources != nil {
		manifest["spec"].(m)["mongodb"].(m)["resources"] = param.Resources.ResourceMap()
	}

	return manifest
}

func prepareMongoDB(param *types.MongoDBComponentParams) error {
	// 创建 账户
	err := global.K8sCli.SecretSet(param.AdminSecretName, param.Namespace, map[string][]byte{
		"username": []byte(param.Admin_user),
		// 密码双重 base64
		"password": []byte(base64.StdEncoding.EncodeToString([]byte(param.Admin_passwd))),
	})
	if err != nil {
		return fmt.Errorf("create admin secret failed: %w", err)
	}

	if param.StorageClassName != "" {
		return nil
	}
	return base.PrepareStorage(param.Hosts, param.Data_path)
}

func clearMongoDB(param *types.MongoDBComponentParams) error {
	if param.StorageClassName != "" {
		return nil
	}
	return base.ClearStorage(param.Hosts, param.Data_path)
}

func generateInfo(name string, param *types.MongoDBComponentParams) *types.MongoDBComponentInfo {
	hosts := make([]string, 0, param.ReplicaCount)
	for i := range param.ReplicaCount {
		hosts = append(hosts, fmt.Sprintf("%s-mongodb-%d.%s-mongodb.%s.%s", name, i, name, param.Namespace, global.Config.ServiceSuffix()))
	}

	return &types.MongoDBComponentInfo{
		SourceType: "internal",
		Hosts:      strings.Join(hosts, ","),
		Port:       28000,
		ReplicaSet: "rs0",
		Username:   param.Username,
		Password:   param.Password,
		SSL:        false,
		AuthSource: "anyshare",
		Options:    nil,
		// MgmtInfo
		MgmtHost: fmt.Sprintf("%s-mgmt-cluster.%s.%s", name, param.Namespace, global.Config.ServiceSuffix()),
		MgmtPort: 30281,
		AdminKey: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", param.Admin_user, param.Admin_passwd))),
	}
}
