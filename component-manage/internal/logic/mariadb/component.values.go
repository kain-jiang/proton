package mariadb

import (
	"encoding/base64"
	"errors"
	"fmt"

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

func checkMariaDB(name string, param, oldParam *types.MariaDBComponentParams) error {
	// 计算 replicaCount，优先取 Hosts 长度，如果没有，再取 ReplicaCount
	param.ReplicaCount = func() int {
		replicas := len(param.Hosts)
		if replicas == 0 {
			return param.ReplicaCount
		}
		return replicas
	}()

	// 计算 AdminSecretName
	param.AdminSecretName = defaultTo(param.AdminSecretName, fmt.Sprintf("mariadb-admin-account-%s", name))
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

func mariadbManifest(name string, param *types.MariaDBComponentParams, pluginInfo *types.MariaDBPluginConfig) map[string]any {
	type m = map[string]any
	type l = []any

	var (
		imagePullPolicy = "IfNotPresent"

		// default values
		defaultSQLMode      = "STRICT_ALL_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION"
		defaultWsrepCtl     = "OFF"
		defaultLCTableNames = 1
		defaultCapacity     = "10Gi"
	)

	manifest := m{
		"apiVersion": "rds.proton.aishu.cn/v1",
		"kind":       "RDSMariaDBCluster",
		"metadata": m{
			"name":      name,
			"namespace": param.Namespace,
		},
		"spec": m{
			"etcd": m{
				"image":           pluginInfo.Images.ETCD,
				"imagePullPolicy": imagePullPolicy,
			},
			"exporter": m{
				"image":           pluginInfo.Images.Exporter,
				"imagePullPolicy": imagePullPolicy,
			},
			"mariadb": m{
				"conf": m{
					"innodb_buffer_pool_size":      "8G",
					"lower_case_table_names":       defaultLCTableNames,
					"sql_mode":                     defaultSQLMode,
					"wsrep_auto_increment_control": defaultWsrepCtl,
				},
				"image":           pluginInfo.Images.MariaDB,
				"imagePullPolicy": imagePullPolicy,
				"logrotate":       m{},
				"resources": m{
					"limits":   m{},
					"requests": m{},
				},
				"service": m{
					"enableDualStack": global.Config.Config.EnableDualStack,
					"port":            3330,
				},
				"storage": m{
					"capacity":         defaultTo(param.StorageCapacity, defaultCapacity),
					"storageClassName": param.StorageClassName,
					"volumeSpec": func() l {
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
			"mgmt": m{
				"conf":            m{},
				"image":           pluginInfo.Images.Mgmt,
				"imagePullPolicy": imagePullPolicy,
				"resources":       m{},
				"service": m{
					"enableDualStack": global.Config.Config.EnableDualStack,
					"port":            8888,
				},
			},
			"replicas":   param.ReplicaCount,
			"secretName": param.AdminSecretName,
		},
	}

	if param.Config != nil {
		if param.Config.Innodb_buffer_pool_size != "" {
			manifest["spec"].(m)["mariadb"].(m)["conf"].(m)["innodb_buffer_pool_size"] = param.Config.Innodb_buffer_pool_size
		}
		if param.Config.LowerCaseTableNames != nil {
			manifest["spec"].(m)["mariadb"].(m)["conf"].(m)["lower_case_table_names"] = *param.Config.LowerCaseTableNames
		}
		if param.Config.Thread_handling != "" {
			manifest["spec"].(m)["mariadb"].(m)["conf"].(m)["thread_handling"] = param.Config.Thread_handling
		}
		if param.Config.Resource_limits_memory != "" {
			manifest["spec"].(m)["mariadb"].(m)["resources"].(m)["limits"].(m)["memory"] = param.Config.Resource_limits_memory
		}
		if param.Config.Resource_requests_memory != "" {
			manifest["spec"].(m)["mariadb"].(m)["resources"].(m)["requests"].(m)["memory"] = param.Config.Resource_requests_memory
		}
	}

	return manifest
}

func prepareMariaDB(param *types.MariaDBComponentParams) error {
	// 创建 账户
	err := global.K8sCli.SecretSet(param.AdminSecretName, param.Namespace, map[string][]byte{
		"username": []byte(param.Admin_user),
		"password": []byte(param.Admin_passwd),
	})
	if err != nil {
		return fmt.Errorf("create admin secret failed: %w", err)
	}
	// 存储类时不需要准备存储目录
	if param.StorageClassName != "" {
		return nil
	}
	return base.PrepareStorage(param.Hosts, param.Data_path)
}

func clearMariaDB(param *types.MariaDBComponentParams) error {
	if param.StorageClassName != "" {
		return nil
	}
	return base.ClearStorage(param.Hosts, param.Data_path)
}

func generateInfo(name string, param *types.MariaDBComponentParams) *types.MariaDBComponentInfo {
	return &types.MariaDBComponentInfo{
		SourceType: "internal",
		RdsType:    "MariaDB",
		Hosts:      fmt.Sprintf("%s-mariadb-master.%s.%s", name, param.Namespace, global.Config.ServiceSuffix()),
		Port:       3330,
		Username:   param.Username,
		Password:   param.Password,
		HostsRead:  fmt.Sprintf("%s-mariadb-cluster.%s.%s", name, param.Namespace, global.Config.ServiceSuffix()),
		PortRead:   3330,
		// MgmtInfo
		MgmtHost: fmt.Sprintf("%s-mgmt-cluster.%s.%s", name, param.Namespace, global.Config.ServiceSuffix()),
		MgmtPort: 8888,
		AdminKey: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", param.Admin_user, param.Admin_passwd))),
	}
}
